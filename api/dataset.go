package api

import "time"

type Dataset struct {
	ID             string `json:"id,omitempty"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	RetentionDays  int    `json:"retentionDays"`
	OrganizationID string `json:"organizationId"`
}

type CreateDatasetRequest struct {
	Name          string `json:"name" validate:"required,min=1,max=255"`
	Description   string `json:"description" validate:"max=255"`
	RetentionDays int    `json:"retentionDays" validate:"required,min=1,max=365"`
}

type DatasetField struct {
	Field string `json:"field"`
	Type  string `json:"type"`
}

type DatasetLog struct {
	Timestamp  time.Time         `json:"Timestamp"`
	DatasetId  string            `json:"DatasetId"`
	TraceId    string            `json:"TraceId"`
	TeamId     string            `json:"TeamId"`
	Body       string            `json:"Body"`
	Attributes map[string]string `json:"Metadata"`
}

type DatasetLogsResponse struct {
	Data []DatasetLog `json:"data"`
	Meta struct {
		TotalRowCount int `json:"totalRowCount"`
	} `json:"meta"`
}

type ListLogsSchema struct {
	OrganizationID string `json:"organizationId"`
	DatasetID      string `json:"datasetId"`
	Levels         string `json:"levels"`
	FromDate       string `json:"fromDate"`
	ToDate         string `json:"toDate"`
	Metadata       string `json:"metadata"`
	Filter         string `json:"filter"`
	Cursor         int    `json:"cursor"`
}

type DatasetFilter struct {
	Field    string          `json:"field"`
	Operator string          `json:"operator"`
	Value    *string         `json:"value,omitzero"`
	Type     *string         `json:"type,omitzero"`
	Children []DatasetFilter `json:"children,omitzero"`
}
