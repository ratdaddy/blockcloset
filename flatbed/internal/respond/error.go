package respond

import (
	"net/http"

	"github.com/ratdaddy/blockcloset/flatbed/internal/logger"
)

func Error(w http.ResponseWriter, r *http.Request, msg string, status int) {
	logger.LogError(w, r, msg)
	http.Error(w, msg, status)
}
