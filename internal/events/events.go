package events

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type Notification struct {
	Records []Record `json:"Records"`
}

type Record struct {
	EventSource string   `json:"eventSource"`
	EventName   string   `json:"eventName"`
	EventTime   string   `json:"eventTime"`
	S3          S3Record `json:"s3"`
}

type S3Record struct {
	Bucket ObjectBucket `json:"bucket"`
	Object Object       `json:"object"`
}

type ObjectBucket struct {
	Name string `json:"name"`
}

type Object struct {
	Key  string `json:"key"`
	Size *int64 `json:"size,omitempty"`
}

func Parse(data []byte) (Notification, error) {
	var notification Notification
	if err := json.Unmarshal(data, &notification); err != nil {
		return Notification{}, fmt.Errorf("decode S3 notification: %w", err)
	}
	if len(notification.Records) == 0 {
		return Notification{}, fmt.Errorf("S3 notification contains no records")
	}
	for index := range notification.Records {
		record := &notification.Records[index]
		if record.EventName == "" || record.EventSource != "aws:s3" {
			return Notification{}, fmt.Errorf("record %d is not an S3 event", index)
		}
		if record.S3.Bucket.Name == "" || record.S3.Object.Key == "" {
			return Notification{}, fmt.Errorf("record %d is missing bucket or object key", index)
		}
		key, err := url.QueryUnescape(record.S3.Object.Key)
		if err != nil {
			return Notification{}, fmt.Errorf("decode object key in record %d: %w", index, err)
		}
		record.S3.Object.Key = key
	}
	return notification, nil
}
