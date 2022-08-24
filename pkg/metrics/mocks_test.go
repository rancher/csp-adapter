package metrics

import (
	"fmt"
	"net/http"
	"strings"
)

type mockPrometheusServer struct {
	nodesForCluster map[string]int
	labelSkip       map[string]struct{}
	validTokens     []string
}

func newMockPrometheusServer() mockPrometheusServer {
	return mockPrometheusServer{
		nodesForCluster: map[string]int{},
		labelSkip:       map[string]struct{}{},
		validTokens:     []string{},
	}
}

func (m *mockPrometheusServer) SetNodesForCluster(nodes int, cluster string, skipLabel bool) {
	m.nodesForCluster[cluster] = nodes
	if skipLabel {
		m.labelSkip[cluster] = struct{}{}
	}
}

func (m *mockPrometheusServer) Clear() {
	m.nodesForCluster = map[string]int{}
	m.validTokens = []string{}
	m.labelSkip = map[string]struct{}{}
}

func (m *mockPrometheusServer) AddAuthToken(token string) {
	m.validTokens = append(m.validTokens, token)
}
func (m *mockPrometheusServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// main serve method for the test http server
	if strings.HasSuffix(r.URL.String(), "/metrics") {
		//do auth and return the node metrics text
		if !m.requestAuthenticated(r) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if !m.requestAuthorized(r) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		requestText := m.getNodeMetricsText()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(requestText))
		return
	}
	w.WriteHeader(http.StatusNotFound)
	return
}

func (m *mockPrometheusServer) getNodeMetricsText() string {
	retText := "# TYPE cluster_manager_nodes gauge\n"
	for cluster, nodeCount := range m.nodesForCluster {
		if _, ok := m.labelSkip[cluster]; ok {
			retText += fmt.Sprintf("%s{} %d\n", nodeGaugeMetricName, nodeCount)
		} else {
			retText += fmt.Sprintf("%s{%s=\"%s\"} %d\n", nodeGaugeMetricName, clusterNameLabel, cluster, nodeCount)
		}
	}
	return retText
}

func (m *mockPrometheusServer) requestAuthorized(req *http.Request) bool {
	header := req.Header.Get("Authorization")
	splitHeader := strings.Split(header, " ")
	if len(splitHeader) < 2 {
		return false
	}
	token := splitHeader[1]
	for _, validToken := range m.validTokens {
		if token == validToken {
			return true
		}
	}
	return false
}

func (m *mockPrometheusServer) requestAuthenticated(req *http.Request) bool {
	return req.Header.Get("Authorization") != ""
}
