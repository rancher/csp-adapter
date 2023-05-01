package aws

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

const fakeAccountNum = "123456789101"

func TestGetRancherLicense(t *testing.T) {
	tests := []struct {
		name                  string // name of the test, to be displayed on failure
		hasNonEmeaLicense     bool   // if the account has a license for the non-EMEA product sku
		hasEmeaLicense        bool   // if the account has a licensed for the EMEA product sku
		hasTestNonEmeaLicense bool   // if the account has a license for the test non-EMEA product sku
		hasTestEmeaLicense    bool   // if the account has a license for the test EMEA product sku
		usesTestIds           bool   // if the account is using a test product id
		includeProductSku     bool   // if the return from aws should include or exclude a product sku
		desiredLicense        string // which license our client should pick - emea, non-emea, or nothing
		errDesired            bool   // if we wanted an error for this test case
	}{
		{
			name:              "test non-emea license",
			hasNonEmeaLicense: true,
			hasEmeaLicense:    false,
			includeProductSku: true,
			desiredLicense:    rancherProductSKUNonEmea,
			errDesired:        false,
		},
		{
			name:              "test emea license",
			hasNonEmeaLicense: false,
			hasEmeaLicense:    true,
			includeProductSku: true,
			desiredLicense:    rancherProductSKUEmea,
			errDesired:        false,
		},
		{
			name:              "test non-emea + emea license - should not occur in reality",
			hasNonEmeaLicense: true,
			hasEmeaLicense:    true,
			includeProductSku: true,
			desiredLicense:    rancherProductSKUNonEmea,
			errDesired:        false,
		},
		{
			name:              "test no valid license",
			hasNonEmeaLicense: false,
			hasEmeaLicense:    false,
			includeProductSku: false,
			desiredLicense:    "",
			errDesired:        true,
		},
		{
			name:              "test no product sku non-emea",
			hasNonEmeaLicense: true,
			hasEmeaLicense:    false,
			includeProductSku: false,
			desiredLicense:    rancherProductSKUNonEmea,
			errDesired:        false,
		},
		{
			name:              "test no product sku emea",
			hasNonEmeaLicense: false,
			hasEmeaLicense:    true,
			includeProductSku: false,
			desiredLicense:    rancherProductSKUEmea,
			errDesired:        false,
		},
		{
			name:                  "test non-emea test license",
			hasTestNonEmeaLicense: true,
			usesTestIds:           true,
			includeProductSku:     true,
			desiredLicense:        rancherProductTestSKUNonEmea,
		},
		{
			name:               "test emea test license",
			hasTestEmeaLicense: true,
			usesTestIds:        true,
			includeProductSku:  true,
			desiredLicense:     rancherProductTestSKUEmea,
		},
		{
			name:                  "test non-emea + emea test license",
			hasTestNonEmeaLicense: true,
			hasTestEmeaLicense:    true,
			usesTestIds:           true,
			includeProductSku:     true,
			desiredLicense:        rancherProductTestSKUNonEmea,
		},
		{
			name:                  "test non-emea test + prod skus, test requested, test used",
			hasNonEmeaLicense:     true,
			hasTestNonEmeaLicense: true,
			usesTestIds:           true,
			includeProductSku:     true,
			desiredLicense:        rancherProductTestSKUNonEmea,
		},
		{
			name:                  "test non-emea test + prod skus, prod requested, prod used",
			hasNonEmeaLicense:     true,
			hasTestNonEmeaLicense: true,
			usesTestIds:           false,
			includeProductSku:     true,
			desiredLicense:        rancherProductSKUNonEmea,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			mockLMClient := mockLicenseManagerClient{}
			client := &client{
				acctNum:         fakeAccountNum,
				lm:              &mockLMClient,
				sts:             &mockSTSClient{accountNumber: fakeAccountNum},
				useTestProducts: test.usesTestIds,
			}
			if test.hasNonEmeaLicense {
				mockLMClient.AddLicenseForSku(rancherProductSKUNonEmea, fakeAccountNum, test.includeProductSku)
			}
			if test.hasTestNonEmeaLicense {
				mockLMClient.AddLicenseForSku(rancherProductTestSKUNonEmea, fakeAccountNum, test.includeProductSku)
			}
			if test.hasEmeaLicense {
				mockLMClient.AddLicenseForSku(rancherProductSKUEmea, fakeAccountNum, test.includeProductSku)
			}
			if test.hasTestEmeaLicense {
				mockLMClient.AddLicenseForSku(rancherProductTestSKUEmea, fakeAccountNum, test.includeProductSku)
			}

			license, err := client.GetRancherLicense(context.Background())
			if test.errDesired {
				assert.Error(t, err, "expected an error but err was nil")
			} else {
				assert.NoError(t, err, "no error was expected, but got an error")
				assert.NotNil(t, license, "expected a valid license but was nil")
				assert.Equal(t, *license.ProductSKU, test.desiredLicense, "received unexpected product sku")
			}
		})
	}
}
