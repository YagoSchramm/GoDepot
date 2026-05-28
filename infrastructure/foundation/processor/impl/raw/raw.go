package raw

import (
	"os"

	"github.com/YagoSchramm/GoDepot/domain/entity"
	"github.com/YagoSchramm/GoDepot/domain/entity/derr"
	"github.com/YagoSchramm/GoDepot/infrastructure/foundation/processor"
)

type RawProcessor struct {
}

func (r *RawProcessor) CanHandle(_ string) bool {
	return true
}

func (r *RawProcessor) Process(file entity.File, opts entity.Options) (entity.Result, error) {
	data, err := os.ReadFile(file.Path)
	if err != nil {
		return entity.Result{}, derr.JoinError("raw: failed to read file: ", err)
	}
	return entity.Result{
		Data:        data,
		ContentType: file.MimeType,
	}, nil
}

func NewRawProcessor() processor.Processor {
	return &RawProcessor{}
}
