package tracking

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// 1x1 transparent PNG pixel (base64 encoded)
// This is a minimal valid PNG file that displays as a transparent 1x1 pixel
const transparentPixelBase64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="

var transparentPixelBytes []byte

func init() {
	var err error
	transparentPixelBytes, err = base64.StdEncoding.DecodeString(transparentPixelBase64)
	if err != nil {
		panic("failed to decode transparent pixel: " + err.Error())
	}
}

// GetTransparentPixel returns a 1x1 transparent PNG image as bytes
func GetTransparentPixel() []byte {
	return transparentPixelBytes
}

// GetPixelContentType returns the content type for PNG images
func GetPixelContentType() string {
	return "image/png"
}

// GenerateTrackingPixelHTML generates the HTML for a tracking pixel
func GenerateTrackingPixelHTML(baseURL, token string) string {
	// Ensure baseURL doesn't end with /
	baseURL = strings.TrimSuffix(baseURL, "/")

	return fmt.Sprintf(
		`<img src="%s/v1/api/track/%s.png" width="1" height="1" style="display:none;width:1px;height:1px;border:0;" alt="" />`,
		baseURL,
		token,
	)
}

// InjectTrackingPixel injects a tracking pixel into HTML content
// It inserts the pixel before the closing </body> tag if present,
// otherwise appends it to the end of the content
func InjectTrackingPixel(htmlContent, baseURL, token string) string {
	pixelHTML := GenerateTrackingPixelHTML(baseURL, token)

	// Try to insert before </body>
	if idx := strings.LastIndex(htmlContent, "</body>"); idx != -1 {
		return htmlContent[:idx] + pixelHTML + htmlContent[idx:]
	}

	if idx := strings.LastIndex(htmlContent, "</BODY>"); idx != -1 {
		return htmlContent[:idx] + pixelHTML + htmlContent[idx:]
	}

	// Try to insert before </html>
	if idx := strings.LastIndex(htmlContent, "</html>"); idx != -1 {
		return htmlContent[:idx] + pixelHTML + htmlContent[idx:]
	}

	if idx := strings.LastIndex(htmlContent, "</HTML>"); idx != -1 {
		return htmlContent[:idx] + pixelHTML + htmlContent[idx:]
	}

	// Fallback: append to end
	return htmlContent + pixelHTML
}
