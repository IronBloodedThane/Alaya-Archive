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

// Sender is the interface the auth handler depends on. Tests provide a fake
// implementation that records calls instead of making HTTP requests.
type Sender interface {
	SendVerification(to, verifyURL string) error
	SendPasswordReset(to, resetURL string) error
}

type Mailer struct {
	apiKey string
	from   string
	client *http.Client
}

// NewMailer creates a Resend-backed mailer. `from` may be a bare address
// (`noreply@example.com`) or a name-and-address pair (`Alaya Archive <noreply@example.com>`).
func NewMailer(apiKey, from string) *Mailer {
	if from != "" && !containsAngleBrackets(from) {
		from = fmt.Sprintf("Alaya Archive <%s>", from)
	}
	return &Mailer{
		apiKey: apiKey,
		from:   from,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func containsAngleBrackets(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '<' {
			return true
		}
	}
	return false
}

type resendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html,omitempty"`
	Text    string   `json:"text,omitempty"`
}

func (m *Mailer) Send(to, subject, textBody, htmlBody string) error {
	if m.apiKey == "" || m.from == "" {
		log.Printf("email not configured; would have sent to %s: %s", to, subject)
		return nil
	}

	body, err := json.Marshal(resendRequest{
		From:    m.from,
		To:      []string{to},
		Subject: subject,
		HTML:    htmlBody,
		Text:    textBody,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewReader(body))
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
		return fmt.Errorf("resend returned %d: %s", resp.StatusCode, respBody)
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
