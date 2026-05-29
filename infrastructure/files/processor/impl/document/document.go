package document

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"

	"github.com/YagoSchramm/GoDepot/domain/entity"
	"github.com/YagoSchramm/GoDepot/domain/entity/derr"
	"github.com/YagoSchramm/GoDepot/infrastructure/files/processor"
)

type textMeta struct {
	Name      string `json:"name"`
	WordCount int    `json:"word_count"`
	LineCount int    `json:"line_count"`
	Preview   string `json:"preview"`
}

type DocumentProcessor struct{}

func NewDocumentProcessor() processor.Processor {
	return &DocumentProcessor{}
}

func (p *DocumentProcessor) CanHandle(mimeType string) bool {
	switch mimeType {
	case "application/pdf", "text/plain", "text/markdown":
		return true
	}
	return false
}

func (p *DocumentProcessor) Process(file entity.File, opts entity.Options) (entity.Result, error) {
	switch file.MimeType {
	case "text/plain", "text/markdown":
		return p.processText(file, opts)
	case "application/pdf":
		return p.processPDF(file)
	}

	return entity.Result{}, derr.NewClientError("UNSUPPORTED_DOCUMENT_TYPE", "unsupported document type")
}

func (p *DocumentProcessor) processText(file entity.File, opts entity.Options) (entity.Result, error) {
	f, err := os.Open(file.Path)
	if err != nil {
		return entity.Result{}, err
	}
	defer f.Close()

	var lines []string
	wordCount := 0
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
		wordCount += len(strings.Fields(line))
	}
	if err := scanner.Err(); err != nil {
		return entity.Result{}, err
	}

	previewLines := min(3, len(lines))
	preview := strings.Join(lines[:previewLines], "\n")

	meta := textMeta{
		Name:      file.Name,
		WordCount: wordCount,
		LineCount: len(lines),
		Preview:   preview,
	}

	data, err := json.Marshal(meta)
	if err != nil {
		return entity.Result{}, err
	}
	return entity.Result{
		Data:        data,
		ContentType: "application/json",
	}, nil
}

type pdfMeta struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	Note string `json:"note"`
}

func (p *DocumentProcessor) processPDF(file entity.File) (entity.Result, error) {
	meta := pdfMeta{
		Name: file.Name,
		Size: file.Size,
		Note: "full PDF parsing not implemented yet",
	}
	data, err := json.Marshal(meta)
	if err != nil {
		return entity.Result{}, err
	}
	return entity.Result{
		Data:        data,
		ContentType: "application/json",
	}, nil
}
