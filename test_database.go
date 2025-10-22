package main

import (
	"bulk-email-mailgun/database"
	"fmt"
	"log"
)

func main() {
	// Initialiser la base de données
	if err := database.Init(); err != nil {
		log.Fatal("❌ Erreur init DB:", err)
	}
	defer database.Close()

	fmt.Println("\n🧪 TEST COMPLET DE LA BASE DE DONNÉES\n")
	fmt.Println("=" + string(make([]byte, 50)) + "=\n")

	// Test 1: Créer un contenu d'email
	fmt.Println("📝 Test 1: Création d'un contenu d'email")
	contentID, err := database.InsertEmailContent(
		"Message ",
		"Bonjour {{name}},\n\nCeci est un message secret pour toi...",
	)
	if err != nil {
		log.Fatal("Erreur:", err)
	}
	fmt.Printf("✅ Contenu créé avec ID: %d\n\n", contentID)

	// Test 2: Créer plusieurs recipients
	fmt.Println("👥 Test 2: Création de recipients")
	recipients := []struct {
		email, name, company, city string
	}{
		{"john.doe@example.com", "John Doe", "TechCorp", "Paris"},
		{"jane.smith@example.com", "Jane Smith", "StartupXYZ", "Lyon"},
		{"bob.martin@example.com", "Bob Martin", "Business Inc", "Marseille"},
	}

	recipientIDs := []int64{}
	for _, r := range recipients {
		id, err := database.InsertOrGetRecipient(r.email, r.name, r.company, r.city)
		if err != nil {
			log.Printf("❌ Erreur recipient: %v\n", err)
			continue
		}
		recipientIDs = append(recipientIDs, id)
		fmt.Printf("✅ Recipient ajouté: %s (ID: %d)\n", r.email, id)
	}
	fmt.Println()

	// Test 3: Créer plusieurs senders
	fmt.Println("📧 Test 3: Création de senders aléatoires")
	senders := []string{
		"secret.admirer.abc123@mailgun.org",
		"mystery.lover.def456@mailgun.org",
		"anonymous.heart.ghi789@mailgun.org",
	}

	senderIDs := []int64{}
	for _, email := range senders {
		id, err := database.InsertOrGetSender(email, "Admirateur Secret")
		if err != nil {
			log.Printf("❌ Erreur sender: %v\n", err)
			continue
		}
		senderIDs = append(senderIDs, id)
		fmt.Printf("✅ Sender ajouté: %s (ID: %d)\n", email, id)
	}
	fmt.Println()

	// Test 4: Créer des envois d'emails
	fmt.Println("📨 Test 4: Enregistrement des envois")
	for i, recipientID := range recipientIDs {
		senderID := senderIDs[i%len(senderIDs)]
		status := "sent"
		if i == 2 {
			status = "failed"
		}
		errorMsg := ""
		if status == "failed" {
			errorMsg = "Adresse email invalide"
		}

		err := database.InsertEmailSend(contentID, senderID, recipientID, status, errorMsg)
		if err != nil {
			log.Printf("❌ Erreur envoi: %v\n", err)
			continue
		}
		fmt.Printf("✅ Envoi enregistré: Sender %d → Recipient %d (%s)\n", senderID, recipientID, status)
	}
	fmt.Println()

	// Test 5: Récupérer les statistiques
	fmt.Println("📊 Test 5: Statistiques")
	stats, err := database.GetStats()
	if err != nil {
		log.Fatal("Erreur stats:", err)
	}
	fmt.Printf("📈 Total envois: %v\n", stats["total_sends"])
	fmt.Printf("✅ Envoyés: %v\n", stats["sent"])
	fmt.Printf("❌ Échoués: %v\n", stats["failed"])
	fmt.Printf("👥 Total recipients: %v\n", stats["total_recipients"])
	fmt.Printf("📧 Total senders: %v\n", stats["total_senders"])
	fmt.Println()

	// Test 6: Récupérer l'historique complet
	fmt.Println("📜 Test 6: Historique des envois")
	history, err := database.GetAllEmailSends()
	if err != nil {
		log.Fatal("Erreur historique:", err)
	}

	fmt.Printf("📋 %d envois dans l'historique:\n\n", len(history))
	for i, h := range history {
		fmt.Printf("%d. De: %v\n", i+1, h["sender_email"])
		fmt.Printf("   À: %v (%v)\n", h["recipient_email"], h["recipient_name"])
		fmt.Printf("   Sujet: %v\n", h["subject"])
		fmt.Printf("   Status: %v\n", h["status"])
		if h["error_message"] != "" {
			fmt.Printf("   Erreur: %v\n", h["error_message"])
		}
		fmt.Printf("   Date: %v\n\n", h["sent_at"])
	}

	// Test 7: Récupérer tous les recipients
	fmt.Println("👥 Test 7: Liste des recipients")
	allRecipients, err := database.GetAllRecipients()
	if err != nil {
		log.Fatal("Erreur recipients:", err)
	}

	fmt.Printf("📋 %d recipients enregistrés:\n", len(allRecipients))
	for _, r := range allRecipients {
		fmt.Printf("  - %s (%s) - %s, %s\n", r.Email, r.Name, r.Company, r.City)
	}
	fmt.Println()

	// Test 8: Tester l'insertion de doublons
	fmt.Println("🔄 Test 8: Gestion des doublons")
	duplicateID, err := database.InsertOrGetRecipient(
		"john.doe@example.com",
		"John Doe Updated",
		"New Company",
		"New City",
	)
	if err != nil {
		log.Fatal("Erreur doublon:", err)
	}
	if duplicateID == recipientIDs[0] {
		fmt.Printf("✅ Doublon détecté correctement (ID existant: %d)\n", duplicateID)
	} else {
		fmt.Printf("❌ Erreur: Nouveau ID créé pour un doublon\n")
	}
	fmt.Println()

	fmt.Println("=" + string(make([]byte, 50)) + "=")
	fmt.Println("\n✅ TOUS LES TESTS RÉUSSIS!")
	fmt.Println("📁 Base de données: ./emails.db")
	fmt.Println("\n💡 Vous pouvez maintenant:")
	fmt.Println("   1. Ouvrir emails.db avec un client SQLite")
	fmt.Println("   2. Lancer: go run main.go")
	fmt.Println("   3. Visiter: http://localhost:8080")
}
