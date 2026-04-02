package config

import (
	"encoding/json"
	"log"
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// Config holds all application configuration
type Config struct {
	Server    ServerConfig    `json:"server"`
	Telegram  TelegramConfig  `json:"telegram"`
	Remnawave RemnawaveConfig `json:"remnawave"`
	Jellyfin  JellyfinConfig  `json:"jellyfin"`
	BEPusdt   BEPusdtConfig   `json:"bepusdt"`
	EZPay     EZPayConfig     `json:"ezpay"`
	Credit    CreditConfig    `json:"credit"`
	AI        AIConfig        `json:"ai"`
	Backup    BackupConfig    `json:"backup"`
	IPChange  IPChangeConfig  `json:"ip_change"`
}

type ServerConfig struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	APISecret string `json:"api_secret"`
}

type TelegramConfig struct {
	BotToken   string  `json:"bot_token"`
	AdminIDs   []int64 `json:"admin_ids"`
	GroupID    int64   `json:"group_id"`
	WebhookURL string  `json:"webhook_url"`
}

type RemnawaveConfig struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

type JellyfinConfig struct {
	URL             string  `json:"url"`
	Token           string  `json:"token"`
	MonthlyPriceRMB float64 `json:"monthly_price_rmb"`
}

type BEPusdtConfig struct {
	URL         string `json:"url"`
	Token       string `json:"token"`
	NotifyURL   string `json:"notify_url"`
	RedirectURL string `json:"redirect_url"`
}

type EZPayConfig struct {
	URL       string `json:"url"`
	PID       int    `json:"pid"`
	Key       string `json:"key"`
	NotifyURL string `json:"notify_url"`
	ReturnURL string `json:"return_url"`
}

type CreditConfig struct {
	Name             string  `json:"name"`
	SignupMin        float64 `json:"signup_min"`
	SignupMax        float64 `json:"signup_max"`
	RMBToTXBRate     float64 `json:"rmb_to_txb_rate"`
	TXBToRMBRate     float64 `json:"txb_to_rmb_rate"`
	BetLossMultiplier float64 `json:"bet_loss_multiplier"`
	BetWinMultiplier  float64 `json:"bet_win_multiplier"`
	LogRetentionDays  int     `json:"log_retention_days"`
}

type AIConfig struct {
	Enabled            bool    `json:"enabled"`
	BaseURL            string  `json:"base_url"`
	APIKey             string  `json:"api_key"`
	Model              string  `json:"model"`
	MessageBatchSize   int     `json:"message_batch_size"`
	CreditMin          float64 `json:"credit_min"`
	CreditMax          float64 `json:"credit_max"`
	LeaderboardInterval int    `json:"leaderboard_interval"`
}

type BackupConfig struct {
	MaxDays   int    `json:"max_days"`
	BackupDir string `json:"backup_dir"`
}

type IPChangeConfig struct {
	CooldownHours int `json:"cooldown_hours"`
}

var (
	current *Config
	mu      sync.RWMutex
	path    string
)

// Load reads config from the given file path
func Load(filePath string) (*Config, error) {
	path = filePath
	cfg, err := readFromFile(filePath)
	if err != nil {
		return nil, err
	}
	mu.Lock()
	current = cfg
	mu.Unlock()
	return cfg, nil
}

// Get returns the current config (thread-safe)
func Get() *Config {
	mu.RLock()
	defer mu.RUnlock()
	return current
}

// Save writes the current config back to the file
func Save() error {
	mu.RLock()
	defer mu.RUnlock()
	return writeToFile(path, current)
}

// Update applies changes and saves
func Update(fn func(cfg *Config)) error {
	mu.Lock()
	fn(current)
	mu.Unlock()
	return Save()
}

// WatchConfig watches the config file for changes and hot-reloads
func WatchConfig() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("[config] failed to create watcher: %v", err)
		return
	}

	go func() {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) {
					cfg, err := readFromFile(path)
					if err != nil {
						log.Printf("[config] hot-reload failed: %v", err)
						continue
					}
					mu.Lock()
					current = cfg
					mu.Unlock()
					log.Println("[config] hot-reloaded successfully")
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("[config] watcher error: %v", err)
			}
		}
	}()

	if err := watcher.Add(path); err != nil {
		log.Printf("[config] failed to watch file: %v", err)
	}
}

func readFromFile(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func writeToFile(filePath string, cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}
