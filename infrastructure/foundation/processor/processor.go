package processor

import "github.com/YagoSchramm/GoDepot/domain/entity"

type Processor interface {
	CanHandle(mimeType string) bool
	Process(file entity.File, opts entity.Options) ([]byte, error)
}
