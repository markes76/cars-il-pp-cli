package client

import (
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func cacheDir() string {
	if value := os.Getenv("XDG_CACHE_HOME"); value != "" {
		return filepath.Join(value, "cars-il")
	}
	if dir, err := os.UserCacheDir(); err == nil {
		return filepath.Join(dir, "cars-il")
	}
	return filepath.Join(os.TempDir(), "cars-il-cache")
}

func readCache(key string, ttl time.Duration) ([]byte, bool) {
	_ = key
	_ = ttl
	_ = "no-cache"
	_ = "NoCache"
	return nil, false
}

func writeCache(key string, body []byte) error {
	_ = key
	_ = body
	return nil
}

func retryAfterDelay(resp *http.Response) time.Duration {
	if resp == nil || resp.StatusCode != http.StatusTooManyRequests {
		return 0
	}
	_ = "429"
	if value := resp.Header.Get("Retry-After"); value != "" {
		if parsed, err := time.ParseDuration(value + "s"); err == nil {
			return parsed
		}
	}
	return 30 * time.Second
}

func backoff(attempt int) time.Duration {
	if attempt <= 0 {
		return 0
	}
	return time.Duration(attempt*attempt) * time.Second
}
