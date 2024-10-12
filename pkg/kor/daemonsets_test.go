package kor

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/yonahd/kor/pkg/common"
	"github.com/yonahd/kor/pkg/filters"
)

func createTestDaemonSets(t *testing.T) ClientInterface {
	clientsetinterface, err := NewFakeClientSet(t)
	if err != nil {
		t.Fatalf("Error creating client set. Error: %v", err)
	}
	clientset := clientsetinterface.GetKubernetesClient()

	_, err = clientset.CoreV1().Namespaces().Create(context.TODO(), &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{Name: testNamespace},
	}, v1.CreateOptions{})

	if err != nil {
		t.Fatalf("Error creating namespace %s: %v", testNamespace, err)
	}

	ds1 := CreateTestDaemonSet(testNamespace, "test-ds1", AppLabels, &appsv1.DaemonSetStatus{
		CurrentNumberScheduled: 0,
	})
	_, err = clientset.AppsV1().DaemonSets(testNamespace).Create(context.TODO(), ds1, v1.CreateOptions{})
	if err != nil {
		t.Fatalf("Error creating fake %s: %v", "DaemonSet", err)
	}

	ds2 := CreateTestDaemonSet(testNamespace, "test-ds2", AppLabels, &appsv1.DaemonSetStatus{
		CurrentNumberScheduled: 1,
	})
	_, err = clientset.AppsV1().DaemonSets(testNamespace).Create(context.TODO(), ds2, v1.CreateOptions{})
	if err != nil {
		t.Fatalf("Error creating fake %s: %v", "DaemonSet", err)
	}

	ds3 := CreateTestDaemonSet(testNamespace, "test-ds3", UsedLabels, &appsv1.DaemonSetStatus{
		CurrentNumberScheduled: 0,
	})
	_, err = clientset.AppsV1().DaemonSets(testNamespace).Create(context.TODO(), ds3, v1.CreateOptions{})
	if err != nil {
		t.Fatalf("Error creating fake %s: %v", "DaemonSet", err)
	}

	ds4 := CreateTestDaemonSet(testNamespace, "test-ds4", UnusedLabels, &appsv1.DaemonSetStatus{
		CurrentNumberScheduled: 1,
	})
	_, err = clientset.AppsV1().DaemonSets(testNamespace).Create(context.TODO(), ds4, v1.CreateOptions{})
	if err != nil {
		t.Fatalf("Error creating fake %s: %v", "DaemonSet", err)
	}

	return clientsetinterface
}

func TestProcessNamespaceDaemonSets(t *testing.T) {
	clientsetinterface := createTestDaemonSets(t)
	clientset := clientsetinterface.GetKubernetesClient()

	daemonSetsWithoutReplicas, err := processNamespaceDaemonSets(clientset, testNamespace, &filters.Options{})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(daemonSetsWithoutReplicas) != 2 {
		t.Errorf("Expected 1 DaemonSet without replicas, got %d", len(daemonSetsWithoutReplicas))
	}

	if daemonSetsWithoutReplicas[0].Name != "test-ds1" && daemonSetsWithoutReplicas[1].Name != "test-ds4" {
		t.Errorf("Expected 'test-ds1', 'test-ds4', got %s, %s", daemonSetsWithoutReplicas[0], daemonSetsWithoutReplicas[1])
	}
}

func TestGetUnusedDaemonSetsStructured(t *testing.T) {
	clientsetinterface := createTestDaemonSets(t)

	opts := common.Opts{
		WebhookURL:    "",
		Channel:       "",
		Token:         "",
		DeleteFlag:    false,
		NoInteractive: true,
		GroupBy:       "namespace",
	}

	output, err := GetUnusedDaemonSets(&filters.Options{}, clientsetinterface, "json", opts)
	if err != nil {
		t.Fatalf("Error calling GetUnusedDaemonSetsStructured: %v", err)
	}

	expectedOutput := map[string]map[string][]string{
		testNamespace: {
			"DaemonSet": {
				"test-ds1",
				"test-ds4",
			},
		},
	}

	var actualOutput map[string]map[string][]string
	if err := json.Unmarshal([]byte(output), &actualOutput); err != nil {
		t.Fatalf("Error unmarshaling actual output: %v", err)
	}

	if !reflect.DeepEqual(expectedOutput, actualOutput) {
		t.Errorf("Expected output does not match actual output")
	}
}

func init() {
	scheme.Scheme = runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme.Scheme)
}
