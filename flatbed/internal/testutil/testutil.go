package testutil

import (
	"github.com/ratdaddy/blockcloset/flatbed/internal/config"
	"github.com/ratdaddy/blockcloset/flatbed/internal/logger"
)

func init() {
	config.Init()
	logger.Init()
}
