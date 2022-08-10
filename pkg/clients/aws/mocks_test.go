package aws

import (
	"context"
	"fmt"
	"time"

	lm "github.com/aws/aws-sdk-go-v2/service/licensemanager"
	"github.com/aws/aws-sdk-go-v2/service/licensemanager/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

const timeFormat = time.RFC3339

type mockLicenseManagerClient struct {
	licenses           map[string]types.GrantedLicense
	checkedOutLicenses map[string]licenseInfo
	licenseCounter     int
}

type mockSTSClient struct {
	accountNumber string
}

type licenseInfo struct {
	checkOutInput lm.CheckoutLicenseInput
	expiryTime    time.Time
}

func (m *mockLicenseManagerClient) AddLicenseForSku(productSku string, accountNumber string, includeSkuInReturn bool) {
	if m.licenses == nil {
		m.licenses = map[string]types.GrantedLicense{}
	}
	licenseArn := fmt.Sprintf("arn:aws:license-manager::%s:license:l-%06d", accountNumber, m.licenseCounter)
	m.licenseCounter++
	license := types.GrantedLicense{
		LicenseArn: &licenseArn,
	}
	if includeSkuInReturn {
		license.ProductSKU = &productSku
	}
	m.licenses[productSku] = license
}

func (m *mockLicenseManagerClient) Clear() {
	m.licenses = map[string]types.GrantedLicense{}
	m.checkedOutLicenses = map[string]licenseInfo{}
	m.licenseCounter = 0
}

func (m *mockLicenseManagerClient) ListReceivedLicenses(ctx context.Context, params *lm.ListReceivedLicensesInput, optFns ...func(*lm.Options)) (*lm.ListReceivedLicensesOutput, error) {
	var productIDs []string
	for _, filter := range params.Filters {
		if *filter.Name == productSKUField {
			productIDs = filter.Values
		}
	}
	var licenses []types.GrantedLicense
	for _, productID := range productIDs {
		if license, ok := m.licenses[productID]; ok {
			licenses = append(licenses, license)
		}
	}
	return &lm.ListReceivedLicensesOutput{
		Licenses: licenses,
	}, nil
}
func (m *mockLicenseManagerClient) CheckoutLicense(ctx context.Context, params *lm.CheckoutLicenseInput, optFns ...func(*lm.Options)) (*lm.CheckoutLicenseOutput, error) {
	consumptionToken := params.ClientToken
	if consumptionToken == nil {
		return nil, fmt.Errorf("unable to checkout license, no consumption token provided")
	}
	expiryTime := time.Now().Add(time.Hour * 24)
	expiryTS := expiryTime.Format(timeFormat)
	m.checkedOutLicenses[*consumptionToken] = licenseInfo{
		checkOutInput: *params,
		expiryTime:    expiryTime,
	}
	return &lm.CheckoutLicenseOutput{
		LicenseConsumptionToken: consumptionToken,
		Expiration:              &expiryTS,
	}, nil
}
func (m *mockLicenseManagerClient) CheckInLicense(ctx context.Context, params *lm.CheckInLicenseInput, optFns ...func(*lm.Options)) (*lm.CheckInLicenseOutput, error) {
	if params.LicenseConsumptionToken == nil {
		return nil, fmt.Errorf("can't check in license without consumption token")
	}
	if _, ok := m.checkedOutLicenses[*params.LicenseConsumptionToken]; ok {
		return nil, fmt.Errorf("no license checked in for consumption token")
	}
	delete(m.checkedOutLicenses, *params.LicenseConsumptionToken)
	return nil, nil
}
func (m *mockLicenseManagerClient) ExtendLicenseConsumption(ctx context.Context, params *lm.ExtendLicenseConsumptionInput, optFns ...func(*lm.Options)) (*lm.ExtendLicenseConsumptionOutput, error) {
	token := params.LicenseConsumptionToken
	if token == nil {
		return nil, fmt.Errorf("no token provided, cannot extend checkout")
	}
	if _, ok := m.checkedOutLicenses[*token]; ok {
		return nil, fmt.Errorf("no license checked in for consumption token")
	}
	return nil, nil
}
func (m *mockLicenseManagerClient) GetLicenseUsage(ctx context.Context, params *lm.GetLicenseUsageInput, optFns ...func(*lm.Options)) (*lm.GetLicenseUsageOutput, error) {
	licenseArn := params.LicenseArn
	if licenseArn == nil {
		return nil, fmt.Errorf("license arn is missing but is required")
	}
	var entitlementUsage []types.EntitlementUsage
	for _, value := range m.checkedOutLicenses {
		// find only the check-outs for this license
		if license, ok := m.licenses[*value.checkOutInput.ProductSKU]; ok {
			if *license.LicenseArn == *licenseArn {
				// add each usage separately, no need to combine into one
				for _, data := range value.checkOutInput.Entitlements {
					entitlementUsage = append(entitlementUsage, types.EntitlementUsage{Name: data.Name, ConsumedValue: data.Value, Unit: data.Unit})
				}
			}
		}
	}
	return &lm.GetLicenseUsageOutput{
		LicenseUsage: &types.LicenseUsage{EntitlementUsages: entitlementUsage}}, nil
}

func (m *mockSTSClient) GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
	return &sts.GetCallerIdentityOutput{Account: &m.accountNumber}, nil
}
