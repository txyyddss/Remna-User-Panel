// Package i18n loads JSON locale catalogs and resolves translated strings.
package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
		body, err := os.ReadFile(filepath.Join(localesDir, entry.Name()))
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

// Translate returns a localized message with zh -> en fallback.
func (c *Catalog) Translate(lang string, key string) string {
	lang = normalizeLanguage(lang)
	for _, candidate := range []string{lang, c.defaultLang, "zh", "en"} {
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

func normalizeLanguage(raw string) string {
	value := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(raw), "_", "-"))
	if value == "" {
		return "zh"
	}
	return value
}
