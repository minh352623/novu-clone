package tracking

import (
	"strings"
	"testing"
)

func TestGetTransparentPixel(t *testing.T) {
	pixel := GetTransparentPixel()

	if len(pixel) == 0 {
		t.Error("GetTransparentPixel() returned empty data")
	}

	// PNG signature: 137 80 78 71 13 10 26 10
	pngSignature := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	if len(pixel) < 8 {
		t.Fatal("Pixel data too short to be a valid PNG")
	}

	for i, b := range pngSignature {
		if pixel[i] != b {
			t.Errorf("PNG signature mismatch at byte %d: got %x, want %x", i, pixel[i], b)
		}
	}
}

func TestGetPixelContentType(t *testing.T) {
	contentType := GetPixelContentType()

	if contentType != "image/png" {
		t.Errorf("GetPixelContentType() = %v, want image/png", contentType)
	}
}

func TestGenerateTrackingPixelHTML(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		token    string
		expected []string // strings that should be present in result
	}{
		{
			name:    "Basic URL and token",
			baseURL: "https://example.com",
			token:   "abc123",
			expected: []string{
				`src="https://example.com/v1/api/track/abc123.png"`,
				`width="1"`,
				`height="1"`,
				`style="display:none`,
			},
		},
		{
			name:    "URL with trailing slash",
			baseURL: "https://example.com/",
			token:   "xyz789",
			expected: []string{
				`src="https://example.com/v1/api/track/xyz789.png"`,
			},
		},
		{
			name:    "Long token",
			baseURL: "https://api.example.com",
			token:   "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
			expected: []string{
				`src="https://api.example.com/v1/api/track/a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6.png"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateTrackingPixelHTML(tt.baseURL, tt.token)

			for _, exp := range tt.expected {
				if !strings.Contains(result, exp) {
					t.Errorf("Result should contain %q\nGot: %s", exp, result)
				}
			}

			// Should be an img tag
			if !strings.HasPrefix(result, "<img ") {
				t.Errorf("Result should start with <img, got: %s", result)
			}

			if !strings.HasSuffix(result, "/>") {
				t.Errorf("Result should end with />, got: %s", result)
			}
		})
	}
}

func TestInjectTrackingPixel(t *testing.T) {
	tests := []struct {
		name        string
		htmlContent string
		baseURL     string
		token       string
		verify      func(t *testing.T, result string)
	}{
		{
			name:        "Inject before </body>",
			htmlContent: "<html><body><p>Hello</p></body></html>",
			baseURL:     "https://example.com",
			token:       "test123",
			verify: func(t *testing.T, result string) {
				if !strings.Contains(result, `<p>Hello</p><img src=`) {
					t.Error("Pixel should be injected before </body>")
				}
				if !strings.Contains(result, "</body></html>") {
					t.Error("Original tags should be preserved")
				}
			},
		},
		{
			name:        "Inject before </BODY> (uppercase)",
			htmlContent: "<HTML><BODY><P>Hello</P></BODY></HTML>",
			baseURL:     "https://example.com",
			token:       "test123",
			verify: func(t *testing.T, result string) {
				if !strings.Contains(result, `<P>Hello</P><img src=`) {
					t.Error("Pixel should be injected before </BODY>")
				}
			},
		},
		{
			name:        "Inject before </html> when no body tag",
			htmlContent: "<html><p>Hello</p></html>",
			baseURL:     "https://example.com",
			token:       "test123",
			verify: func(t *testing.T, result string) {
				if !strings.Contains(result, `<p>Hello</p><img src=`) {
					t.Error("Pixel should be injected before </html>")
				}
			},
		},
		{
			name:        "Append to end when no closing tags",
			htmlContent: "<p>Hello World</p>",
			baseURL:     "https://example.com",
			token:       "test123",
			verify: func(t *testing.T, result string) {
				if !strings.HasSuffix(result, "/>") {
					t.Error("Pixel should be appended at end")
				}
				if !strings.Contains(result, "test123") {
					t.Error("Token should be present in result")
				}
			},
		},
		{
			name:        "Empty HTML content",
			htmlContent: "",
			baseURL:     "https://example.com",
			token:       "test123",
			verify: func(t *testing.T, result string) {
				if !strings.Contains(result, "test123") {
					t.Error("Pixel should be added even to empty content")
				}
			},
		},
		{
			name:        "URL with trailing slash is handled",
			htmlContent: "<body></body>",
			baseURL:     "https://example.com/",
			token:       "abc",
			verify: func(t *testing.T, result string) {
				if strings.Contains(result, "example.com//") {
					t.Error("Double slash should not appear in URL")
				}
				if !strings.Contains(result, "example.com/v1/api/track") {
					t.Error("URL should be properly formed")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InjectTrackingPixel(tt.htmlContent, tt.baseURL, tt.token)
			tt.verify(t, result)

			// Common checks for all tests
			if !strings.Contains(result, tt.token) {
				t.Errorf("Result should contain token %q", tt.token)
			}

			if !strings.Contains(result, "img src=") {
				t.Error("Result should contain img tag")
			}
		})
	}
}

func TestInjectTrackingPixel_PreservesContent(t *testing.T) {
	original := `<!DOCTYPE html>
<html>
<head>
	<title>Test Email</title>
</head>
<body>
	<h1>Welcome!</h1>
	<p>This is a test email with <strong>formatting</strong>.</p>
	<a href="https://example.com">Click here</a>
</body>
</html>`

	result := InjectTrackingPixel(original, "https://track.example.com", "token123")

	// All original content should be preserved
	preservedStrings := []string{
		"<!DOCTYPE html>",
		"<title>Test Email</title>",
		"<h1>Welcome!</h1>",
		"<strong>formatting</strong>",
		`href="https://example.com"`,
		"</html>",
	}

	for _, s := range preservedStrings {
		if !strings.Contains(result, s) {
			t.Errorf("Original content %q was not preserved", s)
		}
	}

	// Tracking pixel should be present
	if !strings.Contains(result, "token123") {
		t.Error("Tracking token should be in result")
	}
}
