package types

type SupportConfig struct {
	SupportEligible bool           `json:"support_eligible"`
	CSP             CSPInfo        `json:"csp"`
	Compliance      ComplianceInfo `json:"compliance"`
	Platform        string         `json:"platform"`
	Product         string         `json:"product"`
}

type CSPInfo struct {
	Name       string `json:"name"`
	AcctNumber string `json:"acct_number"`
}

const (
	StatusInCompliance    = "Compliant"
	StatusNotInCompliance = "NonCompliant"
)

type ComplianceInfo struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
