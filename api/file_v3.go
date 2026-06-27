package api

// FileItemV3 is the canonical V3 representation of a stored file.
type FileItemV3 struct {
	ID          string         `json:"id"`
	Filename    string         `json:"filename"`
	Type        string         `json:"type"`
	Size        int64          `json:"size"`
	URL         string         `json:"url"`
	OriginalURL string         `json:"originalUrl"`
	Metadata    map[string]any `json:"metadata"`
}

// UploadFileV3Response is the data payload returned by the V3 upload endpoints.
type UploadFileV3Response struct {
	ID          string `json:"id"`
	URL         string `json:"url"`
	OriginalURL string `json:"originalUrl"`
}

// PaginationV3 describes the pagination state of a V3 list response.
type PaginationV3 struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

// UploadFileBase64Request is the JSON body for POST /api/v3/file/base64.
type UploadFileBase64Request struct {
	Base64          string `json:"base64"`
	Filename        string `json:"filename"`
	Path            string `json:"path"`
	Metadata        string `json:"metadata"`
	RetentionExempt bool   `json:"retentionExempt"`
}

// LogV3 is a single log entry accepted by POST /api/v3/logs. The root body must
// be an array of these.
type LogV3 struct {
	Level    string         `json:"level"`
	Message  string         `json:"message"`
	Resource string         `json:"resource"`
	Metadata map[string]any `json:"metadata"`
	// Timestamp is accepted for forward-compatibility. The docs' OpenAPI fragment
	// does not define a timestamp field and the ingestion pipeline stamps server
	// time, so this is currently not persisted.
	// TODO: persist client timestamp once the docs define its format/semantics.
	Timestamp any `json:"timestamp"`
}
