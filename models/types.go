package models

type EmailConfig struct {
	SMTPServer    string `json:"smtp_server"`
	SMTPPort      int    `json:"smtp_port"`
	Email         string `json:"email"`
	Password      string `json:"password"`
	Provider      string `json:"provider"`
	MailgunDomain string `json:"mailgun_domain"`
	MailgunAPIKey string `json:"mailgun_api_key"`

	// ✅ Ajout Resend
	ResendAPIKey    string `json:"resend_api_key"`
	ResendFromEmail string `json:"resend_from_email"`
}

type EmailData struct {
	Email string `json:"email"`
}

type SendRequest struct {
	Emails   []EmailData `json:"emails"`
	Subject  string      `json:"subject"`
	Body     string      `json:"body"`
	Provider string      `json:"provider"` // "mailgun", "resend"
}

type ProgressUpdate struct {
	Current    int     `json:"current"`
	Total      int     `json:"total"`
	Sent       int     `json:"sent"`
	Failed     int     `json:"failed"`
	Percentage float64 `json:"percentage"`
}

type UploadResponse struct {
	Success bool        `json:"success"`
	Count   int         `json:"count,omitempty"`
	Emails  []EmailData `json:"emails,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

type ConfigResponse struct {
	SMTPServer    string `json:"smtp_server"`
	SMTPPort      int    `json:"smtp_port"`
	Email         string `json:"email"`
	Provider      string `json:"provider"`
	MailgunDomain string `json:"mailgun_domain,omitempty"`

	ResendFromEmail string `json:"resend_from_email,omitempty"`
}

type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}
