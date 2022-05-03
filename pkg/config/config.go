package config

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Host     string `yaml:"host" envconfig:"GUEST_IDP_HOST"`
	Port     string `yaml:"port" envconfig:"GUEST_IDP_PORT"`
	AdminURL string `yaml:"adminURL" envconfig:"HYDRA_ADMIN_URL"`
}

func (x *Config) Init() {
	x.Host = "localhost"
	x.Port = "4000"
	x.AdminURL = "http://localhost:4445"
}

func defaultConfig() *Config {
	var cfg Config
	cfg.Init()
	return &cfg
}

func readFile(cfg *Config) error {
	configFileName, ok := os.LookupEnv("CONFIG_FILE")
	if !ok {
		configFileName = "config.yaml"
	}

	f, err := os.Open(configFileName)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("No YAML config file %v", configFileName)
			return nil
		} else {
			return fmt.Errorf("Error reading config file %v: %w", configFileName, err)
		}
	}
	defer f.Close()
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		if err != io.EOF {
			return fmt.Errorf("Error parsing YAML: %w", err)
		}
	}
	return nil
}

func readEnv(cfg *Config) error {
	err := envconfig.Process("", cfg)
	return err
}

func GetConfig() (*Config, error) {
	cfg := defaultConfig()

	if err := readFile(cfg); err != nil {
		return cfg, err
	}
	if err := readEnv(cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
