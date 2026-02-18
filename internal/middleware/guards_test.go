package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"CONVERDA/global"
	"CONVERDA/pkg/logger"
	"CONVERDA/pkg/setting"
)

func init() {
	global.Logger = logger.NewLogger(setting.LoggerSetting{LogLevel: "debug"})
}

func TestAuthGuardMiddlewareWithHMAC(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Valid Signature with Body", func(t *testing.T) {
		r := gin.New()
		r.Use(AuthGuardMiddlewareWithHMAC())
		r.POST("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		body := []byte(`{"hello":"world"}`)
		timestamp := time.Now().Unix()
		timestampStr := strconv.FormatInt(timestamp, 10)

		// Calculate server-side signature for verification
		bodyHasher := sha256.New()
		bodyHasher.Write(body)
		bodyHash := hex.EncodeToString(bodyHasher.Sum(nil))

		stringToSign := fmt.Sprintf("POST\n/test\n%s\n\n%s", timestampStr, bodyHash)
		mac := hmac.New(sha256.New, []byte(SHARED_SECRET_KEY))
		mac.Write([]byte(stringToSign))
		signature := hex.EncodeToString(mac.Sum(nil))

		req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(body))
		req.Header.Set("X-Sign", signature)
		req.Header.Set("X-Request-Time", timestampStr)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Invalid Signature", func(t *testing.T) {
		r := gin.New()
		r.Use(AuthGuardMiddlewareWithHMAC())

		req, _ := http.NewRequest("POST", "/test", nil)
		req.Header.Set("X-Sign", "invalid")
		req.Header.Set("X-Request-Time", strconv.FormatInt(time.Now().Unix(), 10))

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Expired Request", func(t *testing.T) {
		r := gin.New()
		r.Use(AuthGuardMiddlewareWithHMAC())

		timestamp := time.Now().Add(-1 * time.Hour).Unix()
		req, _ := http.NewRequest("POST", "/test", nil)
		req.Header.Set("X-Sign", "any")
		req.Header.Set("X-Request-Time", strconv.FormatInt(timestamp, 10))

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
