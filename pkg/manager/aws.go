package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/rancher/csp-adapter/pkg/clients/aws"
	"github.com/rancher/csp-adapter/pkg/clients/k8s"
	"github.com/rancher/csp-adapter/pkg/metrics"
	"github.com/sirupsen/logrus"
)

type AWS struct {
	cancel  context.CancelFunc
	aws     aws.Client
	k8s     k8s.Client
	scraper metrics.Scraper
}

func NewAWS(a aws.Client, k k8s.Client, s metrics.Scraper) *AWS {
	return &AWS{
		aws:     a,
		k8s:     k,
		scraper: s,
	}
}

func (m *AWS) Start(ctx context.Context, errs chan<- error) {
	go m.start(ctx, errs)
}

const (
	managerInterval = 30 * time.Second
	nodesPerLicense = 20
	// same as RFC3339 from time.time without the Z7:00 indicating timezone. Some AWS timestamps have this format
	rfc3339NoTZ = "2006-01-02T15:04:05"
	// keys for the consumption token secret's data. Can't do a straight marshal because we need all values to be strings
	tokenKey  = "consumptionToken"
	nodeKey   = "entitledNodes"
	expiryKey = "expiry"
	awsCSP    = "aws"
)

type licenseCheckoutInfo struct {
	ConsumptionToken string
	EntitledLicenses int
	Expiry           time.Time
}

func (m *AWS) start(ctx context.Context, errs chan<- error) {
	for range ticker(ctx, managerInterval) {
		err := m.runComplianceCheck(ctx)
		if err != nil {
			updError := m.updateAdapterOutput(false, fmt.Sprintf("unable to run compliance check with error: %v", err), "unable to run adapter, check adapter logs")
			if updError != nil {
				errs <- err
			}
			errs <- err
		}
	}
	logrus.Infof("[manager] exiting")
}

// runComplianceCheck compares the number of nodes registered with rancher with the number of entitlements currently
// held * nodesPerLicense. If we are not at the desired value, it checks in currently held entitlements and attempts
// to check out the right amount. If we are and our tokens are about to expire, it extends the checkout period. If
// any part of this fatally fails, the process will return an error
func (m *AWS) runComplianceCheck(ctx context.Context) error {
	license, err := m.aws.GetRancherLicense(ctx)
	if err != nil {
		// TODO: No license is an error state
		return fmt.Errorf("unable to get rancher license, err: %v", err)
	}
	nodeCounts, err := m.scraper.ScrapeAndParse()
	if err != nil {
		return fmt.Errorf("unable to determine number of active nodes: %v", err)
	}
	logrus.Debugf("found %d nodes from rancher metrics", nodeCounts.Total)
	currentCheckoutInfo, err := m.getLicenseCheckoutInfo()
	if err != nil {
		// not a breaking error, just means that we need to assume we have no registered entitlements
		logrus.Warnf("unable to get current license consumption info, will start fresh %v", err)
		currentCheckoutInfo = &licenseCheckoutInfo{
			EntitledLicenses: 0,
			ConsumptionToken: "",
		}
	}
	requiredLicenses := int(math.Ceil(float64(nodeCounts.Total) / float64(nodesPerLicense)))
	logrus.Debugf("have %d licenses checked out, need %d licenses", currentCheckoutInfo.EntitledLicenses, requiredLicenses)
	if currentCheckoutInfo.EntitledLicenses != requiredLicenses {
		// if we know we need a new set of entitlements, checkin what we are currently using since we only hold one
		// checked out set of entitlements at a time
		if currentCheckoutInfo.ConsumptionToken != "" {
			_, err = m.aws.CheckInRancherLicense(ctx, currentCheckoutInfo.ConsumptionToken)
			if err != nil {
				logrus.Warnf("unable to checkin license with error %v", err)
			} else {
				logrus.Debugf("successfully checked in license")
				currentCheckoutInfo.EntitledLicenses = 0
				currentCheckoutInfo.ConsumptionToken = ""
			}
		}
		availableLicenses, err := m.aws.GetNumberOfAvailableEntitlements(ctx, *license)
		logrus.Debugf("found %d entitlements available", availableLicenses)
		if err != nil {
			logrus.Warnf("unable to determine number of available entitlements, will attempt full checkout %v", err)
			// if we can't verify how many licenses are available, assume that we have enough to meet our requirements
			availableLicenses = requiredLicenses
		}
		checkoutAmount := requiredLicenses
		if checkoutAmount > availableLicenses {
			// only checkout what we actually have available to us
			checkoutAmount = availableLicenses
		}
		resp, err := m.aws.CheckoutRancherLicense(ctx, *license, checkoutAmount)
		if err != nil {
			return fmt.Errorf("unable to checkout rancher licenses %v", err)
		}
		logrus.Debugf("successfully checked out license")
		currentCheckoutInfo.ConsumptionToken = *resp.LicenseConsumptionToken
		currentCheckoutInfo.EntitledLicenses = checkoutAmount
		currentCheckoutInfo.Expiry = parseExpirationTimestamp(*resp.Expiration)
	} else {
		newCheckoutInfo, err := m.extendCheckout(ctx, 5*managerInterval, currentCheckoutInfo)
		if err != nil {
			currentCheckoutInfo.EntitledLicenses = 0
			currentCheckoutInfo.ConsumptionToken = ""
			logrus.Warnf("unable to extend license checkout, will assume it failed and reset: %v", err)
		} else {
			currentCheckoutInfo = newCheckoutInfo
		}
	}
	err = m.saveCheckoutInfo(currentCheckoutInfo)
	if err != nil {
		logrus.Warnf("unable to save current checkout info, next run may fail with checkout/checkin")
	}

	var statusMessage string
	if currentCheckoutInfo.EntitledLicenses == requiredLicenses {
		statusMessage = "Rancher server has the required amount of licenses"
	} else {
		statusMessage = fmt.Sprintf("server is not in compliance, wanted %d, but got %d", requiredLicenses, currentCheckoutInfo.EntitledLicenses)
	}

	return m.updateAdapterOutput(currentCheckoutInfo.EntitledLicenses == requiredLicenses, statusMessage, statusMessage)
}

