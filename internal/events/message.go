package events

import "time"

type Message struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Source    string    `json:"soruce"`
	Timestamp time.Time `json:"timestamp"`
	Payload   any       `json:"payload"`
}
