package manager

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/licensemanager/types"
	"github.com/rancher/csp-adapter/pkg/clients/aws"
	"github.com/rancher/csp-adapter/pkg/clients/k8s"
	"github.com/rancher/csp-adapter/pkg/metrics"
	"github.com/sirupsen/logrus"
)

type AWS struct {
	cancel  context.CancelFunc
	aws     aws.Client
	k8s     *k8s.Clients
	scraper metrics.Scraper
}

func NewAWS(a aws.Client, k *k8s.Clients, s metrics.Scraper) *AWS {
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
)

type checkOutInfo struct {
	LicenseConsumptionToken string
	Expiration              time.Time
}

func (m *AWS) start(ctx context.Context, errs chan<- error) {
	// checkedOutLicenses maps <license-arn> -> <consumption-token>
	checkedOutLicenses := make(map[string]*checkOutInfo, 0)
	for range ticker(ctx, managerInterval) {
		licenses, err := m.aws.ListRancherReceivedLicenses(ctx)
		if err != nil {
			errs <- err
			continue
		}
		logrus.Infof("[manager] found %v rancher received licenses", len(licenses))

		nodeCounts, err := m.scraper.ScrapeAndParse()
		if err != nil {
			errs <- err
			continue
		}
		logrus.Debugf("[manager] rancher setup has %v nodes", nodeCounts.Total)

		// current is the number of currently checked out licenses
		current := len(checkedOutLicenses)
		// required is the number of licenses required based on the current node usage
		required := int(math.Ceil(float64(nodeCounts.Total / nodesPerLicense)))
		if current < required {
			underLicensedAmt := required - current
			logrus.Debugf("[manager] OutOfCompliance: under-licensed, checking out an additional %v licenses", underLicensedAmt)

			availableLicenses := m.getAvailableLicenses(licenses, checkedOutLicenses)
			if underLicensedAmt > len(availableLicenses) {
				err := errors.New("OutOfCompliance: setup is over provisioned for available licenses, reduce nodes or purchase additional licenses")
				errs <- err
				continue
			}

			// checkout additional required licenses
			for i := 0; i < underLicensedAmt; i++ {
				res, err := m.aws.CheckoutRancherLicense(ctx, availableLicenses[i])
				if err != nil {
					errs <- err
					continue
				}

				logrus.Debugf("[manager] successfully checked out license %v", *res.LicenseArn)
				expiration, err := time.Parse(time.RFC3339, *res.Expiration)
				if err != nil {
					errs <- fmt.Errorf("couldn't parse license expiration time: %v", err)
					expiration = time.Now().Add(1 * time.Hour)
				}
				checkedOutLicenses[*res.LicenseArn] = &checkOutInfo{
					LicenseConsumptionToken: *res.LicenseConsumptionToken,
					Expiration:              expiration,
				}
			}
		} else if current > required {
			overLicensedAmt := current - required
			logrus.Debugf("[manager] InCompliance: over-licensed, checking in %v licenses", overLicensedAmt)

			var checkedOut []string
			for k := range checkedOutLicenses {
				checkedOut = append(checkedOut, k)
			}

			for i := 0; i < overLicensedAmt; i++ {
				key := checkedOut[i]
				_, err := m.aws.CheckInRancherLicense(ctx, checkedOutLicenses[key].LicenseConsumptionToken)
				if err != nil {
					errs <- err
					continue
				}
				delete(checkedOutLicenses, key)
			}
		} else { // current == required, noop
			logrus.Debugf("[manager] InCompliance: (checked-out=%v) == (required=%v)", current, required)
		}

		// check whether consumption tokens need to be extended for any currently checked out licenses
		for l, info := range checkedOutLicenses {
			timeUntilExpiry := info.Expiration.Sub(time.Now())
			logrus.Debugf("[manager] license %s expires in %s", l, timeUntilExpiry.String())
			if timeUntilExpiry > (5 * managerInterval) { // no need to extend consumption token yet
				continue
			}

			logrus.Debugf("[manager] extending consumption token for license %s", l)
			res, err := m.aws.ExtendRancherLicenseConsumptionToken(ctx, info.LicenseConsumptionToken)
			if err != nil {
				errs <- err
				continue
			}
			expiration, err := time.Parse(time.RFC3339, *res.Expiration)
			if err != nil {
				errs <- fmt.Errorf("couldn't parse license expiration time: %v", err)
				expiration = time.Now().Add(1 * time.Hour)
			}
			info.Expiration = expiration
			info.LicenseConsumptionToken = *res.LicenseConsumptionToken
		}
	}

	logrus.Infof("[manager] exiting")
}

func (m *AWS) getAvailableLicenses(licenses []types.GrantedLicense, checkedOutLicenses map[string]*checkOutInfo) []types.GrantedLicense {
	var availableLicenses []types.GrantedLicense
	for _, license := range licenses {
		if _, ok := checkedOutLicenses[*license.LicenseArn]; !ok && license.Status == types.LicenseStatusAvailable {
			availableLicenses = append(availableLicenses, license)
		}
	}
	return availableLicenses
}

func ticker(ctx context.Context, duration time.Duration) <-chan time.Time {
	ticker := time.NewTicker(duration)
	go func() {
		<-ctx.Done()
		ticker.Stop()
	}()
	return ticker.C
}
