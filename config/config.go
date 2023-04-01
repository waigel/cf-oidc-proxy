package config

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
	"os"
)

type YamlConfig struct {
	OidcProxy OidcProxy `yaml:"oidc-proxy" validate:"required"`
}

type OidcProxy struct {
	Cloudflare *Cloudflare `yaml:"cloudflare"`
	Issuer     string      `yaml:"issuer" validate:"required,http_url"`
}

type Cloudflare struct {
	ApiToken *string `yaml:"apiToken"`
}

func LoadConfiguration() (config *YamlConfig, err error) {
	configPath := os.Getenv("GENERAL_CONFIG_PATH")
	fmt.Println("Using config file: " + configPath)
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %s", err)
	}

	yamlConfig := YamlConfig{}
	err = yaml.Unmarshal(configFile, &yamlConfig)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %s", err)
	}
	if err := ValidateConfig(&yamlConfig); err != nil {
		return nil, fmt.Errorf("error validating config file: %s", err)
	}
	return &yamlConfig, nil
}

func ValidateConfig(config interface{}) error {
	validate := validator.New()
	err := validate.Struct(config)
	if err != nil {
		return fmt.Errorf("error validating config file: %s", err)
	}
	return nil
}
