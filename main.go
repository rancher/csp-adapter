package main

import (
	"fmt"
	"os"

	"github.com/rancher/csp-adapter/pkg/clients/aws"
	"github.com/rancher/csp-adapter/pkg/clients/k8s"
	"github.com/rancher/csp-adapter/pkg/manager"
	"github.com/rancher/csp-adapter/pkg/metrics"
	"github.com/rancher/csp-adapter/pkg/server"
	"github.com/rancher/csp-adapter/pkg/supportconfig"
	"github.com/rancher/wrangler/pkg/k8scheck"
	"github.com/rancher/wrangler/pkg/kubeconfig"
	"github.com/rancher/wrangler/pkg/ratelimit"
	"github.com/rancher/wrangler/pkg/signals"
	"github.com/sirupsen/logrus"
)

var (
	Version   = "dev"
	GitCommit = "HEAD"
)

func main() {
	if err := run(); err != nil {
		logrus.Fatal(err)
	}
}

const (
	debugEnv           = "CATTLE_DEBUG"
	rancherHostnameEnv = "RANCHER_HOSTNAME" // this should come from the server-url setting in future
)

const (
	csp = "aws" // hardcoded for now, AWS only
)

func run() error {
	if os.Getenv(debugEnv) == "true" {
		logrus.SetLevel(logrus.DebugLevel)
	}

	rancherHostname := os.Getenv(rancherHostnameEnv)
	if rancherHostname == "" {
		return fmt.Errorf("%s must be specified in env", rancherHostnameEnv)
	}

	logrus.Infof("csp-adapter version %s is starting", fmt.Sprintf("%s (%s)", Version, GitCommit))

	cfg, err := kubeconfig.GetNonInteractiveClientConfig(os.Getenv("KUBECONFIG")).ClientConfig()
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

	awsClient, err := aws.NewClient(ctx)
	if err != nil {
		return err
	}

	generator, err := supportconfig.NewGenerator(csp, "") //aws.AccountNumber())
	if err != nil {
		return err
	}

	if err := server.ListenAndServe(ctx, cfg, generator); err != nil {
		return err
	}

	m := manager.NewAWS(awsClient, k8sClients, metrics.NewScraper(rancherHostname))

	errs := make(chan error, 1)
	m.Start(ctx, errs)
	defer m.Stop()

	go func() {
		for err := range errs {
			logrus.Errorf("aws manager error: %v", err)
		}
	}()

	<-ctx.Done()

	return nil
}
