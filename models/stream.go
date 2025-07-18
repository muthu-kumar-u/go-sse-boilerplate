package appschema

type EventMessage struct {
	Code           int       `json:"code,omitempty"`
	Data           any       `json:"data,omitempty"`
	Event          string    `json:"event,omitempty"`
	Message        string    `json:"message,omitempty"`
	StreamID       string    `json:"stream_id,omitempty"`
	Completion     int       `json:"stream_completion,omitempty"`
}