package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// EmailContent représente le contenu d'un email
type EmailContent struct {
	ID        int
	Subject   string
	Body      string
	CreatedAt time.Time
}

// Sender représente un expéditeur (email aléatoire généré)
type Sender struct {
	ID          int
	Email       string
	DisplayName string
	CreatedAt   time.Time
}

// Recipient représente un destinataire (importé depuis CSV)
type Recipient struct {
	ID        int
	Email     string
	Name      string
	Company   string
	City      string
	CreatedAt time.Time
}

// EmailSend représente un envoi d'email (historique)
type EmailSend struct {
	ID           int
	ContentID    int
	SenderID     int
	RecipientID  int
	Status       string
	ErrorMessage string
	SentAt       time.Time
}

// Init initialise la connexion SQLite et crée les tables
func Init() error {
	var err error

	// Créer/ouvrir la base de données
	DB, err = sql.Open("sqlite3", "./emails.db")
	if err != nil {
		return fmt.Errorf("erreur ouverture DB: %v", err)
	}

	// Tester la connexion
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("erreur ping DB: %v", err)
	}

	// Créer les tables
	if err = createTables(); err != nil {
		return fmt.Errorf("erreur création tables: %v", err)
	}

	log.Println("✅ SQLite initialisé avec succès")
	return nil
}

// createTables crée toutes les tables nécessaires
func createTables() error {
	schema := `
	-- Table des contenus d'emails
	CREATE TABLE IF NOT EXISTS email_contents (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		subject TEXT NOT NULL,
		body TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Table des expéditeurs (emails aléatoires générés)
	CREATE TABLE IF NOT EXISTS senders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE NOT NULL,
		display_name TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Table des destinataires (importés depuis CSV)
	CREATE TABLE IF NOT EXISTS recipients (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE NOT NULL,
		name TEXT,
		company TEXT,
		city TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Table d'historique des envois
	CREATE TABLE IF NOT EXISTS email_sends (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		content_id INTEGER NOT NULL,
		sender_id INTEGER NOT NULL,
		recipient_id INTEGER NOT NULL,
		status TEXT NOT NULL,
		error_message TEXT,
		sent_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (content_id) REFERENCES email_contents(id),
		FOREIGN KEY (sender_id) REFERENCES senders(id),
		FOREIGN KEY (recipient_id) REFERENCES recipients(id)
	);

	-- Index pour performances
	CREATE INDEX IF NOT EXISTS idx_content_id ON email_sends(content_id);
	CREATE INDEX IF NOT EXISTS idx_sender_id ON email_sends(sender_id);
	CREATE INDEX IF NOT EXISTS idx_recipient_id ON email_sends(recipient_id);
	CREATE INDEX IF NOT EXISTS idx_status ON email_sends(status);
	CREATE INDEX IF NOT EXISTS idx_sent_at ON email_sends(sent_at);
	CREATE INDEX IF NOT EXISTS idx_recipient_email ON recipients(email);
	CREATE INDEX IF NOT EXISTS idx_sender_email ON senders(email);
	`

	_, err := DB.Exec(schema)
	return err
}

