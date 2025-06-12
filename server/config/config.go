package config

import (
	"fmt"
	"server/internal/logger"

	"github.com/spf13/viper"
)

type Config struct {
	GeneralVersion       string `mapstructure:"GENERAL_VERSION"`
	Environment          string `mapstructure:"ENVIRONMENT"`
	ServerPort           int    `mapstructure:"SERVER_PORT"`
	DatabaseDbPath       string `mapstructure:"DB_PATH"`
	DatabaseCacheAddress string `mapstructure:"DB_CACHE_ADDRESS"`
	DatabaseCachePort    int    `mapstructure:"DB_CACHE_PORT"`
	CorsAllowOrigins     string `mapstructure:"CORS_ALLOW_ORIGINS"`
	SecuritySalt         int    `mapstructure:"SECURITY_SALT"`
	SecurityPepper       string `mapstructure:"SECURITY_PEPPER"`
	SecurityJwtSecret    string `mapstructure:"SECURITY_JWT_SECRET"`
}

var ConfigInstance Config

func InitConfig() (Config, error) {
	log := logger.New("config").Function("InitConfig")
	log.Info("Initializing config")

	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	if err := viper.ReadInConfig(); err != nil {
		log.Warn("Could not find .env file", "error", err)
	}

	viper.AutomaticEnv()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return Config{}, log.Err("Fatal error: could not unmarshal config", err)
	}

	log.Info("Successfully initialized config", "config", config)
	return config, validateConfig(config, log)
}

func GetConfig() Config {
	return ConfigInstance
}

func validateConfig(config Config, log logger.Logger) error {
	if config.ServerPort <= 0 {
		return log.Err(
			"Fatal error: invalid server port",
			fmt.Errorf("invalid port: %d", config.ServerPort),
			"port", config.ServerPort,
		)
	}

	ConfigInstance = config
	return nil
}
