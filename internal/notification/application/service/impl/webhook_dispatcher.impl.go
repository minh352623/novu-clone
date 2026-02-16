package impl

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"CONVERDA/internal/notification/application/service"
	"CONVERDA/internal/notification/domain/entity"
	"CONVERDA/internal/notification/domain/repository"

	"github.com/google/uuid"
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
			Timeout: 10 * time.Second,
		},
	}
}

func (d *webhookDispatcherImpl) Dispatch(ctx context.Context, envID uuid.UUID, eventType string, payload interface{}) {
	// 1. Find interested webhooks
	// Use a detached context for background processing if needed, but for now we run in a goroutine
	// so the passed context might be cancelled by the parent request.
	// Best practice: Create a new context for the background worker or use the passed one if we want to wait (plan said async).
	// Implementation plan says: "Execute asynchronously (goroutine)".
	go func() {
		// Create a background context for the async operation
		bgCtx := context.Background()

		webhooks, err := d.webhookRepo.GetByEvent(bgCtx, envID, eventType)
		if err != nil {
			// Log error (system log not available yet, just print for now)
			fmt.Printf("Error fetching webhooks for event %s: %v\n", eventType, err)
			return
		}

		if len(webhooks) == 0 {
			return
		}

		// Prepare Payload
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			fmt.Printf("Error marshaling payload: %v\n", err)
			return
		}

		// Convert payload to map for logging
		var payloadMap map[string]interface{}
		_ = json.Unmarshal(payloadBytes, &payloadMap)

		for _, wh := range webhooks {
			d.triggerWebhook(bgCtx, wh, eventType, payloadBytes, payloadMap)
		}
	}()
}

func (d *webhookDispatcherImpl) triggerWebhook(ctx context.Context, wh *entity.Webhook, eventType string, payloadBytes []byte, payloadMap map[string]interface{}) {
	// 2. Create Audit Log (Pending)
	logID := uuid.New() // V4 for now or use V7 generator if available in utils
	// Since we don't have widely available v7 helper here, let's trust GORM default or use simple New()

	logEntry := &entity.WebhookLog{
		ID:             logID,
		WebhookID:      wh.ID,
		URL:            wh.URL,
		EventType:      eventType,
		RequestPayload: payloadMap,
		Status:         entity.WebhookLogStatusPending,
		TenantID:       wh.TenantID,
		AppID:          wh.AppID,
		CreatedAt:      time.Now(),
	}

	if err := d.webhookLogRepo.Create(ctx, logEntry); err != nil {
		fmt.Printf("Error creating webhook log: %v\n", err)
		// Continue even if logging fails? Debatable. For now, yes.
	}

	// 3. Prepare Request
	req, err := http.NewRequestWithContext(ctx, "POST", wh.URL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		d.updateLogStatus(ctx, logEntry, entity.WebhookLogStatusFailed, 0, err.Error(), 0)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Converda-Webhook/1.0")
	req.Header.Set("X-Converda-Event", eventType)
	req.Header.Set("X-Converda-Delivery", logID.String())

	// Sign payload if secret exists
	if wh.Secret != "" {
		mac := hmac.New(sha256.New, []byte(wh.Secret))
		mac.Write(payloadBytes)
		signature := hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-Converda-Signature", "sha256="+signature)
	}

	// 4. Send Request
	start := time.Now()
	resp, err := d.httpClient.Do(req)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		d.updateLogStatus(ctx, logEntry, entity.WebhookLogStatusFailed, 0, err.Error(), duration)
		return
	}
	defer resp.Body.Close()

	// Read Response (Limit size)
	// TODO: limited reader
	// For now, just simplistic status check
	status := entity.WebhookLogStatusSuccess
	if resp.StatusCode >= 400 {
		status = entity.WebhookLogStatusFailed
	}

	respBody := fmt.Sprintf("Status: %d", resp.StatusCode)

	d.updateLogStatus(ctx, logEntry, status, resp.StatusCode, respBody, duration)
}

func (d *webhookDispatcherImpl) updateLogStatus(ctx context.Context, logEntry *entity.WebhookLog, status string, code int, body string, duration int64) {
	logEntry.Status = status
	logEntry.ResponseCode = code
	logEntry.ResponseBody = &body
	logEntry.DurationMs = duration

	if err := d.webhookLogRepo.Update(ctx, logEntry); err != nil {
		fmt.Printf("Error updating webhook log: %v\n", err)
	}
}
