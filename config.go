package main

import (
	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type Pipeline struct {
  Name       string `validate:"required" mod:"trim"`
	Repository string `mod:"trim"`
	Script     string `validate:"required"`
}

type Config struct {
	Server struct {
		Host      string `validate:"required"`
		Port      int    `validate:"required"`
		JwtSecret string `yaml:"jwtSecret" validate:"required"`
		Workers   int    `validate:"required"`
    NotificationCommand string `yaml:"notificationCommand"`
	}
	Pipelines []Pipeline
}

type ConfigError struct {
	msg string
}

func (e *ConfigError) Error() string {
	return e.msg
}

func LoadConfig(path string) (*Config, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		log.Println("Failed to read config file:", err)
		return nil, err
	}
	config := Config{}
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		log.Println("Failed to parse config file:", err)
		return nil, err
	}
	validate := validator.New()
	err = validate.Struct(config)
	if err != nil {
		log.Println("Failed to validate config file:", err)
		return nil, err
	}
	if config.Server.Workers < 1 {
		log.Println("Invalid number of workers:", config.Server.Workers)
		return nil, &ConfigError{"Invalid number of workers"}
	}

	return &config, nil
}
