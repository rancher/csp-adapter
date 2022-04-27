package aws

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	lm "github.com/aws/aws-sdk-go-v2/service/licensemanager"
	"github.com/aws/aws-sdk-go-v2/service/licensemanager/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Client interface {
	AccountNumber() string
	ListRancherReceivedLicenses(ctx context.Context) ([]types.GrantedLicense, error)
	CheckoutRancherLicense(ctx context.Context, l types.GrantedLicense) (*lm.CheckoutLicenseOutput, error)
	CheckInRancherLicense(ctx context.Context, consumptionToken string) (*lm.CheckInLicenseOutput, error)
	ExtendRancherLicenseConsumptionToken(ctx context.Context, consumptionToken string) (*lm.ExtendLicenseConsumptionOutput, error)
}

type client struct {
	acctNum string
	cfg     aws.Config
	sts     *sts.Client
	lm      *lm.Client
}

func setMetadataService(c *imds.Client) func(o *ec2rolecreds.Options) {
	return func(o *ec2rolecreds.Options) {
		o.Client = c
	}
}

func NewClient(ctx context.Context) (Client, error) {
	// initialize instance metadata service and credential provider from imds
	ms := imds.New(imds.Options{})
	provider := ec2rolecreds.New(setMetadataService(ms))
	_, err := provider.Retrieve(ctx)
	if err != nil {
		return nil, err
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(provider))
	if err != nil {
		return nil, err
	}

	// set region
	res, err := ms.GetRegion(ctx, &imds.GetRegionInput{})
	if err != nil {
		return nil, err
	}
	cfg.Region = res.Region

	logrus.Debugf("aws config: %+v", cfg)

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
	rancherProductSKU       = "65c5a63f-721f-4a70-b814-3ad8ab52dd8f"
	maxResults        int32 = 25
)

func (c *client) ListRancherReceivedLicenses(ctx context.Context) ([]types.GrantedLicense, error) {
	var grantedLicenses []types.GrantedLicense
	input := &lm.ListReceivedLicensesInput{
		Filters: []types.Filter{
			types.Filter{
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

	if len(res.Licenses) > 0 {
		grantedLicenses = append(grantedLicenses, res.Licenses...)
	}

	for {
		if res.NextToken == nil {
			break
		}

		input.NextToken = res.NextToken
		res, err = c.lm.ListReceivedLicenses(ctx, input)
		if err != nil {
			return nil, err
		}

		if len(res.Licenses) > 0 {
			grantedLicenses = append(grantedLicenses, res.Licenses...)
		}
	}

	return grantedLicenses, nil
}

var (
	entitlementDimension = "RKE_NODE_SUPP"
	checkoutCount        = "1"
)

const (
	entitlementUnit = "Count"
)

func (c *client) CheckoutRancherLicense(ctx context.Context, l types.GrantedLicense) (*lm.CheckoutLicenseOutput, error) {
	if l.Issuer == nil || l.Issuer.KeyFingerprint == nil {
		return nil, fmt.Errorf("license %s must have a KeyFingerprint for checkout", *l.LicenseArn)
	}

	token := uuid.New().String()
	res, err := c.lm.CheckoutLicense(ctx, &lm.CheckoutLicenseInput{
		CheckoutType:   types.CheckoutTypeProvisional,
		ClientToken:    &token,
		ProductSKU:     &rancherProductSKU,
		KeyFingerprint: l.Issuer.KeyFingerprint,
		Entitlements: []types.EntitlementData{
			{
				Name:  &entitlementDimension,
				Unit:  entitlementUnit,
				Value: &checkoutCount,
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
