package initialize

import (
	"CONVERDA/global"
	"CONVERDA/pkg/logger"
)

func InitLogger() {
	global.Logger = logger.NewLogger(global.Config.Logger)
}
