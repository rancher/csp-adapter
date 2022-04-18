package metrics

import (
	"fmt"
	"net/http"
	"os"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

const (
	file = "./metrics.txt"
)

// Scraper defines behavior that a Rancher metrics scraper should implement
type Scraper interface {
	ScrapeAndParse() (map[string]*dto.MetricFamily, error)
}

type scraper struct {
	rancherURL string
	cli        *http.Client
}

const (
	rancherHostnameEnv = "RANCHER_HOSTNAME" // this should come from the server-url setting in future
)

func NewScraper() (Scraper, error) {
	rancherURL := os.Getenv(rancherHostnameEnv)
	if rancherURL == "" {
		return nil, fmt.Errorf("%s must be specified in env", rancherHostnameEnv)
	}
	return &scraper{rancherURL: rancherURL}, nil
}

func (s *scraper) ScrapeAndParse() (map[string]*dto.MetricFamily, error) {
	res, err := s.cli.Get(s.rancherURL + "/metrics")
	if err != nil {
		return nil, err
	}

	var parser expfmt.TextParser
	defer res.Body.Close() // parser won't close so do it here
	mf, err := parser.TextToMetricFamilies(res.Body)
	if err != nil {
		return nil, err
	}

	return mf, nil
}
