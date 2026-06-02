package events

import (
	"encoding/json"
	"time"
)

const SchemaVersion = "1.0"

type Envelope struct {
	SchemaVersion string          `json:"schema_version"`
	EventType     string          `json:"event_type"`
	Payload       json.RawMessage `json:"payload"`
	Timestamp     time.Time       `json:"timestamp"`
}

func Wrap(eventType string, payload interface{}) ([]byte, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	env := Envelope{
		SchemaVersion: SchemaVersion,
		EventType:     eventType,
		Payload:       raw,
		Timestamp:     time.Now().UTC(),
	}
	return json.Marshal(env)
}

func Unwrap(data []byte, dest interface{}) (string, error) {
	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return "", err
	}
	return env.EventType, json.Unmarshal(env.Payload, dest)
}
