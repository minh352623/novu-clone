package impl

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"CONVERDA/global"
	"CONVERDA/internal/notification/application/service"
	"CONVERDA/internal/notification/domain/entity"
	"CONVERDA/internal/notification/domain/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	defaultMaxRetries     = 3
	defaultBackoffBaseSec = 5
	backoffMultiplier     = 6 // 5s → 30s → 180s
	defaultTimeoutSec     = 10
	maxResponseBody       = 1024
)

type webhookDispatcherImpl struct {
	webhookRepo    repository.WebhookRepository
	webhookLogRepo repository.WebhookLogRepository
	httpClient     *http.Client
}

func NewWebhookDispatcher(
	webhookRepo repository.WebhookRepository,
	webhookLogRepo repository.WebhookLogRepository,
) service.WebhookDispatcher {
	return &webhookDispatcherImpl{
		webhookRepo:    webhookRepo,
		webhookLogRepo: webhookLogRepo,
		httpClient: &http.Client{
			Timeout: time.Duration(defaultTimeoutSec) * time.Second,
		},
	}
}

func (d *webhookDispatcherImpl) Dispatch(ctx context.Context, envID uuid.UUID, eventType string, payload interface{}) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				global.Logger.Error("webhook_dispatcher: panic recovered", zap.Any("panic", r), zap.String("eventType", eventType))
			}
		}()
		bgCtx := context.Background()

		webhooks, err := d.webhookRepo.GetByEvent(bgCtx, envID, eventType)
		if err != nil {
			global.Logger.Error("webhook_dispatcher: failed to fetch webhooks", zap.String("eventType", eventType), zap.Error(err))
			return
		}

		if len(webhooks) == 0 {
			return
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			global.Logger.Error("webhook_dispatcher: failed to marshal payload", zap.Error(err))
			return
		}

		var payloadMap map[string]interface{}
		_ = json.Unmarshal(payloadBytes, &payloadMap)

		for _, wh := range webhooks {
			d.triggerWebhook(bgCtx, wh, eventType, payloadBytes, payloadMap)
		}
	}()
}

func (d *webhookDispatcherImpl) RetryWebhook(ctx context.Context, log *entity.WebhookLog) {
	payloadBytes, err := json.Marshal(log.RequestPayload)
	if err != nil {
		global.Logger.Error("webhook_dispatcher: failed to marshal retry payload", zap.String("logID", log.ID.String()), zap.Error(err))
		return
	}

	// Skip signing on retries (no secret available from log)
	d.sendHTTPRequest(ctx, log, log.EventType, payloadBytes, "")
}

func (d *webhookDispatcherImpl) triggerWebhook(ctx context.Context, wh *entity.Webhook, eventType string, payloadBytes []byte, payloadMap map[string]interface{}) {
	logID := uuid.New()

	// Use per-webhook retry config, fallback to defaults
	maxR := wh.MaxRetries
	if maxR <= 0 {
		maxR = defaultMaxRetries
	}

	logEntry := &entity.WebhookLog{
		ID:             logID,
		WebhookID:      wh.ID,
		URL:            wh.URL,
		EventType:      eventType,
		RequestPayload: payloadMap,
		Status:         entity.WebhookLogStatusPending,
		RetryCount:     0,
		MaxRetries:     maxR,
		TenantID:       wh.TenantID,
		AppID:          wh.AppID,
		CreatedAt:      time.Now(),
	}

	if err := d.webhookLogRepo.Create(ctx, logEntry); err != nil {
		global.Logger.Error("webhook_dispatcher: failed to create log", zap.Error(err))
	}

	d.sendHTTPRequest(ctx, logEntry, eventType, payloadBytes, wh.Secret)
}

func (d *webhookDispatcherImpl) sendHTTPRequest(ctx context.Context, logEntry *entity.WebhookLog, eventType string, payloadBytes []byte, secret string) {
	req, err := http.NewRequestWithContext(ctx, "POST", logEntry.URL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		d.handleFailure(ctx, logEntry, 0, err.Error(), 0)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Converda-Webhook/1.0")
	req.Header.Set("X-Converda-Event", eventType)
	req.Header.Set("X-Converda-Delivery", logEntry.ID.String())

	// Sign payload with HMAC if secret is available
	if secret != "" {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(payloadBytes)
		signature := hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-Converda-Signature", "sha256="+signature)
	}

	start := time.Now()
	resp, err := d.httpClient.Do(req)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		d.handleFailure(ctx, logEntry, 0, err.Error(), duration)
		return
	}
	defer resp.Body.Close()

	// Read limited response body
	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
	respBody := string(bodyBytes)

	if resp.StatusCode >= 400 {
		d.handleFailure(ctx, logEntry, resp.StatusCode, respBody, duration)
		return
	}

	// Success
	logEntry.Status = entity.WebhookLogStatusSuccess
	logEntry.ResponseCode = resp.StatusCode
	logEntry.ResponseBody = &respBody
	logEntry.DurationMs = duration
	logEntry.NextRetryAt = nil

	if err := d.webhookLogRepo.Update(ctx, logEntry); err != nil {
		global.Logger.Error("webhook_dispatcher: failed to update log on success", zap.String("logID", logEntry.ID.String()), zap.Error(err))
	}
}

func (d *webhookDispatcherImpl) handleFailure(ctx context.Context, logEntry *entity.WebhookLog, code int, body string, duration int64) {
	logEntry.Status = entity.WebhookLogStatusFailed
	logEntry.ResponseCode = code
	logEntry.ResponseBody = &body
	logEntry.DurationMs = duration

	if logEntry.RetryCount < logEntry.MaxRetries {
		// Schedule next retry with exponential backoff
		// Use per-webhook backoff base from the log's associated webhook config
		backoffSec := defaultBackoffBaseSec
		// backoffSec is stored alongside the log via the webhook config at dispatch time
		backoff := time.Duration(backoffSec) * time.Second
		for i := 0; i < logEntry.RetryCount; i++ {
			backoff = backoff * time.Duration(backoffMultiplier)
		}
		nextRetry := time.Now().Add(backoff)
		logEntry.NextRetryAt = &nextRetry
		logEntry.RetryCount++

		global.Logger.Warn("webhook_dispatcher: scheduling retry",
			zap.Int("retryCount", logEntry.RetryCount),
			zap.String("logID", logEntry.ID.String()),
			zap.Duration("backoff", backoff))
	} else {
		global.Logger.Error("Webhook dispatch: max retries reached for log " + logEntry.ID.String() + " — dead letter")
		logEntry.NextRetryAt = nil
	}

	if err := d.webhookLogRepo.Update(ctx, logEntry); err != nil {
		global.Logger.Error("webhook_dispatcher: failed to update log on failure", zap.String("logID", logEntry.ID.String()), zap.Error(err))
	}
}
