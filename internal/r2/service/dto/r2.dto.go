package dto

// ImageDTO represents an image to upload
type ImageDTO struct {
	Key    string `json:"key" binding:"required"`
	Base64 string `json:"base64" binding:"required"`
}

// UploadRequest represents the upload request with multiple images
type UploadRequest struct {
	Images []ImageDTO `json:"images" binding:"required,dive"`
}

// UploadResult represents the result of a single upload
type UploadResult struct {
	Key     string `json:"key"`
	URL     string `json:"url"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// UploadResponse represents the response for batch upload
type UploadResponse struct {
	Results      []UploadResult `json:"results"`
	TotalCount   int            `json:"totalCount"`
	SuccessCount int            `json:"successCount"`
	FailedCount  int            `json:"failedCount"`
}
