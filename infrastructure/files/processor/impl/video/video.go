package video

import (
	"encoding/json"
	"strings"

	"github.com/YagoSchramm/GoDepot/domain/entity"
	"github.com/YagoSchramm/GoDepot/domain/entity/derr"
	"github.com/YagoSchramm/GoDepot/infrastructure/files/processor"
)

type VideoProcessor struct{}

func NewVideoProcessor() processor.Processor {
	return &VideoProcessor{}
}

func (v *VideoProcessor) CanHandle(mimeType string) bool {
	return strings.HasPrefix(mimeType, "video/")
}

func (v *VideoProcessor) Process(file entity.File, opts entity.Options) (entity.Result, error) {
	meta := map[string]any{
		"name":      file.Name,
		"mime_type": file.MimeType,
		"size":      file.Size,
		"note":      "thumbnail extraction not supported yet",
	}

	data, err := json.Marshal(meta)
	if err != nil {
		return entity.Result{}, derr.JoinError("failed to marshal video metadata", err)
	}

	return entity.Result{
		Data:        data,
		ContentType: "application/json",
	}, nil
}
