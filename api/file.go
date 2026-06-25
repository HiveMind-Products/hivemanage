package api

import (
	"mime/multipart"
	"time"
)

type UploadFile struct {
	FileHeader *multipart.FileHeader
	File       multipart.File
}

type Asset struct {
	ID           string    `json:"id"`
	Key          string    `json:"key"`
	OriginalName string    `json:"originalName"`
	Size         int64     `json:"size"`
	Type         string    `json:"type"`
	CreatedAt    time.Time `json:"createdAt"`
}

type AssetResponse struct {
	StorageFiles []*Asset `json:"files"`
	TotalCount   int      `json:"totalCount"`
}

type AssetURLResponse struct {
	URL string `json:"url"`
}
