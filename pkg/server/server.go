package server

import (
	"context"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rancher/csp-adapter/pkg/clients/k8s"
	"github.com/rancher/csp-adapter/pkg/supportconfig"
	"github.com/rancher/dynamiclistener"
	"github.com/rancher/dynamiclistener/server"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
)

const (
	supportConfigPath = "/generate/supportconfig"
)

func ListenAndServe(ctx context.Context, cfg *rest.Config, csp, acctNum string) error {
	clients, err := k8s.New(ctx, cfg)
	if err != nil {
		return err
	}

	if err := setCertificateExpirationDays(); err != nil {
		logrus.Infof("[ListenAndServe] could not set certificate expiration days via environment variable: %v", err)
	}

	gen, err := supportconfig.NewGenerator(csp, acctNum)
	if err != nil {
		return err
	}

	router := mux.NewRouter()
	router.Handle(supportConfigPath, supportconfig.NewHandler(gen))

	return listenAndServe(ctx, clients, router)
}

func setCertificateExpirationDays() error {
	certExpirationDaysKey := "CATTLE_NEW_SIGNED_CERT_EXPIRATION_DAYS"
	if os.Getenv(certExpirationDaysKey) == "" {
		return os.Setenv(certExpirationDaysKey, "3650") // 10 years
	}
	return nil
}

const (
	namespace = "cattle-system"
	tlsName   = "csp-adapter.cattle-system.svc"
	certName  = "cattle-csp-adapter-tls"
	caName    = "cattle-csp-adapter-ca"
)

func listenAndServe(ctx context.Context, clients *k8s.Clients, handler http.Handler) (rErr error) {
	return server.ListenAndServe(ctx, 9443, 0, handler, &server.ListenOpts{
		Secrets:       clients.Secrets,
		CertNamespace: namespace,
		CertName:      certName,
		CAName:        caName,
		TLSListenerConfig: dynamiclistener.Config{
			SANs: []string{
				tlsName,
			},
			FilterCN: dynamiclistener.OnlyAllow(tlsName),
		},
	})
}
