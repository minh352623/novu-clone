package global

import (
	"database/sql"
	"net/http"

	"CONVERDA/pkg/logger"
	"CONVERDA/pkg/setting"

	"gorm.io/gorm"
)

var (
	Config setting.Config
	Logger *logger.Logger
	Http   *http.Client = &http.Client{}
	Pdbc   *sql.DB
	GormDB *gorm.DB
)
