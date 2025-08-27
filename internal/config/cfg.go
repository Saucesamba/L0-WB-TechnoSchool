package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

const CONFIG_FILE = "./config.yaml"

type AppConfig struct {
	Kafka    KafkaConfig `yaml:"kafka"`
	Postgres DBConfig    `yaml:"postgres"`
	HTTP     HTTPConfig  `yaml:"http"`
}

type KafkaConfig struct {
	BrokerAddress string `yaml:"broker_address"`
	GroupID       string `yaml:"group_id"`
	Topic         string `yaml:"topic"`
}

type DBConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

type HTTPConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

func (a *AppConfig) LoadConfig() (*AppConfig, error) {
	cfg := &AppConfig{}
	file, err := os.ReadFile(CONFIG_FILE)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(file, cfg)
	return cfg, err
}
