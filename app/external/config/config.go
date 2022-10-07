package config

import (
	"github.com/go-yaml/yaml"
	"os"
)

type Config struct {
	Store struct {
		Hydra struct {
			IPAddress string `yaml:"ip_address"`
			Username  string `yaml:"username"`
			Password  string `yaml:"password"`
			Service   string `yaml:"service"`
		} `yaml:"hydra"`
		Local struct {
			IPAddress string `yaml:"ip_address"`
			Username  string `yaml:"username"`
			Password  string `yaml:"password"`
			Port      string `json:"port"`
			Database  string `json:"database"`
			Driver    string `json:"driver"`
		}
	} `yaml:"store"`
	HydraBlockSubscritionId string `yaml:"hydra_block_subscription_id"`
	HydraHashSumSalt        string `yaml:"hydra_hash_sum_salt"`
	Server                  struct {
		IPAddress string `yaml:"ip_address"`
		Port      string `yaml:"port"`
	} `yaml:"server"`
}

func NewConfig(configPath string) (*Config, error) {
	config := &Config{}

	file, err := os.Open(configPath)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	d := yaml.NewDecoder(file)

	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}
