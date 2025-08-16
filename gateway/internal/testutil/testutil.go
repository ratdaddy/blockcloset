package testutil

import (
	"github.com/ratdaddy/blockcloset/gateway/internal/config"
	"github.com/ratdaddy/blockcloset/gateway/internal/logger"
)

func init() {
	config.Init()
	logger.Init()
}
