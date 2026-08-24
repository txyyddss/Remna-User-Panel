package abuse

import (
	"crypto/sha256"
	"encoding/base64"
	"regexp"
)

type Service struct {
	repo  Repository
	nodes NodeProvider
}

func NewService(repo Repository, nodes NodeProvider) *Service {
	return &Service{repo: repo, nodes: nodes}
}

type compiledRule struct {
	DomainRule
	regex *regexp.Regexp
}

func compileRules(rules []DomainRule) ([]compiledRule, error) {
	out := make([]compiledRule, 0, len(rules))
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		re, err := regexp.Compile(rule.Expression)
		if err != nil {
			return nil, ErrInvalid
		}
		out = append(out, compiledRule{rule, re})
	}
	return out, nil
}
func tokenDigest(token string) string {
	sum := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func reachesLimit(qps, limit int) bool { return qps >= limit }
func unique(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value != "" && !seen[value] {
			seen[value] = true
			out = append(out, value)
		}
	}
	return out
}
