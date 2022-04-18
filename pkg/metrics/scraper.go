package metrics

import (
	"net/http"
	"strings"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

// Scraper defines behavior that a Rancher metrics scraper should implement
type Scraper interface {
	ScrapeAndParse() (map[string]*dto.MetricFamily, error)
}

type scraper struct {
	rancherURL string
	cli        *http.Client
}

func NewScraper(rancherHostname string) Scraper {
	return &scraper{rancherURL: strings.Join([]string{"https://", rancherHostname}, "")}
}

func (s *scraper) ScrapeAndParse() (map[string]*dto.MetricFamily, error) {
	res, err := s.cli.Get(s.rancherURL + "/metrics")
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	var parser expfmt.TextParser
	mf, err := parser.TextToMetricFamilies(res.Body)
	if err != nil {
		return nil, err
	}

	return mf, nil
}
