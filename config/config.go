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
	Cloudflare *Cloudflare `yaml:"cloudflare" validate:"required"`
	Issuer     string      `yaml:"issuer" validate:"required,http_url"`
}

type Cloudflare struct {
	ApiToken string `yaml:"apiToken" validate:"required,len=40"`
}

func LoadConfiguration() (config *YamlConfig, err error) {
	configPath := os.Getenv("GENERAL_CONFIG_PATH")
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
		return nil, err
	}

	return &yamlConfig, nil
}

func ValidateConfig(config interface{}) (err error) {
	validate := validator.New()
	err = validate.Struct(config)
	if err != nil {
		return fmt.Errorf("error validating config file: %s", err)
	}
	return nil
}
