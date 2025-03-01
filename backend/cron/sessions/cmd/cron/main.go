package main

import (
	"log"
	"time"

	"pariksha/common/pkg/models"
	"pariksha/cron/sessions/internal/config/db"
	"pariksha/cron/sessions/internal/config/env"
)

func cleanupExpiredSessions() error {
	result := db.Sessions.Where("expires_at < ?", time.Now()).Delete(&models.Session{})
	if result.Error != nil {
		return result.Error
	}
	log.Printf("Cleaned up %d expired sessions", result.RowsAffected)
	return nil
}

func main() {
	interval := env.CRON_INTERVAL_HOURS

	log.Printf("Starting sessions cleanup cron job with %d hour interval", interval)

	ticker := time.NewTicker(time.Duration(interval) * time.Hour)
	defer ticker.Stop()

	// Run cleanup immediately on startup
	if err := cleanupExpiredSessions(); err != nil {
		log.Printf("Error during initial cleanup: %v", err)
	}

	// Then run periodically
	for range ticker.C {
		if err := cleanupExpiredSessions(); err != nil {
			log.Printf("Error during cleanup: %v", err)
		}
	}
}
