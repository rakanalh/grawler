package extract

// Css defines an interface that implements searching through CSS
import (
	"bytes"

	"github.com/PuerkitoBio/goquery"
)

type Css struct {
	doc *goquery.Document
}

func NewCss(content []byte) (*Css, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(content))
	if err != nil {
		return nil, err
	}

	return &Css{doc: doc}, nil
}

func (extractor *Css) Extract(selector string) (goquery.Selection, error) {
	return *extractor.doc.Find(selector), nil
}
