package aws

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/licensemanager"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type Client interface {
	AccountNumber() string
}

type client struct {
	acctNum string
	cfg     aws.Config
	lm      *licensemanager.Client
	sts     *sts.Client
}

func NewClient(ctx context.Context) (Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	c := &client{
		cfg: cfg,
		lm:  licensemanager.NewFromConfig(cfg),
		sts: sts.NewFromConfig(cfg),
	}

	acctNum, err := c.getAccountNumber(ctx)
	if err != nil {
		return nil, err
	}

	c.acctNum = acctNum

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
