package email

import (
	"bytes"
	"context"
	"fmt"
	"html/template"

	"CONVERDA/internal/iam/application/service"
	"CONVERDA/pkg/setting"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

type sesEmailService struct {
	client     *sesv2.Client
	from       string
	inviteBase string
	htmlTmpl   *template.Template
	textTmpl   *template.Template
}

// NewSESEmailService creates a new SES email service
func NewSESEmailService(sesSetting setting.SESSetting) (service.EmailService, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(sesSetting.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			sesSetting.AccessKeyId,
			sesSetting.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	client := sesv2.NewFromConfig(cfg)

	// Parse templates at init time (fail-fast)
	htmlTmpl, err := template.New("invitation_html").Parse(invitationHTMLTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML template: %w", err)
	}
	textTmpl, err := template.New("invitation_text").Parse(invitationTextTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse text template: %w", err)
	}

	inviteBase := sesSetting.InviteBaseURL
	if inviteBase == "" {
		inviteBase = "https://app.converda.com/invite"
	}

	return &sesEmailService{
		client:     client,
		from:       sesSetting.FromEmail,
		inviteBase: inviteBase,
		htmlTmpl:   htmlTmpl,
		textTmpl:   textTmpl,
	}, nil
}

// invitationData holds variables for the invitation email template.
type invitationData struct {
	TenantName  string
	RoleName    string
	InviterName string
	AcceptURL   string
	ExpiresIn   string
}

// SendInvitation sends an invitation email
func (s *sesEmailService) SendInvitation(ctx context.Context, toEmail, token, roleName, tenantName, inviterName string) error {
	data := invitationData{
		TenantName:  tenantName,
		RoleName:    roleName,
		InviterName: inviterName,
		AcceptURL:   fmt.Sprintf("%s?token=%s", s.inviteBase, token),
		ExpiresIn:   "7 days",
	}

	// Compile HTML body
	var htmlBuf bytes.Buffer
	if err := s.htmlTmpl.Execute(&htmlBuf, data); err != nil {
		return fmt.Errorf("failed to compile HTML template: %w", err)
	}

	// Compile plain-text body
	var textBuf bytes.Buffer
	if err := s.textTmpl.Execute(&textBuf, data); err != nil {
		return fmt.Errorf("failed to compile text template: %w", err)
	}

	subject := fmt.Sprintf("Invitation to join %s", tenantName)

	input := &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(s.from),
		Destination: &types.Destination{
			ToAddresses: []string{toEmail},
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{
					Data: aws.String(subject),
				},
				Body: &types.Body{
					Html: &types.Content{
						Data: aws.String(htmlBuf.String()),
					},
					Text: &types.Content{
						Data: aws.String(textBuf.String()),
					},
				},
			},
		},
	}

	_, err := s.client.SendEmail(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to send email via SES: %w", err)
	}

	return nil
}
