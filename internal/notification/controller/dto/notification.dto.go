package dto

type SendNotificationRequest struct {
	TemplateCode string                 `json:"template_code" binding:"required"`
	Recipient    string                 `json:"recipient" binding:"required,email"` // Validate email format
	Channel      string                 `json:"channel" binding:"required,oneof=email sms push"`
	Data         map[string]interface{} `json:"data"`
	Language     string                 `json:"language"` // Optional, default 'en'
}

type NotificationResponse struct {
	NotificationID string `json:"notification_id"`
	Status         string `json:"status"`
}
