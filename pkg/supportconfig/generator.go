package supportconfig

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"io"

	"github.com/rancher/csp-adapter/pkg/types"
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
	config := &types.SupportConfig{
		SupportEligible: true,
		CSP: types.CSPInfo{
			Name:       g.csp,
			AcctNumber: g.acctNumber,
		},
		Compliance: types.ComplianceInfo{
			Status: types.StatusInCompliance,
		},
		Platform: "x86_64",
		Product:  "cpe:/o:suse:rancher:v2.6.3",
	}

	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	files := map[string]string{
		"rancher/config.json": string(configData),
	}

	for name, body := range files {
		hdr := &tar.Header{
			Name: name,
			Mode: 0600,
			Size: int64(len(body)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return nil, err
		}
		if _, err := tw.Write(configData); err != nil {
			return nil, err
		}
	}

	return &buf, nil
}
