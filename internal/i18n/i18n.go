// Package i18n loads JSON locale catalogs and resolves translated strings.
package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"remna-user-panel/internal/config"
)

// Catalog stores locale messages keyed by language code.
type Catalog struct {
	defaultLang string
	messages    map[string]map[string]string
}

// Load reads locale JSON files from localesDir.
func Load(localesDir string, defaultLang string) (*Catalog, error) {
	catalog := &Catalog{
		defaultLang: normalizeLanguage(defaultLang),
		messages:    map[string]map[string]string{},
	}
	entries, err := os.ReadDir(localesDir)
	if err != nil {
		return nil, fmt.Errorf("read locales dir: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		body, err := os.ReadFile(filepath.Join(localesDir, entry.Name())) //nolint:gosec // G304: entry from os.ReadDir, not user input.
		if err != nil {
			return nil, fmt.Errorf("read locale %s: %w", entry.Name(), err)
		}
		bucket := map[string]string{}
		if err := json.Unmarshal(body, &bucket); err != nil {
			return nil, fmt.Errorf("decode locale %s: %w", entry.Name(), err)
		}
		lang := normalizeLanguage(strings.TrimSuffix(entry.Name(), ".json"))
		catalog.messages[lang] = bucket
	}
	return catalog, nil
}

// Translate returns a localized message with language fallback chain.
func (c *Catalog) Translate(lang string, key string) string {
	lang = normalizeLanguage(lang)
	seen := map[string]bool{}
	candidates := []string{}
	// Priority: requested lang, then defaultLang, then all other loaded languages.
	for _, candidate := range []string{lang, c.defaultLang} {
		if !seen[candidate] {
			candidates = append(candidates, candidate)
			seen[candidate] = true
		}
	}
	for l := range c.messages {
		if !seen[l] {
			candidates = append(candidates, l)
			seen[l] = true
		}
	}
	for _, candidate := range candidates {
		if value := c.messages[candidate][key]; value != "" {
			return value
		}
	}
	return key
}

// Languages returns loaded language codes.
func (c *Catalog) Languages() []string {
	result := make([]string, 0, len(c.messages))
	for lang := range c.messages {
		result = append(result, lang)
	}
	return result
}

// Messages returns a copy of all loaded locale messages.
func (c *Catalog) Messages() map[string]map[string]string {
	result := make(map[string]map[string]string, len(c.messages))
	for lang, messages := range c.messages {
		result[lang] = make(map[string]string, len(messages))
		for key, value := range messages {
			result[lang][key] = value
		}
	}
	return result
}

func normalizeLanguage(raw string) string {
	return config.NormalizeLanguage(raw)
}
