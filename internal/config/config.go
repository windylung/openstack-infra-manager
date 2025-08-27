package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AuthURL        string // OS_AUTH_URL
	Username       string // OS_USERNAME
	Password       string // OS_PASSWORD
	UserDomainID   string // OS_USER_DOMAIN_ID
	ProjectName    string // OS_PROJECT_NAME
	RegionName     string // OS_REGION_NAME
	AdminProjectID string // OS_ADMIN_PROJECT_ID
}

func Load() (*Config, error) {
	_ = godotenv.Load(".env")

	c := &Config{
		AuthURL:        os.Getenv("OS_AUTH_URL"),
		Username:       os.Getenv("OS_USERNAME"),
		Password:       os.Getenv("OS_PASSWORD"),
		UserDomainID:   os.Getenv("OS_USER_DOMAIN_ID"),
		ProjectName:    os.Getenv("OS_PROJECT_NAME"),
		RegionName:     os.Getenv("OS_REGION_NAME"),
		AdminProjectID: os.Getenv("OS_ADMIN_PROJECT_ID"),
	}
	if c.AuthURL == "" || c.Username == "" || c.Password == "" ||
		c.UserDomainID == "" || c.ProjectName == "" || c.RegionName == "" {
		return nil, errors.New("missing one or more OpenStack envs: OS_AUTH_URL, OS_USERNAME, OS_PASSWORD, OS_USER_DOMAIN_ID, OS_PROJECT_NAME, OS_REGION_NAME, OS_ADMIN_PROJECT_ID")
	}
	return c, nil
}
