package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

type Config struct {
	AppURL  string         `yaml:"app_url"`
	Company  CompanyConfig  `yaml:"company"`
	SnipeIT  SnipeITConfig  `yaml:"snipe_it"`
	SMTP     SMTPConfig     `yaml:"smtp"`
	PO       POConfig       `yaml:"po"`
}

type CompanyConfig struct {
	Name         string `yaml:"name"`
	Address      string `yaml:"address"`
	CityStateZip string `yaml:"city_state_zip"`
	Phone        string `yaml:"phone"`
	Email        string `yaml:"email"`
}

type SnipeITConfig struct {
	URL    string `yaml:"url"`
	APIKey string `yaml:"api_key"`
}

type SMTPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
}

type POConfig struct {
	NextNumber        int    `yaml:"next_number"`
	NumberPrefix      string `yaml:"number_prefix"`
	DefaultTerms      string `yaml:"default_terms"`
	DefaultPaymentType string `yaml:"default_payment_type"`
}

var (
	cfg  *Config
	once sync.Once
)

func Load() (*Config, error) {
	var loadErr error
	once.Do(func() {
		cfg = &Config{}
		loadErr = loadConfig()
	})
	if loadErr != nil {
		return nil, loadErr
	}
	return cfg, nil
}

func loadConfig() error {
	configPaths := []string{}

	exePath, err := os.Executable()
	if err == nil {
		configPaths = append(configPaths, filepath.Dir(exePath))
	}

	cwd, err := os.Getwd()
	if err == nil {
		configPaths = append(configPaths, cwd)
	}

	configPaths = append(configPaths, "/etc/snipe-po")

	for _, path := range configPaths {
		configFile := filepath.Join(path, "settings.yaml")
		if _, err := os.Stat(configFile); err == nil {
			return readConfigFile(configFile)
		}
	}

	for _, path := range configPaths {
		configFile := filepath.Join(path, "settings.example.yaml")
		if _, err := os.Stat(configFile); err == nil {
			return readConfigFile(configFile)
		}
	}

	return fmt.Errorf("no config file found in: %v", configPaths)
}

func readConfigFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, cfg)
}

func Get() *Config {
	return cfg
}
