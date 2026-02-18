package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex" // or base64
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"CONVERDA/global"

	"CONVERDA/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const SHARED_SECRET_KEY = "your-very-secret-and-long-key" // !!! in env
const requestValidityDuration = 200 * time.Second

// AuthGuardMiddlewareWithHMAC by HMAC
func AuthGuardMiddlewareWithHMAC() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		clientSign := ctx.GetHeader("X-Sign")
		requestTimeStr := ctx.GetHeader("X-Request-Time")

		if clientSign == "" || requestTimeStr == "" {
			global.Logger.Warn("HMAC Auth: missing X-Sign or X-Request-Time header")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response.NewAPIError(http.StatusUnauthorized, "Unauthorized", "Missing required signature headers"))
			return
		}

		// 1. Kiểm tra Timestamp
		requestTime, err := strconv.ParseInt(requestTimeStr, 10, 64)
		if err != nil {
			global.Logger.Warn("HMAC Auth: invalid request time format")
			ctx.AbortWithStatusJSON(http.StatusBadRequest, response.NewAPIError(http.StatusBadRequest, "Invalid request", "Invalid request time format"))
			return
		}

		now := time.Now().Unix()
		if now-requestTime > int64(requestValidityDuration.Seconds()) || now-requestTime < -5 { // Chấp nhận trễ 5s
			global.Logger.Warn("HMAC Auth: request timestamp out of bounds")
			ctx.AbortWithStatusJSON(http.StatusBadRequest, response.NewAPIError(http.StatusBadRequest, "Invalid request", "Request timestamp out of bounds"))
			return
		}

		// 2. Tái tạo String-to-Sign
		// Đọc body (nếu có) và chuẩn bị để đọc lại
		var bodyBytes []byte
		if ctx.Request.Body != nil {
			bodyBytes, err = io.ReadAll(ctx.Request.Body)
			if err != nil {
				global.Logger.Error("HMAC Auth: error reading request body", zap.Error(err))
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, response.NewAPIError(http.StatusInternalServerError, "Server Error", "Could not read request body"))
				return
			}
			// Tạo lại body để handler sau có thể đọc
			ctx.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		stringToSign := buildStringToSign(ctx, requestTimeStr, bodyBytes)
		global.Logger.Info("HMAC Auth: server StringToSign", zap.String("content", strings.ReplaceAll(stringToSign, "\n", "\\n")))

		// 3. Tính toán HMAC phía server
		serverSign := calculateHMAC(stringToSign, SHARED_SECRET_KEY)
		global.Logger.Info("HMAC Auth: verification", zap.String("clientSign", clientSign), zap.String("serverSign", serverSign))

		// 4. So sánh chữ ký
		// Sử dụng hmac.Equal để so sánh an toàn, chống timing attacks
		if !hmac.Equal([]byte(clientSign), []byte(serverSign)) {
			global.Logger.Warn("HMAC Auth: invalid signature")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response.NewAPIError(http.StatusUnauthorized, "Unauthorized", "Invalid signature"))
			return
		}

		global.Logger.Info("HMAC Auth: signature verified successfully")
		ctx.Next()
	}
}

func buildStringToSign(ctx *gin.Context, requestTimeStr string, bodyBytes []byte) string {
	method := ctx.Request.Method
	path := ctx.Request.URL.Path // Chỉ path, không có query string

	// Sắp xếp query parameters
	queryParams := ctx.Request.URL.Query()
	var sortedQueryKeys []string
	for k := range queryParams {
		sortedQueryKeys = append(sortedQueryKeys, k)
	}
	sort.Strings(sortedQueryKeys)

	var canonicalQueryParts []string
	for _, k := range sortedQueryKeys {
		// Gin's queryParams[k] is a []string, handle multiple values if necessary
		// For simplicity, we'll take the first one or join them if you expect multiple
		// For this example, let's assume single values for simplicity or take first
		if len(queryParams[k]) > 0 {
			canonicalQueryParts = append(canonicalQueryParts, fmt.Sprintf("%s=%s", k, queryParams[k][0]))
		}
	}
	canonicalQueryString := strings.Join(canonicalQueryParts, "&")

	// HASHING BODY IS BEST PRACTICE.
	bodyHash := ""
	if len(bodyBytes) > 0 {
		bodyHasher := sha256.New()
		bodyHasher.Write(bodyBytes)
		bodyHash = hex.EncodeToString(bodyHasher.Sum(nil))
	}

	// Thứ tự phải nhất quán giữa client và server
	// Ví dụ: METHOD\nPATH\nTIMESTAMP\nSORTED_QUERY_STRING\nBODY_HASH
	parts := []string{
		method,
		path,
		requestTimeStr,
		canonicalQueryString,
		bodyHash,
	}
	return strings.Join(parts, "\n")
}

func calculateHMAC(data string, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil)) // Hoặc base64.StdEncoding.EncodeToString
}
