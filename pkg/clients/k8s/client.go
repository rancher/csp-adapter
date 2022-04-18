package k8s

import (
	"context"

	"github.com/rancher/wrangler/pkg/clients"
	v1 "github.com/rancher/wrangler/pkg/generated/controllers/core/v1"
	"k8s.io/client-go/rest"
)

type Clients struct {
	ConfigMaps v1.ConfigMapClient
	Secrets    v1.SecretController
}

func New(ctx context.Context, rest *rest.Config) (*Clients, error) {
	clients, err := clients.NewFromConfig(rest, nil)
	if err != nil {
		return nil, err
	}

	if err := clients.Start(ctx); err != nil {
		return nil, err
	}

	return &Clients{
		ConfigMaps: clients.Core.ConfigMap(),
		Secrets:    clients.Core.Secret(),
	}, nil
}
