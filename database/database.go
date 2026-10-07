package database

import (
	"log"

	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"sms-backend/config"
	"sms-backend/model"
)

var DB *gorm.DB

func Connect() {
	dsn := config.Cfg.DB.DSN()
	log.Printf("[DATABASE] Connecting to SQL Server at %s:%s database=%s",
		config.Cfg.DB.Host, config.Cfg.DB.Port, config.Cfg.DB.Name)

	var gormLogger logger.Interface
	if config.Cfg.App.Env == "development" {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Error)
	}

	db, err := gorm.Open(sqlserver.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		log.Fatalf("[DATABASE] Failed to connect: %v", err)
	}

	log.Println("[DATABASE] Connected successfully")
	DB = db

	autoMigrate()
}

func autoMigrate() {
	log.Println("[DATABASE] Running auto migration...")

	err := DB.AutoMigrate(
		&model.User{},
		&model.Location{},
		&model.Department{},
		&model.RiskReport{},
		&model.ReportFile{},
		&model.ReportActivity{},
		&model.ReportEditRequest{},
	)
	if err != nil {
		log.Fatalf("[DATABASE] Migration failed: %v", err)
	}

	log.Println("[DATABASE] Migration completed")
}
