package events

import "time"

type Event struct {
	Name      string
	Timestamp time.Time
	Payload   interface{}
}
