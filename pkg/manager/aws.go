package manager

import (
	"context"
	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rancher/csp-adapter/pkg/clients/aws"
	"github.com/rancher/csp-adapter/pkg/clients/k8s"
	"github.com/rancher/csp-adapter/pkg/metrics"
	"github.com/sirupsen/logrus"
	apiError "k8s.io/apimachinery/pkg/api/errors"
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

const (
	cspStatusName    = "aws-csp-status"
	cspComponentName = "aws-csp"
)

func (m *AWS) updateComplianceStatus(newStatus bool, newReason string) error {
	currentStatus, err := m.k8s.RancherStatus.Get(cspStatusName, metav1.GetOptions{})
	if err != nil {
		// if we don't have the status yet, the csp-adapter will need to create it
		if apiError.IsNotFound(err) {
			statusToCreate := &v3.RancherStatus{
				ComponentStatus: newStatus,
				Reason:          newReason,
				ComponentName:   cspComponentName,
				ObjectMeta: metav1.ObjectMeta{
					Name: cspStatusName,
				},
			}
			statusToCreate, err = m.k8s.RancherStatus.Create(statusToCreate)
			if err != nil {
				return err
			}
		}
		return err
	}
	// if we already have the component, just update the status/reason to be consistent
	updatedStatus := currentStatus.DeepCopy()
	updatedStatus.ComponentStatus = newStatus
	updatedStatus.Reason = newReason
	updatedStatus, err = m.k8s.RancherStatus.Update(updatedStatus)
	if err != nil {
		return err
	}
	return nil
}
