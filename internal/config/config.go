package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

const (
	DefaultMaxConn  = 10
	DefaultMinConn  = 1
	MaxAllowedConns = 100
)

type Config struct {
	HTTPAddr        string        `mapstructure:"HTTP_ADDR"`
	GRPCAddr        string        `mapstructure:"GRPC_ADDR"`
	PrometheusAddr  string        `mapstructure:"PROMETHEUS_ADDR"`
	ShutdownTimeout time.Duration `mapstructure:"SHUTDOWN_TIMEOUT"`

	DatabaseURL string `mapstructure:"DB_URL"`
	DBMaxConn   int    `mapstructure:"DB_MAX_CONN"`
	DBMinConn   int    `mapstructure:"DB_MIN_CONN"`

	JWTSecret string        `mapstructure:"JWT_SECRET"`
	TokenTTL  time.Duration `mapstructure:"TOKEN_TTL"`

	LogLevel string `mapstructure:"LOG_LEVEL"`
}

func LoadConfig() *Config {
	setDefaults()

	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Printf("Ошибка при чтении файла конфигурации: %v\n", err)
		}

		log.Println("Используются значения по умолчанию или из переменных окружения")
	}

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		log.Printf("Ошибка при разборе конфигурации: %v\n", err)
		return getDefaultConfig()
	}

	// Ограничения для DBMaxConn и DBMinConn
	if config.DBMaxConn <= 0 || config.DBMaxConn > MaxAllowedConns {
		log.Printf("Некорректное значение DBMaxConn (%d), используется значение по умолчанию: %d\n", config.DBMaxConn, DefaultMaxConn)
		config.DBMaxConn = DefaultMaxConn
	}

	if config.DBMinConn < 0 || config.DBMinConn > MaxAllowedConns {
		log.Printf("Некорректное значение DBMinConn (%d), используется значение по умолчанию: %d\n", config.DBMinConn, DefaultMinConn)
		config.DBMinConn = DefaultMinConn
	}

	return config
}

func setDefaults() {
	viper.SetDefault("HTTP_ADDR", ":8080")
	viper.SetDefault("GRPC_ADDR", ":3000")
	viper.SetDefault("PROMETHEUS_ADDR", ":9000")
	viper.SetDefault("SHUTDOWN_TIMEOUT", "5s")

	viper.SetDefault("DB_URL", "postgres://postgres:postgres@localhost:5432/avito?sslmode=disable")
	viper.SetDefault("DB_MAX_CONN", 10)
	viper.SetDefault("DB_MIN_CONN", 5)

	viper.SetDefault("JWT_SECRET", "supersecretkey")
	viper.SetDefault("TOKEN_TTL", "24h")

	viper.SetDefault("LOG_LEVEL", "info")
}

func getDefaultConfig() *Config {
	return &Config{
		HTTPAddr:        ":8080",
		GRPCAddr:        ":3000",
		PrometheusAddr:  ":9000",
		ShutdownTimeout: 5 * time.Second,

		DatabaseURL: "postgres://postgres:postgres@localhost:5432/avito?sslmode=disable",
		DBMaxConn:   10,
		DBMinConn:   5,

		JWTSecret: "supersecretkey",
		TokenTTL:  24 * time.Hour,

		LogLevel: "info",
	}
}
