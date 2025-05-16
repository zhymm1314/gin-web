package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type RabbitMQ struct {
	Host     string `mapstructure:"host" json:"host" yaml:"host"`
	Port     int    `mapstructure:"port" json:"port" yaml:"port"`
	Username string `mapstructure:"username" json:"username" yaml:"username"`
	Password string `mapstructure:"password" json:"password" yaml:"password"`
	Vhost    string `mapstructure:"vhost" json:"vhost" yaml:"vhost"`
}

type RabbitMQConfig struct {
	Host              string `yaml:"host"`
	Port              int    `yaml:"port"`
	Username          string `yaml:"username"`
	Password          string `yaml:"password"`
	ReconnectInterval int    `yaml:"reconnect_interval"`
	Vhost             string `yaml:"vhost"`
}

type ConsumerConfig struct {
	Queue       string `yaml:"queue"`
	Concurrency int    `yaml:"concurrency"`
	Handler     string `yaml:"handler"`
}

type AppConfig struct {
	RabbitMQ  RabbitMQConfig   `yaml:"rabbitmq"`
	Consumers []ConsumerConfig `yaml:"consumers"`
}

func LoadConfig(path string) (*AppConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
