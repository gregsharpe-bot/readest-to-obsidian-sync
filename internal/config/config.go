package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	AWSRegion     string
	QueueURL      string
	S3Bucket      string
	S3Prefix      string
	ObsidianVault string
	NotesFolder   string
}

func FromEnvironment() (Config, error) {
	values := []string{"AWS_REGION", "SQS_QUEUE_URL", "S3_BUCKET"}

	var cfg Config
	missing := make([]string, 0, len(values))
	for _, name := range values {
		value := os.Getenv(name)
		if value == "" {
			missing = append(missing, name)
		}
		switch name {
		case "AWS_REGION":
			cfg.AWSRegion = value
		case "SQS_QUEUE_URL":
			cfg.QueueURL = value
		case "S3_BUCKET":
			cfg.S3Bucket = value
		}
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required environment variables: %v", missing)
	}
	cfg.S3Prefix = strings.Trim(os.Getenv("S3_PREFIX"), "/")
	cfg.ObsidianVault = os.Getenv("OBSIDIAN_VAULT")
	if cfg.ObsidianVault == "" {
		cfg.ObsidianVault = "/var/lib/obsidian-vault"
	}
	cfg.NotesFolder = os.Getenv("OBSIDIAN_NOTES_FOLDER")
	if cfg.NotesFolder == "" {
		cfg.NotesFolder = "Readest"
	}
	return cfg, nil
}
