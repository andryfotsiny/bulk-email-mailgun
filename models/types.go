package models

type EmailConfig struct {
	SMTPServer    string `json:"smtp_server"`
	SMTPPort      int    `json:"smtp_port"`
	Email         string `json:"email"`
	Password      string `json:"password"`
	Provider      string `json:"provider"`
	MailgunDomain string `json:"mailgun_domain"`
	MailgunAPIKey string `json:"mailgun_api_key"`
}

type EmailData struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Company string `json:"company"`
	City    string `json:"city"`
}

type SendRequest struct {
	Emails  []EmailData `json:"emails"`
	Subject string      `json:"subject"`
	Body    string      `json:"body"`
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
}

type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}
