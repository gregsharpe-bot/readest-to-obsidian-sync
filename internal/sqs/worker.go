package sqs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"

	"github.com/gregsharpe-bot/readest-to-obsidian-sync/internal/events"
)

type Client interface {
	ReceiveMessage(context.Context, *awssqs.ReceiveMessageInput, ...func(*awssqs.Options)) (*awssqs.ReceiveMessageOutput, error)
	DeleteMessage(context.Context, *awssqs.DeleteMessageInput, ...func(*awssqs.Options)) (*awssqs.DeleteMessageOutput, error)
}

type Processor interface {
	Process(context.Context, events.Record) error
}

type Worker struct {
	Client              Client
	QueueURL            string
	Bucket              string
	Logger              *slog.Logger
	Processor           Processor
	AcknowledgeFailures bool
}

func (w Worker) Run(ctx context.Context) error {
	for {
		output, err := w.Client.ReceiveMessage(ctx, &awssqs.ReceiveMessageInput{
			QueueUrl:            &w.QueueURL,
			MaxNumberOfMessages: 10,
			WaitTimeSeconds:     20,
			VisibilityTimeout:   300,
		})
		if err != nil {
			return err
		}
		if len(output.Messages) == 0 {
			return nil
		}
		for _, message := range output.Messages {
			if err := w.handleMessage(ctx, message); err != nil {
				w.Logger.Warn("message was not handled", "error", err)
				continue
			}
		}
	}
}

func (w Worker) handleMessage(ctx context.Context, message types.Message) error {
	if message.Body == nil || message.ReceiptHandle == nil {
		return fmt.Errorf("message is missing body or receipt handle")
	}
	notification, err := events.Parse([]byte(*message.Body))
	if err != nil {
		return err
	}
	var processingErr error
	for _, record := range notification.Records {
		if w.Bucket != "" && record.S3.Bucket.Name != w.Bucket {
			return fmt.Errorf("event bucket %q does not match configured bucket", record.S3.Bucket.Name)
		}
		fields := []any{
			"message_id", value(message.MessageId),
			"event_name", record.EventName,
			"event_time", record.EventTime,
			"s3_bucket", record.S3.Bucket.Name,
			"s3_object_key", record.S3.Object.Key,
		}
		if record.S3.Object.Size != nil {
			fields = append(fields, "s3_object_size", *record.S3.Object.Size)
		}
		w.Logger.Info("received S3 event", fields...)
		if w.Processor != nil {
			if err := w.Processor.Process(ctx, record); err != nil {
				processingErr = errors.Join(processingErr, err)
			}
		}
	}
	if processingErr != nil && !w.AcknowledgeFailures {
		return processingErr
	}
	_, err = w.Client.DeleteMessage(ctx, &awssqs.DeleteMessageInput{
		QueueUrl:      &w.QueueURL,
		ReceiptHandle: message.ReceiptHandle,
	})
	return errors.Join(processingErr, err)
}

func value(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
