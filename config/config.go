package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Configuration struct {
	App       AppConfig
	DB        DBConfig
	JWT       JWTConfig
	SMTP      SMTPConfig
	Storage   StorageConfig
	Scheduler SchedulerConfig
	Seeder    SeederConfig
}

type AppConfig struct {
	Port string
	Env  string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type JWTConfig struct {
	Secret      string
	ExpireHours int
}

type SMTPConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

type StorageConfig struct {
	UploadDir     string
	MaxFileSizeMB int64
}

type SchedulerConfig struct {
	SLACheckIntervalMinutes int
}

type SeederConfig struct {
	AdminUsername string
	AdminPassword string
	AdminName     string
	AdminEmail    string
}

var Cfg *Configuration

func Load() {
	if err := godotenv.Load(); err != nil {
		log.Println("[CONFIG] No .env file found, reading from environment variables")
	} else {
		log.Println("[CONFIG] .env loaded successfully")
	}

	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	jwtExpire, _ := strconv.Atoi(getEnv("JWT_EXPIRE_HOURS", "24"))
	maxFile, _ := strconv.ParseInt(getEnv("MAX_FILE_SIZE_MB", "10"), 10, 64)
	slaInterval, _ := strconv.Atoi(getEnv("SLA_CHECK_INTERVAL_MINUTES", "30"))

	Cfg = &Configuration{
		App: AppConfig{
			Port: getEnv("APP_PORT", "8080"),
			Env:  getEnv("APP_ENV", "development"),
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "1433"),
			User:     getEnv("DB_USER", "sa"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "sms_db"),
		},
		JWT: JWTConfig{
			Secret:      getEnv("JWT_SECRET", "changeme"),
			ExpireHours: jwtExpire,
		},
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", ""),
			Port:     smtpPort,
			User:     getEnv("SMTP_USER", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", ""),
		},
		Storage: StorageConfig{
			UploadDir:     getEnv("UPLOAD_DIR", "./uploads"),
			MaxFileSizeMB: maxFile,
		},
		Scheduler: SchedulerConfig{
			SLACheckIntervalMinutes: slaInterval,
		},
		Seeder: SeederConfig{
			AdminUsername: getEnv("SEED_ADMIN_USERNAME", "admin"),
			AdminPassword: getEnv("SEED_ADMIN_PASSWORD", "Admin@123"),
			AdminName:     getEnv("SEED_ADMIN_NAME", "Administrator"),
			AdminEmail:    getEnv("SEED_ADMIN_EMAIL", "admin@yusen.com"),
		},
	}

	log.Printf("[CONFIG] App running on port %s | env: %s", Cfg.App.Port, Cfg.App.Env)
}

func (d *DBConfig) DSN() string {
	return fmt.Sprintf(
		"sqlserver://%s:%s@%s:%s?database=%s",
		d.User, d.Password, d.Host, d.Port, d.Name,
	)
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
