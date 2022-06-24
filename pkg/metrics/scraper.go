package metrics

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/prometheus/common/expfmt"
	"k8s.io/client-go/rest"
)

// Scraper defines behavior that a Rancher metrics scraper should implement
type Scraper interface {
	ScrapeAndParse() (*NodeCounts, error)
}

type scraper struct {
	metricsURL string
	cli        *http.Client
	cfg        *rest.Config
}

func NewScraper(rancherHost string, cfg *rest.Config) Scraper {
	return &scraper{
		metricsURL: strings.Join([]string{"https://", rancherHost, "/metrics"}, ""),
		cli:        &http.Client{},
		cfg:        cfg,
	}
}

const (
	nodeGaugeMetricName = "cluster_manager_nodes"
)

type NodeCounts struct {
	Total int
}

func (s *scraper) ScrapeAndParse() (*NodeCounts, error) {
	req, err := http.NewRequest(http.MethodGet, s.metricsURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.cfg.BearerToken))

	res, err := s.cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error got %v response", res.StatusCode)
	}

	var parser expfmt.TextParser
	metricFamilies, err := parser.TextToMetricFamilies(res.Body)
	if err != nil {
		return nil, err
	}

	nodeMetricFamily, ok := metricFamilies[nodeGaugeMetricName]
	if !ok {
		return nil, fmt.Errorf("no metric with name %s found in rancher /metrics output", nodeGaugeMetricName)
	}

	var nodeCount int
	for _, metric := range nodeMetricFamily.GetMetric() {
		nodeCount += int(metric.GetGauge().GetValue())
	}

	return &NodeCounts{
		Total: nodeCount,
	}, nil
}
