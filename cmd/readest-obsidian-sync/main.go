package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/gregsharpe-bot/readest-to-obsidian-sync/internal/config"
	workersqs "github.com/gregsharpe-bot/readest-to-obsidian-sync/internal/sqs"
)

func main() {
	if len(os.Args) != 2 || os.Args[1] != "run" {
		_, _ = os.Stderr.WriteString("usage: readest-obsidian-sync run\n")
		os.Exit(2)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.FromEnvironment()
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(), awsconfig.WithRegion(cfg.AWSRegion))
	if err != nil {
		logger.Error("load AWS configuration", "error", err)
		os.Exit(1)
	}
	worker := workersqs.Worker{
		Client:   awssqs.NewFromConfig(awsCfg),
		QueueURL: cfg.QueueURL,
		Bucket:   cfg.S3Bucket,
		Logger:   logger,
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := worker.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("worker stopped with error", "error", err)
		os.Exit(1)
	}
}
