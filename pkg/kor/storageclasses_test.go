package kor

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/yonahd/kor/pkg/common"
	"github.com/yonahd/kor/pkg/filters"
)

func createTestStorageClass(t *testing.T) ClientInterface {
	clientsetinterface := SetConfigsForTests(t)
	clientset := clientsetinterface.GetKubernetesClient()

	sc1 := CreateTestStorageClass("test-sc1", "kor.com")
	_, err := clientset.StorageV1().StorageClasses().Create(context.TODO(), sc1, v1.CreateOptions{})
	if err != nil {
		t.Fatalf("Error creating fake %s: %v", "StorageClass", err)
	}

	return clientsetinterface
}

func TestRetrieveUsedStorageClassesFromPVCs(t *testing.T) {
	clientsetinterface := createTestPvcs(t)
	clientset := clientsetinterface.GetKubernetesClient()
	usedStorageClasses, err := retrieveUsedStorageClasses(clientset)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !contains(usedStorageClasses, "test-sc1") {
		t.Errorf("Expected 'test-sc1', got %v", usedStorageClasses)
	}
}

func TestRetrieveUsedStorageClassesFromPVs(t *testing.T) {
	clientsetinterface := createTestPvs(t)
	clientset := clientsetinterface.GetKubernetesClient()
	usedStorageClasses, err := retrieveUsedStorageClasses(clientset)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !contains(usedStorageClasses, "test-sc1") {
		t.Errorf("Expected 'test-sc1', got %v", usedStorageClasses)
	}
}

func TestProcessStorageClasses(t *testing.T) {
	clientsetinterface := createTestStorageClass(t)
	clientset := clientsetinterface.GetKubernetesClient()
	unusedStorageClasses, err := processStorageClasses(clientset, &filters.Options{})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(unusedStorageClasses) != 1 {
		t.Errorf("Expected 1 used StorageClasses, got %d", len(unusedStorageClasses))
	}

	if unusedStorageClasses[0].Name != "test-sc1" {
		t.Errorf("Expected 'test-sc1', got %s", unusedStorageClasses[0])
	}
}

func TestGetUnusedStorageClassesStructured(t *testing.T) {
	clientsetinterface := createTestStorageClass(t)

	opts := common.Opts{
		WebhookURL:    "",
		Channel:       "",
		Token:         "",
		DeleteFlag:    false,
		NoInteractive: true,
		GroupBy:       "namespace",
	}

	output, err := GetUnusedStorageClasses(&filters.Options{}, clientsetinterface, "json", opts)
	if err != nil {
		t.Fatalf("Error calling GetUnusedStorageClasses: %v", err)
	}

	expectedOutput := map[string]map[string][]string{
		"": {
			"StorageClass": {"test-sc1"},
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
