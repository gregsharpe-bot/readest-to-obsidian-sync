package sqs

import (
	"context"
	"log/slog"
	"testing"

	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type fakeClient struct {
	receives []*awssqs.ReceiveMessageOutput
	deleted  int
}

func (f *fakeClient) ReceiveMessage(context.Context, *awssqs.ReceiveMessageInput, ...func(*awssqs.Options)) (*awssqs.ReceiveMessageOutput, error) {
	output := f.receives[0]
	f.receives = f.receives[1:]
	return output, nil
}

func (f *fakeClient) DeleteMessage(context.Context, *awssqs.DeleteMessageInput, ...func(*awssqs.Options)) (*awssqs.DeleteMessageOutput, error) {
	f.deleted++
	return &awssqs.DeleteMessageOutput{}, nil
}

func TestWorkerDrainsMessagesAndExits(t *testing.T) {
	body := `{"Records":[{"eventSource":"aws:s3","eventName":"ObjectCreated:Put","eventTime":"2026-01-01T00:00:00Z","s3":{"bucket":{"name":"books"},"object":{"key":"book.epub"}}}]}`
	receipt := "receipt"
	client := &fakeClient{receives: []*awssqs.ReceiveMessageOutput{
		{Messages: []types.Message{{Body: &body, ReceiptHandle: &receipt}}},
		{},
	}}
	worker := Worker{Client: client, QueueURL: "queue", Logger: slog.New(slog.NewJSONHandler(testWriter{t}, nil))}
	if err := worker.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	if client.deleted != 1 {
		t.Fatalf("deleted %d messages, want 1", client.deleted)
	}
}

func TestWorkerDoesNotDeleteMalformedMessage(t *testing.T) {
	body := `{not-json}`
	receipt := "receipt"
	client := &fakeClient{receives: []*awssqs.ReceiveMessageOutput{{Messages: []types.Message{{Body: &body, ReceiptHandle: &receipt}}}, {}}}
	worker := Worker{Client: client, QueueURL: "queue", Logger: slog.New(slog.NewJSONHandler(testWriter{t}, nil))}
	if err := worker.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	if client.deleted != 0 {
		t.Fatal("malformed message was deleted")
	}
}

type testWriter struct{ t *testing.T }

func (w testWriter) Write(p []byte) (int, error) {
	w.t.Log(string(p))
	return len(p), nil
}
