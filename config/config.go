package config

import (
	"github.com/urfave/cli/v3"
)

type format = string

const (
	FORMAT_TABLE format = "table"
	FORMAT_CSV   format = "csv"
	FORMAT_JSON  format = "json"
	FORMAT_NONE  format = "none"
)

type Config struct {
	// Azure credentials and settings
	Azure AzureConfig

	// Report and filtering settings
	ExpiringWithinDays int
	IncludeExpired     bool
	GroupSources       bool
	OutputFile         string
	Format             format
	WebhookURL         string
}

type AzureConfig struct {
	TenantID     string
	ClientID     string
	ClientSecret string
}

func FromCmd(cmd *cli.Command) (*Config, error) {
	// Azure credentials and settings
	cfgAzure := AzureConfig{
		TenantID:     cmd.String("az-tenant-id"),
		ClientID:     cmd.String("az-client-id"),
		ClientSecret: cmd.String("az-client-secret"),
	}

	// Build the main config from command arguments
	cfg := &Config{
		Azure:              cfgAzure,
		ExpiringWithinDays: cmd.Int("expiring-within-days"),
		IncludeExpired:     cmd.Bool("include-expired"),
		GroupSources:       cmd.Bool("group-sources"),
		OutputFile:         cmd.String("output-file"),
		Format:             cmd.String("format"),
		WebhookURL:         cmd.String("webhook-url"),
	}

	return cfg, nil
}
