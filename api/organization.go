package api

type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CreateOrganizationRequest struct {
	Name string `json:"name" validate:"required,min=1,max=128"`
}

type UpdateOrganizationRequest struct {
	Name string `json:"name" validate:"required,min=1,max=128"`
}

type OrganizationStats struct {
	TotalFiles   int   `json:"totalFiles"`
	TotalSize    int64 `json:"totalSize"`
	TotalLogs    int   `json:"totalLogs"`
	TotalTokens  int   `json:"totalTokens"`
	TotalDataset int   `json:"totalDataset"`
}

// StorageCategory is an aggregate of stored files grouped by their top-level
// MIME category (image, video, audio, application, ...).
type StorageCategory struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
	Size     int64  `json:"size"`
}

// LargestFile is a single asset surfaced in the usage view.
type LargestFile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Size int64  `json:"size"`
}

// DatasetLogCount is the log volume for a single dataset.
type DatasetLogCount struct {
	DatasetID string `json:"datasetId"`
	Name      string `json:"name"`
	Count     int    `json:"count"`
}

// LogDayCount is the log volume for a single day (YYYY-MM-DD).
type LogDayCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// OrganizationUsage is the detailed usage breakdown surfaced on the Usage tab.
type OrganizationUsage struct {
	TotalFiles        int               `json:"totalFiles"`
	TotalSize         int64             `json:"totalSize"`
	StorageByCategory []StorageCategory `json:"storageByCategory"`
	LargestFiles      []LargestFile     `json:"largestFiles"`
	TotalLogs         int               `json:"totalLogs"`
	LogsByDataset     []DatasetLogCount `json:"logsByDataset"`
	LogsTimeseries    []LogDayCount     `json:"logsTimeseries"`
	LoggingAvailable  bool              `json:"loggingAvailable"`
}
