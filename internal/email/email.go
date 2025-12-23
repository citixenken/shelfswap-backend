package email

import (
	"fmt"
	"log"

	"github.com/resend/resend-go/v2"
)

type EmailService interface {
	SendRequestNotification(toEmail, ownerName, bookTitle, requesterEmail string) error
	SendPasswordReset(to, token string) error
	SendContactEmail(fromEmail, subject, body string) error
}

type ConsoleEmailService struct{}

func NewConsoleEmailService() *ConsoleEmailService {
	return &ConsoleEmailService{}
}

func (s *ConsoleEmailService) SendPasswordReset(to, token string) error {
	log.Printf("--------------------------------------------------")
	log.Printf("EMAIL SENT (Console)")
	log.Printf("To: %s", to)
	log.Printf("Subject: Password Reset Request")
	log.Printf("Body: Click here to reset your password: http://localhost:5173/reset-password?token=%s", token)
	log.Printf("--------------------------------------------------")
	return nil
}

func (s *ConsoleEmailService) SendRequestNotification(toEmail, ownerName, bookTitle, requesterEmail string) error {
	log.Printf("--------------------------------------------------")
	log.Printf("SENDING EMAIL")
	log.Printf("To: %s", toEmail)
	log.Printf("Subject: New Book Request: %s", bookTitle)
	log.Printf("Body: Hi %s,\n\nYou have a new request for your book '%s' from %s.\n\nIf you're interested in swapping, please reach out to them directly at %s to arrange a meeting place and time for the exchange.\n\nCheers,\nThe ShelfSwap Team", ownerName, bookTitle, requesterEmail, requesterEmail)
	log.Printf("--------------------------------------------------")
	return nil
}

func (s *ConsoleEmailService) SendContactEmail(fromEmail, subject, body string) error {
	log.Printf("--------------------------------------------------")
	log.Printf("SENDING CONTACT EMAIL")
	log.Printf("To: admin@shelfswap.com")
	log.Printf("From: %s", fromEmail)
	log.Printf("Subject: %s", subject)
	log.Printf("Body: %s", body)
	log.Printf("--------------------------------------------------")
	return nil
}

type ResendEmailService struct {
	client *resend.Client
}

func NewResendEmailService(apiKey string) *ResendEmailService {
	client := resend.NewClient(apiKey)
	return &ResendEmailService{
		client: client,
	}
}

func (s *ResendEmailService) SendRequestNotification(toEmail, ownerName, bookTitle, requesterEmail string) error {
	htmlBody := fmt.Sprintf(`
		<p>Hi %s,</p>
		<p>You have a new request for your book <strong><em>%s</em></strong> from <strong>%s</strong>.</p>
		<p>If you're interested in swapping, please reach out to them directly at <a href="mailto:%s">%s</a> to arrange a convenient meeting place and time for the exchange.</p>
		<p>Cheers,<br>The ShelfSwap Team</p>
	`, ownerName, bookTitle, requesterEmail, requesterEmail, requesterEmail)

	params := &resend.SendEmailRequest{
		From:    "ShelfSwap Team <hello@shelfswap.io>",
		To:      []string{toEmail},
		Subject: fmt.Sprintf("New Book Request: %s", bookTitle),
		Html:    htmlBody,
	}

	_, err := s.client.Emails.Send(params)
	if err != nil {
		log.Printf("Error sending email: %v", err)
		return err
	}
	return nil
}

func (s *ResendEmailService) SendPasswordReset(to, token string) error {
	params := &resend.SendEmailRequest{
		From:    "ShelfSwap <onboarding@resend.dev>",
		To:      []string{to},
		Subject: "Password Reset Request",
		Html:    fmt.Sprintf("<p>Click here to reset your password: <a href=\"http://localhost:5173/reset-password?token=%s\">Reset Password</a></p>", token),
	}

	_, err := s.client.Emails.Send(params)
	if err != nil {
		log.Printf("Error sending email: %v", err)
		return err
	}
	return nil
}

func (s *ResendEmailService) SendContactEmail(fromEmail, subject, body string) error {
	params := &resend.SendEmailRequest{
		From:    "ShelfSwap Contact Form <hello@shelfswap.io>",
		To:      []string{"hello@shelfswap.io"},
		Subject: fmt.Sprintf("Contact Form: %s", subject),
		Html:    fmt.Sprintf("<p>Feedback from: %s</p><p>%s</p>", fromEmail, body),
	}

	_, err := s.client.Emails.Send(params)
	if err != nil {
		log.Printf("Error sending email: %v", err)
		return err
	}
	return nil
}
