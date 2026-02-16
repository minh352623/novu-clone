package initialize

import (
	"database/sql"
	"fmt"
	"time"

	"CONVERDA/global"
	"CONVERDA/internal/infrastructure/persistence/plugin"

	// Postgresql driver
	_ "github.com/lib/pq"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitPostgresql() {
	m := global.Config.Postgresql

	dsn := "host=%s user=%s password=%s dbname=%s port=%v sslmode=%s TimeZone=%s"
	var s = fmt.Sprintf(dsn, m.Host, m.Username, m.Password, m.DbName, m.Port, m.SslMode, m.TimeZone)

	db, err := sql.Open("postgres", s)
	if err != nil {
		global.Logger.Error("Failed to connect to database", zap.Error(err))
		panic(err)
	}

	global.Logger.Info("Postgresql init success")
	global.Pdbc = db

	// Initialize GORM
	gormDB, err := gorm.Open(postgres.Open(s), &gorm.Config{})
	if err != nil {
		global.Logger.Error("Failed to initialize GORM", zap.Error(err))
		panic(err)
	}

	// Register RLS Plugin
	if err := gormDB.Use(plugin.NewRLSPlugin()); err != nil {
		global.Logger.Error("Failed to register RLS plugin", zap.Error(err))
		panic(err)
	}

	global.GormDB = gormDB
	global.Logger.Info("GORM init success")

	// set Pool
	SetPool()
}

func SetPool() {
	global.Pdbc.SetMaxOpenConns(global.Config.Postgresql.MaxConn)
	global.Pdbc.SetConnMaxLifetime(time.Duration(global.Config.Postgresql.MaxLifeTime))
	global.Pdbc.SetConnMaxIdleTime(time.Duration(global.Config.Postgresql.IdleTimeOut))
}
