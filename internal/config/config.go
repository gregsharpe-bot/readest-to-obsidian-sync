package config

import (
	"fmt"
	"os"
)

type Config struct {
	AWSRegion     string
	QueueURL      string
	S3Bucket      string
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
	cfg.ObsidianVault = os.Getenv("OBSIDIAN_VAULT")
	if cfg.ObsidianVault == "" {
		cfg.ObsidianVault = "/var/lib/obsidian-vault"
	}
	cfg.NotesFolder = os.Getenv("OBSIDIAN_NOTES_FOLDER")
	if cfg.NotesFolder == "" {
		cfg.NotesFolder = "07 - Inbox"
	}
	return cfg, nil
}
