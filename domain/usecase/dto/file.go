package dto

import "io"

type FileContentRequest struct {
	Name    string
	Width   int
	Height  int
	Format  string
	Quality int
}

type UploadFileRequest struct {
	Name        string
	ContentType string
	Reader      io.Reader
}

type CachedFileResult struct {
	Data        []byte `json:"data"`
	ContentType string `json:"content_type"`
}
