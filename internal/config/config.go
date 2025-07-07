package config

import (
	"OrdersService/pkg/logging"
	"github.com/ilyakaznacheev/cleanenv"
	"sync"
)

type HTTPConfig struct {
	Port string `yaml:"port" env-default:"8080"`
}

type DBConfig struct {
	Host           string `yaml:"host" env-default:"localhost"`
	Port           string `yaml:"port" env-default:"5432"`
	Username       string `yaml:"username" env-default:"user"`
	Password       string `yaml:"password" env-default:"user"`
	Name           string `yaml:"name" env-default:"orders"`
	MaxConnections int    `yaml:"max_connections" env-default:"25"`
	MinConnections int    `yaml:"min_connections" env-default:"5"`
}

type KafkaConfig struct {
	Brokers  []string `yaml:"brokers" env-default:"localhost:9092"`
	Topic    string   `yaml:"topic" env-default:"orders"`
	GroupID  string   `yaml:"group_id" env-default:"orders-service-group"`
	MinBytes int      `yaml:"min_bytes" env-default:"10000"`
	MaxBytes int      `yaml:"max_bytes" env-default:"10000000"`
}
type Config struct {
	HTTP  HTTPConfig  `yaml:"http"`
	DB    DBConfig    `yaml:"database"`
	Kafka KafkaConfig `yaml:"kafka"`
}

var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		logger := logging.GetLogger()
		logger.Info("read application configuration")
		instance = &Config{}
		if err := cleanenv.ReadConfig("config.yaml", instance); err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			logger.Info(help)
			logger.Fatal(err)
		}
	})

	return instance
}
