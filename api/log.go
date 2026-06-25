package api

type Log struct {
	Level    string         `json:"level" validate:"required,max=32"`
	Message  string         `json:"message" validate:"required,max=8192"`
	Resource string         `json:"resource" validate:"max=255"`
	Metadata map[string]any `json:"metadata"`
}
