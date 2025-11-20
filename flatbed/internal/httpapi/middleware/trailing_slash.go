package middleware

import (
	"net/http"
	"strings"
)

// StripTrailingSlashForBuckets strips trailing slashes from bucket-only paths
// (e.g., /bucket/) but preserves them in object paths (e.g., /bucket/key/).
// This is critical for S3 compatibility where /bucket/folder/ and /bucket/folder
// are different objects.
func StripTrailingSlashForBuckets(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Don't strip trailing slash from root
		if path == "/" {
			next.ServeHTTP(w, r)
			return
		}

		// Count slashes to determine if this is a bucket-only path
		slashCount := strings.Count(path, "/")

		// If path is /something/ (exactly 2 slashes, starts and ends with /),
		// this is a bucket path with trailing slash - strip it
		if slashCount == 2 && strings.HasPrefix(path, "/") && strings.HasSuffix(path, "/") {
			r.URL.Path = strings.TrimSuffix(path, "/")
		}

		// For paths with more slashes (object keys), preserve trailing slashes
		// because /bucket/folder/ and /bucket/folder are different objects in S3

		next.ServeHTTP(w, r)
	})
}
