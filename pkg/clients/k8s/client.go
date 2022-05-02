package k8s

import (
	"context"

	"github.com/rancher/lasso/pkg/controller"
	"github.com/rancher/rancher/pkg/generated/controllers/management.cattle.io"
	mgmtv3 "github.com/rancher/rancher/pkg/generated/controllers/management.cattle.io/v3"
	"github.com/rancher/wrangler/pkg/clients"
	v1 "github.com/rancher/wrangler/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/generic"
	"k8s.io/apimachinery/pkg/runtime"
	authnv1 "k8s.io/client-go/kubernetes/typed/authentication/v1"
	authzv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/client-go/rest"
)

type Clients struct {
	ConfigMaps    v1.ConfigMapClient
	Secrets       v1.SecretController
	SAR           authzv1.SubjectAccessReviewInterface
	TR            authnv1.TokenReviewInterface
	RancherStatus mgmtv3.RancherStatusClient
	Settings      mgmtv3.SettingCache
}

func New(ctx context.Context, rest *rest.Config) (*Clients, error) {
	clients, err := clients.NewFromConfig(rest, nil)
	if err != nil {
		return nil, err
	}
	scheme := runtime.NewScheme()
	factory, err := controller.NewSharedControllerFactoryFromConfig(rest, scheme)
	if err != nil {
		return nil, err
	}
	opts := &generic.FactoryOptions{
		SharedControllerFactory: factory,
	}
	mgmt, err := management.NewFactoryFromConfigWithOptions(rest, opts)

	if err := clients.Start(ctx); err != nil {
		return nil, err
	}

	return &Clients{
		ConfigMaps:    clients.Core.ConfigMap(),
		Secrets:       clients.Core.Secret(),
		SAR:           clients.K8s.AuthorizationV1().SubjectAccessReviews(),
		TR:            clients.K8s.AuthenticationV1().TokenReviews(),
		RancherStatus: mgmt.Management().V3().RancherStatus(),
		Settings:      mgmt.Management().V3().Setting().Cache(),
	}, nil
}
