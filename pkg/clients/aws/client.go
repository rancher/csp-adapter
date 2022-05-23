// Package aws provides a high-level aws client for CSP functionality, including license check-in, checkout, and extension
package aws

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	lm "github.com/aws/aws-sdk-go-v2/service/licensemanager"
	"github.com/aws/aws-sdk-go-v2/service/licensemanager/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Client interface {
	// AccountNumber gets the account number for the AWS account this client will issue calls to
	AccountNumber() string
	// GetRancherLicense returns the license which is for the rancher product sku
	GetRancherLicense(ctx context.Context) (*types.GrantedLicense, error)
	// CheckoutRancherLicense checks out the license for entitlementAmt entitlements to RKE_NODE_SUPP
	CheckoutRancherLicense(ctx context.Context, l types.GrantedLicense, entitlementAmt int) (*lm.CheckoutLicenseOutput, error)
	// CheckInRancherLicense checks in a license using the provided consumptionToken
	CheckInRancherLicense(ctx context.Context, consumptionToken string) (*lm.CheckInLicenseOutput, error)
	// ExtendRancherLicenseConsumptionToken extends the Expiry time of the provided consumptionToken
	ExtendRancherLicenseConsumptionToken(ctx context.Context, consumptionToken string) (*lm.ExtendLicenseConsumptionOutput, error)
	// GetNumberOfAvailableEntitlements gets the number of RKE_NODE_SUPP entitlements available on license
	GetNumberOfAvailableEntitlements(ctx context.Context, license types.GrantedLicense) (int, error)
}

type client struct {
	acctNum string
	cfg     aws.Config
	sts     *sts.Client
	lm      *lm.Client
}

func NewClient(ctx context.Context) (Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	logrus.Debugf("aws config region: %+v", cfg.Region)

	c := &client{
		cfg: cfg,
		sts: sts.NewFromConfig(cfg),
		lm:  lm.NewFromConfig(cfg),
	}

	acctNum, err := c.getAccountNumber(ctx)
	if err != nil {
		return nil, err
	}

	c.acctNum = acctNum

	logrus.Debugf("account number: %s", acctNum)

	return c, nil
}

func (c *client) AccountNumber() string {
	return c.acctNum // set in constructor
}

// getAccountNumber returns the account number of the account to which the associated IAM user belongs.
func (c *client) getAccountNumber(ctx context.Context) (string, error) {
	var in sts.GetCallerIdentityInput
	out, err := c.sts.GetCallerIdentity(ctx, &in) // no permissions required to make this call
	if err != nil {
		return "", err
	}

	if out.Account == nil || len(*out.Account) == 0 {
		return "", errors.New("account number empty in aws sts response")
	}

	return *out.Account, nil
}

var (
	productSKUField         = "ProductSKU"
	rancherProductSKU       = "0b87d4fa-d1fe-41d8-830b-67d4ec381549"
	maxResults        int32 = 1
)

func (c *client) GetRancherLicense(ctx context.Context) (*types.GrantedLicense, error) {
	// per aws engineering, there should only ever be at most one license for a given product sku.
	input := &lm.ListReceivedLicensesInput{
		Filters: []types.Filter{
			{
				Name:   &productSKUField,
				Values: []string{rancherProductSKU},
			},
		},
		MaxResults: &maxResults,
	}

	res, err := c.lm.ListReceivedLicenses(ctx, input)
	if err != nil {
		return nil, err
	}

	return &res.Licenses[0], nil
}

var (
	entitlementDimension = "RKE_NODE_SUPP"
)

const (
	entitlementUnit = "Count"
)

func (c *client) CheckoutRancherLicense(ctx context.Context, l types.GrantedLicense, entitlementAmt int) (*lm.CheckoutLicenseOutput, error) {
	if l.Issuer == nil || l.Issuer.KeyFingerprint == nil {
		return nil, fmt.Errorf("license %s must have a KeyFingerprint for checkout", *l.LicenseArn)
	}

	token := uuid.New().String()
	entitlementStr := fmt.Sprintf("%d", entitlementAmt)
	res, err := c.lm.CheckoutLicense(ctx, &lm.CheckoutLicenseInput{
		CheckoutType:   types.CheckoutTypeProvisional,
		ClientToken:    &token,
		ProductSKU:     &rancherProductSKU,
		KeyFingerprint: l.Issuer.KeyFingerprint,
		Entitlements: []types.EntitlementData{
			{
				Name:  &entitlementDimension,
				Unit:  entitlementUnit,
				Value: &entitlementStr,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *client) CheckInRancherLicense(ctx context.Context, consumptionToken string) (*lm.CheckInLicenseOutput, error) {
	res, err := c.lm.CheckInLicense(ctx, &lm.CheckInLicenseInput{LicenseConsumptionToken: &consumptionToken})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *client) ExtendRancherLicenseConsumptionToken(ctx context.Context, consumptionToken string) (*lm.ExtendLicenseConsumptionOutput, error) {
	res, err := c.lm.ExtendLicenseConsumption(ctx, &lm.ExtendLicenseConsumptionInput{LicenseConsumptionToken: &consumptionToken})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *client) GetNumberOfAvailableEntitlements(ctx context.Context, license types.GrantedLicense) (int, error) {
	res, err := c.lm.GetLicenseUsage(ctx, &lm.GetLicenseUsageInput{LicenseArn: license.LicenseArn})
	if err != nil {
		// this function can't guarantee availability, so return 0 and an err so the caller can sort this out
		return 0, err
	}
	maxEntitlements, err := getMaxRKEEntitlements(license)
	if err != nil {
		// if we can't figure out how many RKE nodes we can support at max, we can't see how many we have left
		return 0, err
	}
	total := 0
	for _, usage := range res.LicenseUsage.EntitlementUsages {
		if *usage.Name == entitlementDimension {
			consumedValue, err := strconv.Atoi(*usage.ConsumedValue)
			if err != nil {
				return 0, err
			}
			total += consumedValue
		}
	}
	// this should be safe to do - we rely on licenseManager to control if we are/are not allowed to go over
	return maxEntitlements - total, nil
}

func getMaxRKEEntitlements(license types.GrantedLicense) (int, error) {
	for _, entitlement := range license.Entitlements {
		if *entitlement.Name == entitlementDimension {
			return int(*entitlement.MaxCount), nil
		}
	}
	return 0, fmt.Errorf("entitlement %s not found on license for %s", entitlementDimension, *license.LicenseArn)
}
