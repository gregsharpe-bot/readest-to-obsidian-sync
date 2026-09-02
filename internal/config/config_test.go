package config

import "testing"

func TestFromEnvironment(t *testing.T) {
	t.Setenv("AWS_REGION", "eu-west-2")
	t.Setenv("SQS_QUEUE_URL", "https://example.invalid/queue")
	t.Setenv("S3_BUCKET", "books")
	got, err := FromEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	if got.AWSRegion != "eu-west-2" || got.QueueURL == "" || got.S3Bucket != "books" {
		t.Fatalf("unexpected config: %+v", got)
	}
}

func TestFromEnvironmentRequiresAllValues(t *testing.T) {
	t.Setenv("AWS_REGION", "")
	t.Setenv("SQS_QUEUE_URL", "")
	t.Setenv("S3_BUCKET", "")
	if _, err := FromEnvironment(); err == nil {
		t.Fatal("expected missing configuration error")
	}
}
