package email

import (
	"context"
	"fmt"

	"CONVERDA/internal/iam/application/service"
	"CONVERDA/pkg/setting"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

type sesEmailService struct {
	client *sesv2.Client
	from   string
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

	return &sesEmailService{
		client: client,
		from:   sesSetting.FromEmail,
	}, nil
}

// SendInvitation sends an invitation email
func (s *sesEmailService) SendInvitation(ctx context.Context, toEmail, token, roleName, tenantName string) error {
	subject := "Invitation to join " + tenantName
	htmlBody := fmt.Sprintf(`
		<h1>You have been invited to join %s</h1>
		<p>Role: <strong>%s</strong></p>
		<p>Click the link below to accept the invitation:</p>
		<a href="https://app.converda.com/invite?token=%s">Accept Invitation</a>
		<p>This link expires in 7 days.</p>
	`, tenantName, roleName, token)

	textBody := fmt.Sprintf("You have been invited to join %s as %s. Token: %s", tenantName, roleName, token)

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
						Data: aws.String(htmlBody),
					},
					Text: &types.Content{
						Data: aws.String(textBody),
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
