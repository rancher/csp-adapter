package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rancher/csp-adapter/pkg/clients/aws"
	"github.com/rancher/csp-adapter/pkg/clients/k8s"
	"github.com/rancher/csp-adapter/pkg/manager"
	"github.com/rancher/csp-adapter/pkg/metrics"
	"github.com/rancher/wrangler/pkg/k8scheck"
	"github.com/rancher/wrangler/pkg/ratelimit"
	"github.com/rancher/wrangler/pkg/signals"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
)

var (
	Version   = "dev"
	GitCommit = "HEAD"
)

func main() {
	if err := run(); err != nil {
		logrus.Fatalf("csp-adapter failed to run with error: %v", err)
	}
}

const (
	debugEnv   = "CATTLE_DEBUG"
	devModeEnv = "CATTLE_DEV_MODE"
	awsCSP     = "aws"
)

func run() error {
	if os.Getenv(debugEnv) == "true" {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.Infof("csp-adapter version %s is starting", fmt.Sprintf("%s (%s)", Version, GitCommit))

	cfg, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	cfg.RateLimiter = ratelimit.None
	ctx := signals.SetupSignalContext()
	err = k8scheck.Wait(ctx, *cfg)
	if err != nil {
		return err
	}

	k8sClients, err := k8s.New(ctx, cfg)
	if err != nil {
		return err
	}

	devMode := os.Getenv(devModeEnv) == "true"

	awsClient, err := aws.NewClient(ctx, devMode)
	if err != nil {
		registerErr := registerStartupError(k8sClients, createCSPInfo(awsCSP, "unknown"), err)
		if registerErr != nil {
			return fmt.Errorf("unable to start or register manager error, start error: %v, register error: %v", err, registerErr)
		}
		return fmt.Errorf("failed to start, unable to start aws client: %v", err)
	}

	hostname, err := k8sClients.GetRancherHostname()
	if err != nil {
		registerErr := registerStartupError(k8sClients, createCSPInfo(awsCSP, awsClient.AccountNumber()), err)
		if registerErr != nil {
			return fmt.Errorf("unable to start or register manager error, start error: %v, register error: %v", err, registerErr)
		}
		return fmt.Errorf("failed to start, unable to get hostname: %v", err)
	}

	m := manager.NewAWS(awsClient, k8sClients, metrics.NewScraper(hostname, cfg))

	errs := make(chan error, 1)
	m.Start(ctx, errs)
	go func() {
		for err := range errs {
			logrus.Errorf("aws manager error: %v", err)
		}
	}()

	<-ctx.Done()

	return nil
}

// createCSPInfo creates a manager.CSPInfo from a provided csp name and account number
func createCSPInfo(csp, acctNumber string) manager.CSPInfo {
	return manager.CSPInfo{
		Name:       csp,
		AcctNumber: acctNumber,
	}
}

// registerStartupError registers that an error occurred when starting the manager for the cloud account represented by
// cspInfo if we could start our k8s clients but couldn't init some other part of the manager infra, we need to
// report this to the user and save the error so it can be included in the supportconfig bundle
func registerStartupError(clients *k8s.Clients, cspInfo manager.CSPInfo, startupErr error) error {
	defaultConfig := manager.GetDefaultSupportConfig(clients)
	defaultConfig.Compliance = manager.ComplianceInfo{
		Status:  manager.StatusNotInCompliance,
		Message: fmt.Sprintf("CSP adapter unable to start due to error: %v", startupErr),
	}
	defaultConfig.CSP = cspInfo
	marshalledConfig, err := json.Marshal(defaultConfig)
	if err != nil {
		return err
	}
	err = clients.UpdateUserNotification(false, "Marketplace Adapter: unable to start csp adapter, check adapter logs")
	if err != nil {
		return err
	}
	err = clients.UpdateCSPConfigOutput(marshalledConfig)
	return err
}
