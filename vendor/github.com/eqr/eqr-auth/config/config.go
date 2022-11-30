package config

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server struct {
		Host         string `yaml:"host"`
		InternalPort int    `yaml:"internalPort"`
	}
	Deploy struct {
		Host string `yaml:"host"`
	}
	JWT struct {
		Secret string `yaml:"secret"`
	}
}

// NewConfig creates Config structure from provided file
func NewConfig(configPath string) (*Config, error) {
	config := &Config{}

	// Open config file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}

// ValidateConfigPath just makes sure, that the path provided is a file,
// that can be read
func ValidateConfigPath(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		return err
	}
	if s.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a normal file", path)
	}
	return nil
}

func InitConfig(configPath string) *Config {
	log.Println("loading configuration")

	// Validate the path first
	if err := ValidateConfigPath(configPath); err != nil {
		log.Fatalf("cannot validate config path %s: %v", configPath, err.Error())
	}

	cfg, err := NewConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	return cfg
}
