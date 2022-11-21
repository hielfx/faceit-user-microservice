package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig
	Mongo  MongoConfig
	Redis  RedisConfig
}

type ServerConfig struct {
	Addr string
	Port int
}

type MongoConfig struct {
	URI string
	DB  string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// GetConfigFromFile - retrieves the config from the config file
func GetConfigFromFile(filepath string) (*Config, error) {
	v, err := LoadConfigFile(filepath)
	if err != nil {
		return nil, err
	}
	cfg, err := ParseConfig(v)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// LoadConfigFile - returns a new viper.Viper with the filepath configuration
func LoadConfigFile(filepath string) (*viper.Viper, error) {
	v := viper.New()

	v.SetConfigFile(filepath)
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		logrus.WithField("filepath", filepath).Errorf("Error in config.LoadConfigFile -> error: %s", err)
		return nil, err
	}

	return v, nil
}

// ParseConfig - unmarshal viper.Viper into Config struct
func ParseConfig(v *viper.Viper) (*Config, error) {
	var config Config

	if err := v.Unmarshal(&config); err != nil {
		logrus.Errorf("Error in config.ParseConfig -> error: %s", err)
		return nil, err
	}

	return &config, nil
}
