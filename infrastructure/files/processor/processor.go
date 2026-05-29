package processor

import "github.com/YagoSchramm/GoDepot/domain/entity"

type Processor interface {
	CanHandle(mimeType string) bool
	Process(file entity.File, opts entity.Options) (entity.Result, error)
}

type Registry struct {
	processors []Processor
}

func NewRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) Register(p Processor) {
	r.processors = append(r.processors, p)
}

func (r *Registry) Resolve(mimeType string) Processor {
	for _, p := range r.processors {
		if p.CanHandle(mimeType) {
			return p
		}
	}
	return nil
}
