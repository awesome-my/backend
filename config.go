package awesomemy

import (
	"fmt"
	"os"
	"strconv"

	"github.com/goccy/go-yaml"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
)

type Config struct {
	Postgres        PostgresConfig       `yaml:"postgres"`
	Redis           RedisConfig          `yaml:"redis"`
	Http            HttpConfig           `yaml:"http"`
	Authentication  AuthenticationConfig `yaml:"authentication"`
	FrontendBaseURL string               `yaml:"frontend_base_url"`
}

type PostgresConfig struct {
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
}

func (pc PostgresConfig) DSN() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable", pc.User, pc.Password, pc.Host, pc.Port, pc.Name)
}

type RedisConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

func (rc RedisConfig) ConnnectionString() string {
	return rc.Host + ":" + strconv.Itoa(rc.Port)
}

type HttpConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Cors struct {
		Origin []string `yaml:"origin"`
	} `yaml:"cors"`
}

func (hc HttpConfig) Address() string {
	return hc.Host + ":" + strconv.Itoa(hc.Port)
}

type AuthenticationConfig struct {
	Session struct {
		Prefix   string   `yaml:"prefix"`
		Name     string   `yaml:"name"`
		Persist  bool     `yaml:"persist"`
		SameSite string   `yaml:"same_site"`
		Secure   bool     `yaml:"secure"`
		Lifetime Duration `yaml:"lifetime"`
	} `yaml:"session"`
	OAuth2 AuthenticationOAuth2Config `yaml:"oauth2"`
}

type AuthenticationOAuth2Config struct {
	GitHub struct {
		ClientID     string `yaml:"client_id"`
		ClientSecret string `yaml:"client_secret"`
	} `yaml:"github"`
}

func (aac AuthenticationOAuth2Config) OAuth2Config(provider string) *oauth2.Config {
	switch provider {
	case "github":
		return &oauth2.Config{
			ClientID:     aac.GitHub.ClientID,
			ClientSecret: aac.GitHub.ClientSecret,
			Endpoint:     endpoints.GitHub,
			Scopes:       []string{"read:user", "user:email"},
		}
	}

	return nil
}

func ParseConfigFromFile(fp string) (Config, error) {
	b, err := os.ReadFile(fp)
	if err != nil {
		return Config{}, err
	}

	return ParseConfig(b)
}

func ParseConfig(b []byte) (Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
