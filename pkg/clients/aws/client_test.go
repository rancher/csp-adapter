package aws

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

const fakeAccountNum = "123456789101"

func TestGetRancherLicense(t *testing.T) {
	tests := []struct {
		name              string // name of the test, to be displayed on failure
		hasNonEmeaLicense bool   // if the account has a license for the non-EMEA product sku
		hasEmeaLicense    bool   // if the account has a licensed for the EMEA product sku
		includeProductSku bool   // if the return from aws should include or exclude a product sku
		desiredLicense    string // which license our client should pick - emea, non-emea, or nothing
		errDesired        bool   // if we wanted an error for this test case
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
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			mockLMClient := mockLicenseManagerClient{}
			client := &client{
				acctNum: fakeAccountNum,
				lm:      &mockLMClient,
				sts:     &mockSTSClient{accountNumber: fakeAccountNum},
			}
			if test.hasNonEmeaLicense {
				mockLMClient.AddLicenseForSku(rancherProductSKUNonEmea, fakeAccountNum, test.includeProductSku)
			}
			if test.hasEmeaLicense {
				mockLMClient.AddLicenseForSku(rancherProductSKUEmea, fakeAccountNum, test.includeProductSku)
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
