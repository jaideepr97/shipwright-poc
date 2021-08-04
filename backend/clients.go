package main

import (
	"fmt"

	shipwrightClient "github.com/shipwright-io/build/pkg/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func getRestConfigFromBytes(kubeConfigBytes []byte) (*rest.Config, error) {
	kubeClientConfig, err := clientcmd.NewClientConfigFromBytes(kubeConfigBytes)
	if err != nil {
		return nil, fmt.Errorf("Unable to create client config from kubeconfig bytes, error: %v", err)
	}

	restConfig, err := kubeClientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("Unable to create rest config, error: %v", err)
	}

	return restConfig, nil
}

func getk8sClient(restConfig *rest.Config) (*kubernetes.Clientset, error) {
	k8sClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("Unable to create k8s client object, error: %v", err)
	}

	return k8sClient, nil
}

func getShipwrightClient(restConfig *rest.Config) (*shipwrightClient.Clientset, error) {
	shipwrightClient, err := shipwrightClient.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("Unable to create shipwright client object, error: %v", err)
	}

	return shipwrightClient, nil
}
