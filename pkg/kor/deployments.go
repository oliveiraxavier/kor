package kor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/argoproj/argo-rollouts/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/yonahd/kor/pkg/common"
	"github.com/yonahd/kor/pkg/filters"
)

func processNamespaceDeployments(clientsetinterface ClientInterface, namespace string, filterOpts *filters.Options) ([]ResourceInfo, error) {
	clientset := clientsetinterface.GetKubernetesClient()
	clientsetrollout := clientsetinterface.GetArgoRolloutsClient()
	deploymentsList, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: filterOpts.IncludeLabels})
	if err != nil {
		return nil, err
	}

	var deploymentsWithoutReplicas []ResourceInfo

	for _, deployment := range deploymentsList.Items {
		if pass, _ := filter.SetObject(&deployment).Run(filterOpts); pass {
			continue
		}

		if deployment.Labels["kor/used"] == "false" {
			reason := "Marked with unused label"

			rolloutWithSameNameDeployment, _ := RetrieveArgoRolloutsWithDeploymentWithSameName(clientsetrollout, deployment.Name, namespace)
			if rolloutWithSameNameDeployment == "" {
				deploymentsWithoutReplicas = append(deploymentsWithoutReplicas, ResourceInfo{Name: deployment.Name, Reason: reason})
				continue
			}
		}

		if *deployment.Spec.Replicas == 0 {
			rolloutWithSameNameDeployment, _ := RetrieveArgoRolloutsWithDeploymentWithSameName(clientsetrollout, deployment.Name, namespace)
			if rolloutWithSameNameDeployment == "" {
				reason := "Deployment has no replicas"
				deploymentsWithoutReplicas = append(deploymentsWithoutReplicas, ResourceInfo{Name: deployment.Name, Reason: reason})
			}
		}
	}

	return deploymentsWithoutReplicas, nil
}

func GetUnusedDeployments(filterOpts *filters.Options, clientsetinterface ClientInterface, outputFormat string, opts common.Opts) (string, error) {
	clientset := clientsetinterface.GetKubernetesClient()
	resources := make(map[string]map[string][]ResourceInfo)
	for _, namespace := range filterOpts.Namespaces(clientset) {
		diff, err := processNamespaceDeployments(clientsetinterface, namespace, filterOpts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to process namespace %s: %v\n", namespace, err)
			continue
		}
		if opts.DeleteFlag {
			if diff, err = DeleteResource(diff, clientset, namespace, "Deployment", opts.NoInteractive); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to delete Deployment %s in namespace %s: %v\n", diff, namespace, err)
			}
		}
		switch opts.GroupBy {
		case "namespace":
			resources[namespace] = make(map[string][]ResourceInfo)
			resources[namespace]["Deployment"] = diff
		case "resource":
			appendResources(resources, "Deployment", namespace, diff)
		}
	}

	var outputBuffer bytes.Buffer
	var jsonResponse []byte
	switch outputFormat {
	case "table":
		outputBuffer = FormatOutput(resources, opts)
	case "json", "yaml":
		var err error
		if jsonResponse, err = json.MarshalIndent(resources, "", "  "); err != nil {
			return "", err
		}
	}

	unusedDeployments, err := unusedResourceFormatter(outputFormat, outputBuffer, opts, jsonResponse)
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}

	return unusedDeployments, nil
}

func RetrieveArgoRolloutsWithDeploymentWithSameName(clientset versioned.Interface, rolloutname string, namespace string) (string, error) {
	rollout, err := clientset.ArgoprojV1alpha1().Rollouts(namespace).Get(context.TODO(), rolloutname, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	return rollout.GetName(), nil
}
