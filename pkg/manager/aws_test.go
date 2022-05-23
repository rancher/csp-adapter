package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/rancher/csp-adapter/pkg/mocks"
	"github.com/stretchr/testify/assert"
)

type testScenario struct {
	numRancherNodes     int
	numAWSEntitlements  int
	currentEntitlements int
	result              testResult
}

type testResult struct {
	errResult           bool
	inCompliance        bool
	numUsedEntitlements int
	cachedToken         bool
}

func (s *testScenario) runScenario(t *testing.T) {
	mockAWSClient := mocks.NewMockAWSClient(s.numAWSEntitlements)
	var secretData map[string]string
	if s.currentEntitlements != 0 {
		output, _ := mockAWSClient.CheckoutRancherLicense(context.TODO(), mockAWSClient.License, s.currentEntitlements)
		checkedOut := strconv.Itoa(s.currentEntitlements)
		secretData = map[string]string{
			tokenKey:  *output.LicenseConsumptionToken,
			expiryKey: *output.Expiration,
			nodeKey:   checkedOut,
		}
	}
	mockK8sClient := mocks.NewMockK8sClient(secretData)
	mockScraper := mocks.NewMockScraper(s.numRancherNodes)
	mockAWS := AWS{
		aws:     mockAWSClient,
		k8s:     mockK8sClient,
		scraper: mockScraper,
	}
	err := mockAWS.runComplianceCheck(context.TODO())

	// check that the results were as expected
	if s.result.errResult {
		assert.Error(t, err, fmt.Sprintf("Scenario: %v", s))
	} else {
		assert.NoError(t, err, fmt.Sprintf("Scenario: %v", s))
	}
	var expectedCompliance string
	if s.result.inCompliance {
		expectedCompliance = StatusInCompliance
		assert.Equalf(t, "", mockK8sClient.CurrentNotificationMessage, "No notification expected for a pass")
	} else {
		expectedCompliance = StatusNotInCompliance
		assert.NotEqualf(t, "", mockK8sClient.CurrentNotificationMessage, "Some notification expected for a pass")
	}
	var config CSPSupportConfig
	err = json.Unmarshal(mockK8sClient.CurrentSupportConfig, &config)
	assert.NoError(t, err, "expected to be able to marshall config output to a cspSupportConfig")
	actualCompliance := config.Compliance.Status
	assert.Equal(t, expectedCompliance, actualCompliance, fmt.Sprintf("Scenario: %v", s))
	actualEntitlements := 0
	for _, value := range mockAWSClient.CheckedOutEntitlements {
		actualEntitlements += value
	}
	assert.Equal(t, s.result.numUsedEntitlements, actualEntitlements, fmt.Sprintf("Scenario: %v", s))
	if s.result.cachedToken {
		assert.NotNil(t, mockK8sClient.CurrentSecretData, fmt.Sprintf("Scenario: %v", s))
		_, ok := mockK8sClient.CurrentSecretData[tokenKey]
		assert.Equal(t, true, ok, fmt.Sprintf("No stored token for Scenario: %v", s))
	}
}

//TestCheckout tests basic scenarios where we don't have anything previously checked out
func TestCheckout(t *testing.T) {
	scenarios := []testScenario{
		{
			numRancherNodes:     20,
			numAWSEntitlements:  2,
			currentEntitlements: 0,
			result: testResult{
				errResult:           false,
				inCompliance:        true,
				numUsedEntitlements: 1,
				cachedToken:         true,
			},
		},
		{
			numRancherNodes:     40,
			numAWSEntitlements:  2,
			currentEntitlements: 0,
			result: testResult{
				errResult:           false,
				inCompliance:        true,
				numUsedEntitlements: 2,
				cachedToken:         true,
			},
		},
		{
			numRancherNodes:     40,
			numAWSEntitlements:  1,
			currentEntitlements: 0,
			result: testResult{
				errResult:           false,
				inCompliance:        false,
				numUsedEntitlements: 1,
				cachedToken:         true,
			},
		},
	}
	for _, scenario := range scenarios {
		scenario.runScenario(t)
	}
}

//TestCheckInCheckout tests scenarios where we already have some items checked out
func TestCheckInCheckout(t *testing.T) {
	scenarios := []testScenario{
		{
			numRancherNodes:     40,
			numAWSEntitlements:  2,
			currentEntitlements: 1,
			result: testResult{
				errResult:           false,
				inCompliance:        true,
				numUsedEntitlements: 2,
				cachedToken:         true,
			},
		},
		{
			numRancherNodes:     60,
			numAWSEntitlements:  2,
			currentEntitlements: 2,
			result: testResult{
				errResult:           false,
				inCompliance:        false,
				numUsedEntitlements: 2,
				cachedToken:         true,
			},
		},
		{
			numRancherNodes:     20,
			numAWSEntitlements:  2,
			currentEntitlements: 2,
			result: testResult{
				errResult:           false,
				inCompliance:        true,
				numUsedEntitlements: 1,
				cachedToken:         true,
			},
		},
	}
	for _, scenario := range scenarios {
		scenario.runScenario(t)
	}
}
