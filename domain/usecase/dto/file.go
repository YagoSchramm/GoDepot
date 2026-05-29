package dto

type FileContentRequest struct {
	Name    string
	Width   int
	Height  int
	Format  string
	Quality int
}

type CachedFileResult struct {
	Data        []byte `json:"data"`
	ContentType string `json:"content_type"`
}
