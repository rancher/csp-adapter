package mocks

import (
	"context"
	"fmt"
	"strconv"
	"time"

	lm "github.com/aws/aws-sdk-go-v2/service/licensemanager"
	"github.com/aws/aws-sdk-go-v2/service/licensemanager/types"
)

type MockAWSClient struct {
	AWSAccountNumber       string
	License                types.GrantedLicense
	CheckedOutEntitlements map[string]int
	CheckoutTokenCtr       int
}

const (
	rkeEntitlement = "RKE_NODE_SUPP"
	fakeAWSAccount = "111111111111"
	fakeLicenseID  = "l-12345"
)

func NewMockAWSClient(maxEntitlements int) *MockAWSClient {
	fakeLicenseArn := fmt.Sprintf("arn:aws:license-manager::%s:license:%s", fakeAWSAccount, fakeLicenseID)
	entitlementName := rkeEntitlement
	// technically not a lossless conversion, but we should never run a test with values this high
	maxCount := int64(maxEntitlements)
	return &MockAWSClient{
		AWSAccountNumber: fakeAWSAccount,
		License: types.GrantedLicense{
			LicenseArn: &fakeLicenseArn,
			Entitlements: []types.Entitlement{{
				Name:     &entitlementName,
				Unit:     types.EntitlementUnitCount,
				MaxCount: &maxCount,
			}},
		},
		CheckedOutEntitlements: map[string]int{},
	}
}

func (m *MockAWSClient) AccountNumber() string {
	return m.AWSAccountNumber
}

func (m *MockAWSClient) GetRancherLicense(ctx context.Context) (*types.GrantedLicense, error) {
	return &m.License, nil
}

func (m *MockAWSClient) CheckoutRancherLicense(ctx context.Context, l types.GrantedLicense, entitlementAmt int) (*lm.CheckoutLicenseOutput, error) {
	if *l.LicenseArn != *m.License.LicenseArn {
		//TODO: Not found aws error mock
		return nil, fmt.Errorf("license not found")
	}
	// remove from checkedIn and append to checkedOut
	consumptionToken := m.genConsumptionToken()
	currentTotal := 0
	for _, value := range m.CheckedOutEntitlements {
		currentTotal += value
	}
	if currentTotal > m.getMaxRKEEntitlements() {
		//TODO: maybe return with less entitlements?
		return nil, fmt.Errorf("can't checkout license - over entitlements")
	}
	m.CheckedOutEntitlements[consumptionToken] = entitlementAmt

	expiryTime := time.Now().Add(1 * time.Hour).Format(time.RFC3339)
	name := rkeEntitlement
	value := strconv.Itoa(entitlementAmt)
	return &lm.CheckoutLicenseOutput{
		// our checkouts are always provisional
		CheckoutType: types.CheckoutTypeProvisional,
		EntitlementsAllowed: []types.EntitlementData{{
			Name:  &name,
			Value: &value,
			Unit:  types.EntitlementDataUnitCount,
		}},
		Expiration:              &expiryTime,
		LicenseArn:              l.LicenseArn,
		LicenseConsumptionToken: &consumptionToken,
	}, nil
}

func (m *MockAWSClient) CheckInRancherLicense(ctx context.Context, consumptionToken string) (*lm.CheckInLicenseOutput, error) {
	//TODO: not found consumption token aws error mock
	_, ok := m.CheckedOutEntitlements[consumptionToken]
	if !ok {
		return nil, fmt.Errorf("invalid token")
	}
	delete(m.CheckedOutEntitlements, consumptionToken)
	return &lm.CheckInLicenseOutput{}, nil
}

func (m *MockAWSClient) ExtendRancherLicenseConsumptionToken(ctx context.Context, consumptionToken string) (*lm.ExtendLicenseConsumptionOutput, error) {
	_, ok := m.CheckedOutEntitlements[consumptionToken]
	if !ok {
		return nil, fmt.Errorf("invalid token")
	}
	expiryTime := time.Now().Add(1 * time.Hour).Format(time.RFC3339)
	return &lm.ExtendLicenseConsumptionOutput{
		LicenseConsumptionToken: &consumptionToken,
		Expiration:              &expiryTime,
	}, nil
}

func (m *MockAWSClient) GetNumberOfAvailableEntitlements(ctx context.Context, license types.GrantedLicense) (int, error) {
	currentTotal := 0
	for _, value := range m.CheckedOutEntitlements {
		currentTotal += value
	}
	maxEntitlements := m.getMaxRKEEntitlements()
	remaining := maxEntitlements - currentTotal
	if remaining < 0 {
		return 0, fmt.Errorf("over entitlements")
	}
	return remaining, nil
}

func (m *MockAWSClient) genConsumptionToken() string {
	m.CheckoutTokenCtr++
	return fmt.Sprintf("%d", m.CheckoutTokenCtr)
}

func (m *MockAWSClient) getMaxRKEEntitlements() int {
	for _, entitlement := range m.License.Entitlements {
		if *entitlement.Name == rkeEntitlement {
			return int(*entitlement.MaxCount)
		}
	}
	return 0
}
