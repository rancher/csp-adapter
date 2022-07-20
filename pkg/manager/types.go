package manager

import (
	"fmt"
	"strings"

	"github.com/rancher/csp-adapter/pkg/clients/k8s"
)

type CSPSupportConfig struct {
	SupportEligible bool           `json:"support_eligible,omitempty"`
	Platform        string         `json:"platform"`
	Product         string         `json:"product"`
	CSP             CSPInfo        `json:"csp"`
	Compliance      ComplianceInfo `json:"compliance"`
}

type CSPInfo struct {
	Name       string `json:"name"`
	AcctNumber string `json:"acct_number"`
}

const (
	StatusInCompliance    = "Compliant"
	StatusNotInCompliance = "NonCompliant"
	defaultPlatform       = "x86_64"
	// SUSE support config reads EC2 as being for AWS, we want to use the same syntax to be consistent
	awsSupportConfigCSP = "EC2"
)

type ComplianceInfo struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// GetDefaultSupportConfig produces a CSPSupportConfig with values that could be inferred from k8s
func GetDefaultSupportConfig(client k8s.Client) CSPSupportConfig {
	rancherVersion, err := client.GetRancherVersion()
	if err != nil {
		rancherVersion = "unknown"
	}
	product := createProductString(rancherVersion)
	return CSPSupportConfig{
		SupportEligible: true,
		Platform:        defaultPlatform,
		Product:         product,
	}
}

func createProductString(rancherVersion string) string {
	// rancher version that comes from k8s is prefixed with a v, but suse lists the product version without a v
	productVersion := strings.TrimPrefix(rancherVersion, "v")
	return fmt.Sprintf("cpe:/o:suse:rancher:%s", productVersion)
}
