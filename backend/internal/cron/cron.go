package cron

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/user/remna-user-panel/internal/config"
	"github.com/user/remna-user-panel/internal/database"
	"github.com/user/remna-user-panel/internal/sdk/jellyfin"
	"github.com/user/remna-user-panel/internal/services"
)

// Start initializes and starts all cron jobs
func Start(credit *services.CreditService, payment *services.PaymentService) {
	c := cron.New()

	// Daily backup at 3:00 AM
	c.AddFunc("0 3 * * *", func() {
		backup()
	})

	// Daily credit log cleanup at 4:00 AM
	c.AddFunc("0 4 * * *", func() {
		if err := credit.CleanupOldLogs(); err != nil {
			slog.Error("cron: credit cleanup failed", "error", err)
		}
	})

	// Check Jellyfin account expiry every hour
	c.AddFunc("0 * * * *", func() {
		cleanupExpiredJellyfin()
	})

	// Check subscription expiry every hour
	c.AddFunc("30 * * * *", func() {
		checkExpiredSubscriptions()
	})

	// Cancel stale pending payments every 5 minutes.
	c.AddFunc("@every 5m", func() {
		if payment == nil {
			return
		}
		if err := payment.CancelExpiredPendingOrders(); err != nil {
			slog.Error("cron: pending payment cleanup failed", "error", err)
		}
	})

	c.Start()
	slog.Info("cron: scheduled tasks started")
}

func backup() {
	cfg := config.Get()
	backupDir := cfg.Backup.BackupDir
	if backupDir == "" {
		backupDir = "./backups"
	}

	timestamp := time.Now().Format("20060102_150405")

	// Backup database
	dbBackupPath := filepath.Join(backupDir, fmt.Sprintf("db_%s.sqlite3", timestamp))
	if err := database.Backup(dbBackupPath); err != nil {
		slog.Error("backup: database backup failed", "error", err)
	} else {
		slog.Info("backup: database backed up", "path", dbBackupPath)
	}

	// Backup config
	configData, err := os.ReadFile("config.json")
	if err == nil {
		configBackupPath := filepath.Join(backupDir, fmt.Sprintf("config_%s.json", timestamp))
		os.MkdirAll(backupDir, 0755)
		os.WriteFile(configBackupPath, configData, 0644)
		slog.Info("backup: config backed up", "path", configBackupPath)
	}

	// Clean up old backups (> max_days)
	cleanupOldBackups(backupDir, cfg.Backup.MaxDays)
}

func cleanupOldBackups(dir string, maxDays int) {
	if maxDays <= 0 {
		maxDays = 10
	}
	cutoff := time.Now().AddDate(0, 0, -maxDays)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			path := filepath.Join(dir, e.Name())
			if err := os.Remove(path); err == nil {
				slog.Info("backup: removed old backup", "path", path)
			}
		}
	}
}

func cleanupExpiredJellyfin() {
	rows, err := database.DB().Query(
		"SELECT id, user_id, jellyfin_user_id FROM jellyfin_accounts WHERE expires_at < ?",
		time.Now(),
	)
	if err != nil {
		slog.Error("cron: query expired jellyfin accounts", "error", err)
		return
	}
	defer rows.Close()

	cfg := config.Get()
	for rows.Next() {
		var id, userID int64
		var jfUserID string
		if err := rows.Scan(&id, &userID, &jfUserID); err != nil {
			slog.Error("cron: scan jellyfin row", "error", err)
			continue
		}

		// Delete from Jellyfin
		jfClient := jellyfinClient(cfg)
		if jfClient != nil {
			if err := jfClient.DeleteUser(jfUserID); err != nil {
				slog.Error("cron: delete jellyfin user failed", "jellyfin_id", jfUserID, "error", err)
				continue
			}
		}

		// Remove from database
		database.DB().Exec("DELETE FROM jellyfin_accounts WHERE id = ?", id)
		database.DB().Exec("UPDATE users SET jellyfin_user_id = '' WHERE id = ?", userID)
		slog.Info("cron: expired Jellyfin account removed", "user_id", userID)
	}
	if err := rows.Err(); err != nil {
		slog.Error("cron: jellyfin cleanup iteration error", "error", err)
	}
}

func checkExpiredSubscriptions() {
	if _, err := database.DB().Exec(
		"UPDATE subscriptions SET status = 'expired', updated_at = ? WHERE status = 'active' AND expires_at < ?",
		time.Now(), time.Now(),
	); err != nil {
		slog.Error("cron: check expired subscriptions failed", "error", err)
	}
}

// Helper to create Jellyfin client
func jellyfinClient(cfg *config.Config) interface{ DeleteUser(string) error } {
	if cfg.Jellyfin.URL == "" || cfg.Jellyfin.Token == "" {
		return nil
	}
	return jellyfin.NewClient(cfg.Jellyfin.URL, cfg.Jellyfin.Token)
}
