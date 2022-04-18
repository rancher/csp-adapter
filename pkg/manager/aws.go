package manager

import (
	"context"

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
	ctx, m.cancel = context.WithCancel(ctx)
	go m.start(ctx, errs)
}

func (m *AWS) Stop() {
	if m.cancel == nil {
		return
	}

	m.cancel()
}

func (m *AWS) start(ctx context.Context, errs chan<- error) {
	for {
		select {
		case <-ctx.Done():
			logrus.Infof("context cancelled, manager exiting")
			return
		}
	}
}
