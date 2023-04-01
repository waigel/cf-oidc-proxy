package config

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
	"os"
)

type Permission struct {
	Name string `yaml:"name" validate:"required"`
	Id   string `yaml:"id" validate:"required,len=32"`
}

type Resource struct {
	Name  string `yaml:"name" validate:"required"`
	Value string `yaml:"value" validate:"required"`
}

type Policy struct {
	Name        string       `yaml:"name" validate:"required"`
	Effect      string       `yaml:"effect" validate:"required,oneof=allow deny reject Allow Deny Reject"`
	Permissions []Permission `yaml:"permissions" validate:"required,min=1"`
	Resources   []Resource   `yaml:"resources" validate:"required,min=1"`
}

type Condition struct {
	RequestIP struct {
		AllowOidcActor bool     `yaml:"allow-oidc-actor" validate:"eq=true|eq=false"`
		Whitelist      []string `yaml:"whitelist"`
		Blacklist      []string `yaml:"blacklist"`
	} `yaml:"request-ip" validate:"required"`
}
type Matchers struct {
	Operator string            `yaml:"operator" validate:"required,oneof=StringEquals StringNotEquals StringEqualsIgnoreCase StringNotEqualsIgnoreCase"`
	Claims   map[string]string `yaml:"claims" validate:"required,min=1"`
}
type Entities struct {
	Matchers []Matchers `yaml:"matchers" validate:"required,min=1"`
}

type Role struct {
	Name       string    `yaml:"name" validate:"required"`
	Conditions Condition `yaml:"conditions"`
	TTL        int       `yaml:"ttl" validate:"number"`
	Policies   []Policy  `yaml:"policies" validate:"required,min=1"`
	Entities   Entities  `yaml:"entities" validate:"required"`
}

type RoleConfig struct {
	Roles []Role `yaml:"groups" validate:"required"`
}

func (cfg *RoleConfig) GetRoleByName(name string) (foundGroup *Role, err error) {
	match := false
	group := &Role{}
	for _, groups := range cfg.Roles {
		if groups.Name == name {
			group = &groups
			match = true
			break
		}
	}
	if (group == &Role{} || group == nil) || !match {
		return nil, fmt.Errorf("role %s not found in config", name)
	}
	return group, nil
}

func LoadRoleConfiguration() (config *RoleConfig, err error) {
	configPath := os.Getenv("ROLES_CONFIG_PATH")
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %s", err)
	}

	roleConfig := RoleConfig{}
	err = yaml.Unmarshal(configFile, &roleConfig)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %s", err)
	}
	if err := validateRoleConfig(roleConfig); err != nil {
		return nil, err
	}
	return &roleConfig, nil
}

func validateRoleConfig(config RoleConfig) error {
	v := validator.New()

	if err := v.Struct(config); err != nil {
		return err
	}

	for _, group := range config.Roles {
		if err := v.Struct(group); err != nil {
			return err
		}

		for _, policy := range group.Policies {
			if err := v.Struct(policy); err != nil {
				return err
			}

			for _, permission := range policy.Permissions {
				if err := v.Struct(permission); err != nil {
					return err
				}
			}

			for _, resource := range policy.Resources {
				if err := v.Struct(resource); err != nil {
					return err
				}
			}
		}

		if err := v.Struct(group.Conditions); err != nil {
			return err
		}

		for _, whitelist := range group.Conditions.RequestIP.Whitelist {
			if err := v.Var(whitelist, "cidrv4"); err != nil {
				return err
			}
		}

		for _, blacklist := range group.Conditions.RequestIP.Blacklist {
			if err := v.Var(blacklist, "cidrv4"); err != nil {
				return err
			}
		}
	}
	return nil
}
