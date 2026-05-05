package email

import (
	"fmt"
	"os"

	resend "github.com/resend/resend-go/v2"
)

type Service struct {
	client *resend.Client
	from   string
	to     string
}

func NewService() *Service {
	return &Service{
		client: resend.NewClient(os.Getenv("RESEND_API_KEY")),
		from:   os.Getenv("RESEND_FROM_EMAIL"),
		to:     os.Getenv("NOTIFY_EMAIL"),
	}
}

func (s *Service) Send(subject, html string) error {
	_, err := s.client.Emails.Send(&resend.SendEmailRequest{
		From:    s.from,
		To:      []string{s.to},
		Subject: subject,
		Html:    html,
	})
	if err != nil {
		return fmt.Errorf("resend: %w", err)
	}
	return nil
}
