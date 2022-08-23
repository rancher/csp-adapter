package metrics

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/rest"
)

func TestScrapeAndParse(t *testing.T) {
	metricsServer := newMockPrometheusServer()
	server := httptest.NewServer(&metricsServer)
	defer server.Close()
	tests := []struct {
		name                  string
		localClusterNodes     int
		skipLocalLabel        bool
		numOtherClusters      int
		nodesPerOtherClusters int
		skipOtherLabel        bool
		authed                bool
		expectedTotal         int
		expectedError         bool
	}{
		{
			name:                  "no downstream, local only",
			localClusterNodes:     2,
			skipLocalLabel:        false,
			numOtherClusters:      0,
			nodesPerOtherClusters: 0,
			skipOtherLabel:        false,
			authed:                true,
			expectedTotal:         0,
			expectedError:         false,
		},
		{
			name:                  "downstream and local",
			localClusterNodes:     2,
			skipLocalLabel:        false,
			numOtherClusters:      1,
			nodesPerOtherClusters: 2,
			skipOtherLabel:        false,
			authed:                true,
			expectedTotal:         2,
			expectedError:         false,
		},
		{
			name:                  "multiple downstream and local",
			localClusterNodes:     2,
			skipLocalLabel:        false,
			numOtherClusters:      3,
			nodesPerOtherClusters: 4,
			skipOtherLabel:        false,
			authed:                true,
			expectedTotal:         12,
			expectedError:         false,
		},
		{
			name:                  "local not labeled, local nodes excluded",
			localClusterNodes:     2,
			skipLocalLabel:        true,
			numOtherClusters:      0,
			nodesPerOtherClusters: 0,
			skipOtherLabel:        false,
			authed:                true,
			expectedTotal:         0,
			expectedError:         false,
		},
		{
			name:                  "local not labeled, downstream not labeled, all nodes excluded",
			localClusterNodes:     2,
			skipLocalLabel:        true,
			numOtherClusters:      2,
			nodesPerOtherClusters: 3,
			skipOtherLabel:        true,
			authed:                true,
			expectedTotal:         0,
			expectedError:         false,
		},
		{
			name:                  "local labeled, downstream not labeled, all nodes excluded",
			localClusterNodes:     2,
			skipLocalLabel:        false,
			numOtherClusters:      2,
			nodesPerOtherClusters: 4,
			skipOtherLabel:        true,
			authed:                true,
			expectedTotal:         0,
			expectedError:         false,
		},
		{
			name:                  "error, no auth",
			localClusterNodes:     2,
			skipLocalLabel:        false,
			numOtherClusters:      0,
			nodesPerOtherClusters: 0,
			skipOtherLabel:        false,
			authed:                false,
			expectedTotal:         0,
			expectedError:         true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metricsServer.Clear()
			metricsServer.SetNodesForCluster(test.localClusterNodes, localClusterID, test.skipLocalLabel)
			for i := 0; i < test.numOtherClusters; i++ {
				metricsServer.SetNodesForCluster(test.nodesPerOtherClusters, fmt.Sprintf("cluster-%d", i), test.skipOtherLabel)
			}
			config := &rest.Config{BearerToken: "abc123abc123abc123"}
			if test.authed {
				metricsServer.AddAuthToken(config.BearerToken)
			}
			metricsScraper := scraper{
				metricsURL: fmt.Sprintf("%s/metrics", server.URL),
				cli:        &http.Client{},
				cfg:        config,
			}
			res, err := metricsScraper.ScrapeAndParse()
			if test.expectedError {
				assert.Error(t, err, "expected an error but err was nil")
			} else {
				assert.NoError(t, err, "expected no error but there was an error")
				assert.NotNil(t, res, "expected a result but was nil")
				assert.Equal(t, test.expectedTotal, res.Total, "did not get expected number of nodes")
			}
		})
	}
}
