package events

import (
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantKey string
		wantN   int
		wantErr bool
	}{
		{"valid event", `{"Records":[{"eventSource":"aws:s3","eventName":"ObjectCreated:Put","eventTime":"2026-01-01T00:00:00Z","s3":{"bucket":{"name":"books"},"object":{"key":"folder/book+one.epub","size":42}}}]}`, "folder/book one.epub", 1, false},
		{"multiple records", `{"Records":[{"eventSource":"aws:s3","eventName":"ObjectCreated:Put","s3":{"bucket":{"name":"books"},"object":{"key":"one"}}},{"eventSource":"aws:s3","eventName":"ObjectRemoved:Delete","s3":{"bucket":{"name":"books"},"object":{"key":"two"}}}]}`, "one", 2, false},
		{"malformed JSON", `{`, "", 0, true},
		{"missing fields", `{"Records":[{"eventName":"s3:ObjectCreated:Put","s3":{"bucket":{"name":"books"},"object":{}}}]}`, "", 0, true},
		{"unexpected message", `{"Type":"SubscriptionConfirmation"}`, "", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && (len(got.Records) != tt.wantN || got.Records[0].S3.Object.Key != tt.wantKey) {
				t.Fatalf("Parse() = %+v", got)
			}
		})
	}
}

func TestParseFixture(t *testing.T) {
	data, err := os.ReadFile("../../testdata/sqs/s3-object-created.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Parse(data); err != nil {
		t.Fatal(err)
	}
}

func TestParseDuplicateMessagesIsHarmless(t *testing.T) {
	input := []byte(`{"Records":[{"eventSource":"aws:s3","eventName":"ObjectCreated:Put","s3":{"bucket":{"name":"books"},"object":{"key":"same"}}}]}`)
	first, err := Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Parse(input)
	if err != nil || first.Records[0].S3.Object.Key != second.Records[0].S3.Object.Key {
		t.Fatalf("duplicate parse was not stable: %v", err)
	}
}
