package mocks

import (
	"github.com/rancher/csp-adapter/pkg/metrics"
)

type MockScraper struct {
	Nodes int
}

func NewMockScraper(numNodes int) *MockScraper {
	return &MockScraper{
		Nodes: numNodes,
	}
}

func (m *MockScraper) ScrapeAndParse() (*metrics.NodeCounts, error) {
	// TODO: Error case
	return &metrics.NodeCounts{
		Total: m.Nodes,
	}, nil
}
