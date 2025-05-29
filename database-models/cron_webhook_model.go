package database_models

import (
	"log"
	"net/http"

	"gorm.io/gorm"
)

type CronWebhookModel struct {
	gorm.Model
	URL     string `gorm:"not null;unique"`
	Enabled bool   `gorm:"default:true"`
	Cron    string `gorm:"not null"`
}

// Handle websocket run from model
func RunWebhook(db *gorm.DB, m *CronWebhookModel) error {
	// send HTTP request to the webhook URL
	resp, err := http.Post(m.URL, "application/json", nil)
	if err != nil {
		// log the error
		log.Printf("failed to send request to webhook %s: %v", m.URL, err)

		return nil
	}
	defer resp.Body.Close()
	// check if the response status is OK
	if resp.StatusCode != http.StatusOK {
		log.Printf("webhook %s returned non-OK status: %s", m.URL, resp.Status)
		return nil
	}

	return nil
}
