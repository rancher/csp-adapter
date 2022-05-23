package mocks

import (
	corev1 "k8s.io/api/core/v1"
	apierror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type MockK8sClient struct {
	CurrentSecretData          map[string]string
	CurrentSupportConfig       []byte
	CurrentNotificationMessage string
	RancherHostName            string
	RancherVersion             string
}

func NewMockK8sClient(secretData map[string]string) *MockK8sClient {
	return &MockK8sClient{
		CurrentSecretData:    secretData,
		CurrentSupportConfig: nil,
	}
}

func (m *MockK8sClient) GetConsumptionTokenSecret() (*corev1.Secret, error) {
	if m.CurrentSecretData != nil {
		binData := map[string][]byte{}
		for key, value := range m.CurrentSecretData {
			binData[key] = []byte(value)
		}
		return &corev1.Secret{
			StringData: m.CurrentSecretData,
			Data:       binData,
		}, nil
	}
	return nil, apierror.NewNotFound(schema.GroupResource{Group: "", Resource: "secret"}, "test-secret")
}

func (m *MockK8sClient) UpdateConsumptionTokenSecret(data map[string]string) error {
	// todo: mock error
	m.CurrentSecretData = data
	return nil
}

func (m *MockK8sClient) UpdateCSPConfigOutput(marshalledData []byte) error {
	//todo: mock error
	m.CurrentSupportConfig = marshalledData
	return nil
}

func (m *MockK8sClient) UpdateUserNotification(isInCompliance bool, message string) error {
	if !isInCompliance {
		m.CurrentNotificationMessage = message
	}
	return nil
}

func (m *MockK8sClient) GetRancherHostname() (string, error) {
	return m.RancherHostName, nil
}

func (m *MockK8sClient) GetRancherVersion() (string, error) {
	return m.RancherVersion, nil
}
