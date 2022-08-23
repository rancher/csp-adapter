package metrics

import (
	"fmt"
	"net/http"
	"strings"

	prometheusClient "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/sirupsen/logrus"
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
	clusterNameLabel    = "cluster_id"
	localClusterID      = "local"
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
		isMetricForLocal, err := isMetricForLocalCluster(metric)
		clusterNodeCount := int(metric.GetGauge().GetValue())
		logrus.Debugf("scraper found nodes: %d, isMetricForLocal: %t, err: %v", clusterNodeCount, isMetricForLocal, err)
		if err != nil {
			logrus.Warnf("error when attempting to determine if count was for local cluster: %s, will not include %d nodes in total", err.Error(), clusterNodeCount)
			continue
		}
		if !isMetricForLocal {
			nodeCount += clusterNodeCount
		}

	}

	return &NodeCounts{
		Total: nodeCount,
	}, nil
}

func isMetricForLocalCluster(metric *prometheusClient.Metric) (bool, error) {
	for _, label := range metric.GetLabel() {
		if label.Name != nil && *label.Name == clusterNameLabel {
			return label.Value != nil && *label.Value == localClusterID, nil
		}
	}
	return false, fmt.Errorf("unable to determine if metric is for local cluster due to missing label")
}
