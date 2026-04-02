package config

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestConfigDataRace(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Write an initial config file
	initialCfg := `{"server": {"port": 8080}, "ai": {"enabled": true}}`
	err := os.WriteFile(configPath, []byte(initialCfg), 0644)
	if err != nil {
		t.Fatalf("failed to write initial config: %v", err)
	}

	// Load the config
	_, err = Load(configPath)
	if err != nil {
		t.Fatalf("failed to load initial config: %v", err)
	}

	var wg sync.WaitGroup

	// Start 10 goroutines reading the config
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				cfg := Get()
				_ = cfg.Server.Port
				_ = cfg.AI.Enabled
			}
		}()
	}

	// Start 1 goroutine updating the config
	wg.Add(1)
	go func() {
		defer wg.Done()
		for j := 0; j < 50; j++ {
			err := Update(func(cfg *Config) {
				cfg.Server.Port = 9090 + j
				cfg.AI.Enabled = !cfg.AI.Enabled
			})
			if err != nil {
				t.Errorf("update failed: %v", err)
			}
		}
	}()

	wg.Wait()
}
