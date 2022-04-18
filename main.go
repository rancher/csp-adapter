package main

import (
	"fmt"
	"os"

	"github.com/rancher/csp-adapter/pkg/server"
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
	debugEnv   = "CATTLE_DEBUG"
	acctNumEnv = "CSP_ACCOUNT_NUMBER"
	cspEnv     = "CSP_NAME"
)

func run() error {
	if os.Getenv(debugEnv) == "true" {
		logrus.SetLevel(logrus.DebugLevel)
	}

	acctNum := os.Getenv(acctNumEnv)
	csp := os.Getenv(cspEnv)
	if acctNum == "" || csp == "" {
		logrus.Fatalf("env vars: %s, %s must be set", cspEnv, acctNumEnv)
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

	if err := server.ListenAndServe(ctx, cfg, csp, acctNum); err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}
