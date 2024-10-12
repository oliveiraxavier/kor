package kor

import (
	"testing"

	"github.com/argoproj/argo-rollouts/pkg/client/clientset/versioned"
	fakeargorollouts "github.com/argoproj/argo-rollouts/pkg/client/clientset/versioned/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

type FakeClientSet struct {
	coreClient             *fake.Clientset
	coreClientArgoRollouts *fakeargorollouts.Clientset
}

// GetArgoRolloutsClient implements ClientInterface.
func (c *FakeClientSet) GetArgoRolloutsClient() versioned.Interface {
	return c.coreClientArgoRollouts
}

// GetKubernetesClient implements ClientInterface.
func (c *FakeClientSet) GetKubernetesClient() kubernetes.Interface {
	return c.coreClient
}

func NewFakeClientSet(t *testing.T) (ClientInterface, error) {
	coreClient := fake.NewSimpleClientset()
	coreClientArgoRollouts := fakeargorollouts.NewSimpleClientset()

	// Return the ClientSet struct
	return &FakeClientSet{
		coreClient:             coreClient,
		coreClientArgoRollouts: coreClientArgoRollouts,
	}, nil
}

func GetFakeKubeClient(t *testing.T) (ClientInterface, error) {
	clientsetinterface, err := NewFakeClientSet(t)

	if err != nil {
		t.Fatalf("Error creating fake clientset. Error: %v", err)
	}

	return clientsetinterface, nil
}

func SetConfigsForTests(t *testing.T) ClientInterface {
	clientsetinterface, err := GetFakeKubeClient(t)
	if err != nil {
		t.Fatalf("Error on setting config: %v", err)
	}

	return clientsetinterface
}

// func setupTest(tb testing.TB, t *testing.T) (ClientInterface, kubernetes.Interface, versioned.Interface) {
// 	clientsetinterface := SetConfigsForTests(t)
// 	clientset := clientsetinterface.GetKubernetesClient()
// 	clientsetargorollouts := clientsetinterface.GetArgoRolloutsClient()
// 	return clientsetinterface, clientset, clientsetargorollouts
// }
