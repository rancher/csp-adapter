package supportconfig

import (
	"bytes"
	"io"
)

type Generator interface {
	SupportConfig() (io.Reader, error)
}

type generator struct {
	csp        string
	acctNumber string
}

func NewGenerator(csp, acctNumber string) (Generator, error) {
	return &generator{
		csp:        csp,
		acctNumber: acctNumber,
	}, nil
}

// AWSConfig creates an aws specific support config
func (g *generator) SupportConfig() (io.Reader, error) {
	return bytes.NewBuffer(nil), nil
}
