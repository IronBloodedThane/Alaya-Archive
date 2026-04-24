package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Mailer struct {
	apiKey    string
	fromEmail string
	fromName  string
	client    *http.Client
}

func NewMailer(apiKey, fromEmail string) *Mailer {
	return &Mailer{
		apiKey:    apiKey,
		fromEmail: fromEmail,
		fromName:  "Alaya Archive",
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

type sgMessage struct {
	Personalizations []sgPersonalization `json:"personalizations"`
	From             sgAddr              `json:"from"`
	Subject          string              `json:"subject"`
	Content          []sgContent         `json:"content"`
}

type sgPersonalization struct {
	To []sgAddr `json:"to"`
}

type sgAddr struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type sgContent struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func (m *Mailer) Send(to, subject, textBody, htmlBody string) error {
	if m.apiKey == "" || m.fromEmail == "" {
		log.Printf("email not configured; would have sent to %s: %s", to, subject)
		return nil
	}

	msg := sgMessage{
		Personalizations: []sgPersonalization{{To: []sgAddr{{Email: to}}}},
		From:             sgAddr{Email: m.fromEmail, Name: m.fromName},
		Subject:          subject,
		Content: []sgContent{
			{Type: "text/plain", Value: textBody},
			{Type: "text/html", Value: htmlBody},
		},
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.sendgrid.com/v3/mail/send", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("sendgrid returned %d: %s", resp.StatusCode, respBody)
	}
	return nil
}

func (m *Mailer) SendVerification(to, verifyURL string) error {
	subject := "Verify your Alaya Archive account"
	text := fmt.Sprintf("Welcome to Alaya Archive!\n\nPlease verify your email by visiting:\n\n%s\n\nThis link expires in 24 hours. If you didn't create an account, you can ignore this message.", verifyURL)
	html := fmt.Sprintf(`<p>Welcome to <strong>Alaya Archive</strong>!</p>
<p>Please verify your email by clicking the link below:</p>
<p><a href="%s">Verify email</a></p>
<p>Or copy this URL into your browser:<br><code>%s</code></p>
<p>This link expires in 24 hours. If you didn't create an account, you can ignore this message.</p>`, verifyURL, verifyURL)
	return m.Send(to, subject, text, html)
}

func (m *Mailer) SendPasswordReset(to, resetURL string) error {
	subject := "Reset your Alaya Archive password"
	text := fmt.Sprintf("Someone requested a password reset for your account. If that was you, visit:\n\n%s\n\nThis link expires in 1 hour. If you didn't request it, ignore this email.", resetURL)
	html := fmt.Sprintf(`<p>Someone requested a password reset for your Alaya Archive account. If that was you, click the link below:</p>
<p><a href="%s">Reset password</a></p>
<p>This link expires in 1 hour. If you didn't request it, you can safely ignore this email.</p>`, resetURL)
	return m.Send(to, subject, text, html)
}
