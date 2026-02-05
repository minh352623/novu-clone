package initialize

import (
	"CONVERDA/global"

	"github.com/gin-gonic/gin"
)

func Run() (*gin.Engine, int) {
	// 1> Read config -> environment variables
	LoadConfig()
	InitLogger()
	InitPostgresql()

	// 3> Initialize router
	r := InitRouter(global.Pdbc)

	// 4> Initialize other services if needed (e.g., cache, message queue, etc.)
	return r, global.Config.Server.Port
}