// InsertEmailContent insère un contenu d'email et retourne son ID
func InsertEmailContent(subject, body string) (int64, error) {
	query := `INSERT INTO email_contents (subject, body) VALUES (?, ?)`
	result, err := DB.Exec(query, subject, body)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// InsertOrGetSender insère un sender ou retourne son ID s'il existe
func InsertOrGetSender(email, displayName string) (int64, error) {
	// Vérifier si le sender existe déjà
	var id int64
	query := `SELECT id FROM senders WHERE email = ?`
	err := DB.QueryRow(query, email).Scan(&id)

	if err == sql.ErrNoRows {
		// Insérer le nouveau sender
		insertQuery := `INSERT INTO senders (email, display_name) VALUES (?, ?)`
		result, err := DB.Exec(insertQuery, email, displayName)
		if err != nil {
			return 0, err
		}
		return result.LastInsertId()
	}

	return id, err
}

// InsertOrGetRecipient insère un recipient ou retourne son ID s'il existe
func InsertOrGetRecipient(email, name, company, city string) (int64, error) {
	// Vérifier si le recipient existe déjà
	var id int64
	query := `SELECT id FROM recipients WHERE email = ?`
	err := DB.QueryRow(query, email).Scan(&id)

	if err == sql.ErrNoRows {
		// Insérer le nouveau recipient
		insertQuery := `INSERT INTO recipients (email, name, company, city) VALUES (?, ?, ?, ?)`
		result, err := DB.Exec(insertQuery, email, name, company, city)
		if err != nil {
			return 0, err
		}
		return result.LastInsertId()
	}

	return id, err
}

// InsertEmailSend enregistre un envoi d'email
func InsertEmailSend(contentID, senderID, recipientID int64, status, errorMessage string) error {
	query := `
		INSERT INTO email_sends (content_id, sender_id, recipient_id, status, error_message)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := DB.Exec(query, contentID, senderID, recipientID, status, errorMessage)
	return err
}

// GetAllEmailSends récupère tous les envois avec leurs détails
func GetAllEmailSends() ([]map[string]interface{}, error) {
	query := `
		SELECT 
			es.id,
			s.email as sender_email,
			s.display_name as sender_name,
			r.email as recipient_email,
			r.name as recipient_name,
			r.company,
			r.city,
			ec.subject,
			ec.body,
			es.status,
			es.error_message,
			es.sent_at
		FROM email_sends es
		JOIN email_contents ec ON es.content_id = ec.id
		JOIN senders s ON es.sender_id = s.id
		JOIN recipients r ON es.recipient_id = r.id
		ORDER BY es.sent_at DESC
	`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var (
			id, senderEmail, senderName, recipientEmail, recipientName string
			company, city, subject, body, status, errorMessage, sentAt string
		)

		err := rows.Scan(&id, &senderEmail, &senderName, &recipientEmail,
			&recipientName, &company, &city, &subject, &body, &status, &errorMessage, &sentAt)
		if err != nil {
			return nil, err
		}

		results = append(results, map[string]interface{}{
			"id":              id,
			"sender_email":    senderEmail,
			"sender_name":     senderName,
			"recipient_email": recipientEmail,
			"recipient_name":  recipientName,
			"company":         company,
			"city":            city,
			"subject":         subject,
			"body":            body,
			"status":          status,
			"error_message":   errorMessage,
			"sent_at":         sentAt,
		})
	}

	return results, nil
}

// GetStats récupère les statistiques globales
func GetStats() (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			SUM(CASE WHEN status = 'sent' THEN 1 ELSE 0 END) as sent,
			SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed
		FROM email_sends
	`

	var total, sent, failed int
	err := DB.QueryRow(query).Scan(&total, &sent, &failed)
	if err != nil {
		return nil, err
	}

	// Compter les recipients et senders
	var recipientCount, senderCount int
	DB.QueryRow("SELECT COUNT(*) FROM recipients").Scan(&recipientCount)
	DB.QueryRow("SELECT COUNT(*) FROM senders").Scan(&senderCount)

	return map[string]interface{}{
		"total_sends":      total,
		"sent":             sent,
		"failed":           failed,
		"total_recipients": recipientCount,
		"total_senders":    senderCount,
	}, nil
}

// GetRecipientsByEmail recherche des recipients par email
func GetRecipientsByEmail(email string) ([]Recipient, error) {
	query := `SELECT id, email, name, company, city, created_at FROM recipients WHERE email LIKE ?`
	rows, err := DB.Query(query, "%"+email+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipients []Recipient
	for rows.Next() {
		var r Recipient
		err := rows.Scan(&r.ID, &r.Email, &r.Name, &r.Company, &r.City, &r.CreatedAt)
		if err != nil {
			return nil, err
		}
		recipients = append(recipients, r)
	}

	return recipients, nil
}

// GetAllRecipients récupère tous les recipients
func GetAllRecipients() ([]Recipient, error) {
	query := `SELECT id, email, name, company, city, created_at FROM recipients ORDER BY created_at DESC`
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipients []Recipient
	for rows.Next() {
		var r Recipient
		err := rows.Scan(&r.ID, &r.Email, &r.Name, &r.Company, &r.City, &r.CreatedAt)
		if err != nil {
			return nil, err
		}
		recipients = append(recipients, r)
	}

	return recipients, nil
}

// DeleteOldSends supprime les envois plus vieux que X jours
func DeleteOldSends(days int) (int64, error) {
	query := `DELETE FROM email_sends WHERE sent_at < datetime('now', '-' || ? || ' days')`
	result, err := DB.Exec(query, days)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Close ferme la connexion à la base de données
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
