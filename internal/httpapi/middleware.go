package httpapi

import (
	"net/http"
	"time"
)

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		// frame-ancestors 允许 Telegram Mini App 嵌入（telegram.org / *.telegram.org）。
		// 不再设置 X-Frame-Options，因为该头部不支持指定多个来源。
		w.Header().Set("Content-Security-Policy", "default-src 'self' https://telegram.org https://*.telegram.org; img-src 'self' data: https:; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline' https://telegram.org https://*.telegram.org; connect-src 'self' https:; frame-ancestors 'self' https://telegram.org https://*.telegram.org")
		next.ServeHTTP(w, r)
	})
}

func requestBodyLimit(limit int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, limit)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// client is a shared HTTP client with a reasonable timeout for external calls.
var telegramHTTPClient = &http.Client{Timeout: 30 * time.Second}
