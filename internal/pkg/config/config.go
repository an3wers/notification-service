package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env      string         `yaml:"env" env:"ENV" env-default:"local"`
	Server   ServerConfig   `yaml:"server_config"`
	Database DatabaseConfig `yaml:"database_config"`
	// RabbitMQ RabbitMQConfig
	SMTP    SMTPConfig    `yaml:"smtp_config"`
	Storage StorageConfig `yaml:"storage_config"`
	Logger  LoggerConfig  `yaml:"logger_config"`
}

type ServerConfig struct {
	Port            string `yaml:"port" env:"PORT" env-default:"3020"`
	Host            string `yaml:"host" env:"HOST" env-default:"localhost"`
	ReadTimeout     int    `yaml:"read_timeout" env:"READ_TIMEOUT" env-default:"20"`
	WriteTimeout    int    `yaml:"write_timeout" env:"WRITE_TIMEOUT" env-default:"20"`
	ShutdownTimeout int    `yaml:"shutdown_timeout" env:"SHUTDOWN_TIMEOUT" env-default:"10"`
	IdleTimeout     int    `yaml:"idle_timeout" env:"IDLE_TIMEOUT" env-default:"60"`
}

type DatabaseConfig struct {
	Host            string `env:"DATABASE_HOST" env-default:"localhost"`
	Port            int    `env:"DATABASE_PORT" env-default:"5432"`
	User            string `env:"DATABASE_USER" env-default:"postgres"`
	Password        string `env:"DATABASE_PASSWORD" env-default:"postgres"`
	DBName          string `env:"DATABASE_NAME" env-default:"postgres"`
	SSLMode         string `yaml:"sslmode" env:"DATABASE_SSLMODE" env-default:"disable"`
	MaxOpenConns    int    `yaml:"max_open_conns" env:"DATABASE_MAX_OPEN_CONNS" env-default:"10"`
	MaxIdleConns    int    `yaml:"max_idle_conns" env:"DATABASE_MAX_IDLE_CONNS" env-default:"5"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime" env:"DATABASE_CONN_MAX_LIFETIME" env-default:"5"`
}

type SMTPConfig struct {
	Host     string `env:"SMTP_HOST" env-default:"localhost"`
	Port     int    `env:"SMTP_PORT" env-default:"25"`
	Username string `env:"SMTP_USER" env-default:""`
	Password string `env:"SMTP_PASSWORD" env-default:""`
	From     string `env:"SMTP_FROM" env-default:""`
	TLS      bool   `env:"SMTP_SECURE" env-default:"false"`
	Timeout  int    `env:"SMTP_TIMEOUT" env-default:"5"`
}

type StorageConfig struct {
	Provider    string `yaml:"provider" env-default:"local"`
	LocalPath   string `yaml:"local_path" env-default:"./uploads"`
	S3Bucket    string `yaml:"s3_bucket" env-default:""`
	S3Region    string `yaml:"s3_region" env-default:""`
	S3Endpoint  string `yaml:"s3_endpoint" env-default:""`
	MaxFileSize int64  `yaml:"max_file_size" env-default:"62914560"`
}

type LoggerConfig struct {
	Level      string `yaml:"level" env-default:"info"`
	Format     string `yaml:"format" env-default:"console"`
	OutputPath string `yaml:"output_path" env-default:"./logs"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