// extendCheckout extends the checkout of the licenses in info if info.Expiry is within minTimeTillExpiry
func (m *AWS) extendCheckout(ctx context.Context, minTimeTillExpiry time.Duration, info *licenseCheckoutInfo) (*licenseCheckoutInfo, error) {
	timeUntilExpiry := info.Expiry.Sub(time.Now())
	if timeUntilExpiry > minTimeTillExpiry { // no need to extend consumption token yet
		return info, nil
	}
	logrus.Debugf("extending consumption token")
	res, err := m.aws.ExtendRancherLicenseConsumptionToken(ctx, info.ConsumptionToken)
	if err != nil {
		return nil, err
	}
	return &licenseCheckoutInfo{
		ConsumptionToken: *res.LicenseConsumptionToken,
		Expiry:           parseExpirationTimestamp(*res.Expiration),
		EntitledLicenses: info.EntitledLicenses,
	}, nil
}

// getLicenseCheckoutInfo retrieves checkoutInfo from the cache in k8s - we cache to k8s to recover from pod restart
// returns an error if it couldn't parse every one of the values from the cache
func (m *AWS) getLicenseCheckoutInfo() (*licenseCheckoutInfo, error) {
	secret, err := m.k8s.GetConsumptionTokenSecret()
	if err != nil {
		return nil, err
	}
	token, tOk := secret.Data[tokenKey]
	licenses, lOk := secret.Data[nodeKey]
	expiry, eOk := secret.Data[expiryKey]
	if !(tOk && lOk && eOk) {
		// if we couldn't extract the token or node counts, we can't return accurate checkout info
		return nil, fmt.Errorf("couldn't license consumption info from secret")
	}
	numLicenses, err := strconv.Atoi(string(licenses))
	if err != nil {
		return nil, fmt.Errorf("unable to parse the number of nodes the license token is for %v", err)
	}
	expiryTime, err := time.Parse(time.RFC3339, string(expiry))
	if err != nil {
		return nil, fmt.Errorf("unable to parse the token's expiry time %v", err)
	}
	return &licenseCheckoutInfo{
		ConsumptionToken: string(token),
		EntitledLicenses: numLicenses,
		Expiry:           expiryTime,
	}, nil
}

// saveCheckoutInfo saves the checkoutInfo to the k8s cache. If this fails, returns an error
func (m *AWS) saveCheckoutInfo(info *licenseCheckoutInfo) error {
	return m.k8s.UpdateConsumptionTokenSecret(map[string]string{
		tokenKey:  info.ConsumptionToken,
		nodeKey:   fmt.Sprintf("%d", info.EntitledLicenses),
		expiryKey: info.Expiry.Format(time.RFC3339),
	})
}

// updateAdapterOutput uses the k8s client to update the status objects signaling compliance/non-compliance to other apps
func (m *AWS) updateAdapterOutput(inCompliance bool, configMessage string, notificationMessage string) error {
	config := GetDefaultSupportConfig(m.k8s)
	config.CSP = CSPInfo{
		Name:       awsCSP,
		AcctNumber: m.aws.AccountNumber(),
	}
	rancherVersion, err := m.k8s.GetRancherVersion()
	if err != nil {
		return fmt.Errorf("unable to get rancher version: %v", err)
	}
	config.Product = createProductString(rancherVersion)
	info := ComplianceInfo{
		Message: configMessage,
	}
	if inCompliance {
		info.Status = StatusInCompliance
	} else {
		info.Status = StatusNotInCompliance
	}
	config.Compliance = info
	err = m.k8s.UpdateUserNotification(inCompliance, notificationMessage)
	if err != nil {
		// don't bother marshalling the config if we can't report the error to the user
		return err
	}
	marshalled, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("unable to marshall config: %v", err)
	}
	return m.k8s.UpdateCSPConfigOutput(marshalled)
}

func ticker(ctx context.Context, duration time.Duration) <-chan time.Time {
	ticker := time.NewTicker(duration)
	go func() {
		<-ctx.Done()
		ticker.Stop()
	}()
	return ticker.C
}

// parseExpirationTimestamp parses the timestamp from aws into a time.Time object
func parseExpirationTimestamp(expirationTS string) time.Time {
	// timestamps from extendLicenseCheckout seem to be RFC3339. However, timestamps from checkoutLicense are of the
	// modified form. Optimize for the standardized case
	expiration, err := time.Parse(time.RFC3339, expirationTS)
	if err != nil {
		logrus.Warnf("unable to parse timestamp with rfc3339 format, falling back to modified format: %v", err)
		expiration, err = time.Parse(rfc3339NoTZ, expirationTS)
	}
	if err != nil {
		logrus.Warnf("couldn't parse license expiration time: %v, defaulting to 1 hour renewal", err)
		expiration = time.Now().Add(1 * time.Hour)
	}
	return expiration
}
