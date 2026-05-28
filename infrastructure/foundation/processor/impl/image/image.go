package image

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"strings"

	"github.com/YagoSchramm/GoDepot/domain/entity"
	"github.com/YagoSchramm/GoDepot/domain/entity/derr"
	"github.com/YagoSchramm/GoDepot/infrastructure/foundation/processor"
	"github.com/disintegration/imaging"
)

type ImageProcessor struct {
}

func NewImageProcessor() processor.Processor {
	return &ImageProcessor{}
}

func (p *ImageProcessor) CanHandle(mimeType string) bool {
	switch mimeType {
	case "image/jpeg", "image/png", "image/gif":
		return true
	default:
		return false
	}
}

func (p *ImageProcessor) Process(f entity.File, opts entity.Options) (entity.Result, error) {
	file, err := os.Open(f.Path)
	if err != nil {
		return entity.Result{}, derr.JoinError("Failed to open the image file: ", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return entity.Result{}, derr.JoinError("Failed to decode the image: ", err)
	}

	if opts.Width <= 0 || opts.Height <= 0 {
		return entity.Result{}, derr.JoinError("Failed to resize the image", err)
	}

	img = imaging.Resize(img, opts.Width, opts.Height, imaging.Lanczos)

	format := strings.ToLower(opts.Format)
	if format == "" {
		format = strings.TrimPrefix(f.MimeType, "image/")
	}

	var buf bytes.Buffer
	contentType := "image/jpeg"
	switch format {
	case "png":
		contentType = "image/png"
		err = png.Encode(&buf, img)
	default:
		quality := opts.Quality
		if quality == 0 {
			quality = 85
		}
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
	}

	if err != nil {
		return entity.Result{}, derr.JoinError("image: failed to encode: ", err)
	}

	return entity.Result{
		Data:        buf.Bytes(),
		ContentType: contentType,
	}, nil
}
