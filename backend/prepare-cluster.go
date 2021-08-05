package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

var dec = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

var manifestURLs = []string{"https://storage.googleapis.com/tekton-releases/pipeline/previous/v0.25.0/release.yaml",
	"https://github.com/shipwright-io/build/releases/download/v0.5.1/release.yaml",
	"https://github.com/shipwright-io/build/releases/download/nightly/default_strategies.yaml"}

func createk8sResourceOnCluster(mapper *restmapper.DeferredDiscoveryRESTMapper, dynamicClient dynamic.Interface, resourceManifestBytes []byte) error {
	// reference: https://ymmt2005.hatenablog.com/entry/2020/04/14/An_example_of_using_dynamic_client_of_k8s.io/client-go

	// Decode YAML manifest into unstructured.Unstructured
	unstructuredObj := &unstructured.Unstructured{}
	_, gvk, err := dec.Decode([]byte(resourceManifestBytes), nil, unstructuredObj)
	if err != nil {
		if err.Error() == "Object 'Kind' is missing in 'null'" {
			return nil
		}
		return fmt.Errorf("Error decoding manifest: %v", err)
	}

	// Find GVR
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return fmt.Errorf("Error finding GVR: %v", err)
	}

	// Obtain REST interface for the GVR
	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// namespaced resources should specify the namespace
		dr = dynamicClient.Resource(mapping.Resource).Namespace(unstructuredObj.GetNamespace())
	} else {
		// for cluster-wide resources
		dr = dynamicClient.Resource(mapping.Resource)
	}

	// Marshal object into JSON
	data, err := json.Marshal(unstructuredObj)
	if err != nil {
		return err
	}

	// Create or Update the resource
	_, err = dr.Patch(context.Background(), unstructuredObj.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{
		FieldManager: "sample-controller",
	})

	return err
}

func applyManifestsToCluster(manifestYAMLURLs []string, restConfig *rest.Config) error {

	discoveryClient, err := getDiscoveryClient(restConfig)
	if err != nil {
		return fmt.Errorf("could not get discovery client: %v", err)
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(discoveryClient))

	dynamicClient, err := getDynamicClient(restConfig)
	if err != nil {
		return fmt.Errorf("could not get dynamic client: %v", err)
	}

	for _, manifestYAMLURL := range manifestYAMLURLs {
		manifestYAML, err := http.Get(manifestYAMLURL)
		time.Sleep(3 * time.Second)
		if err != nil {
			return fmt.Errorf("Unable to retrieve manifest from URL: %v", err)
		}
		defer manifestYAML.Body.Close()

		manifestYAMLBytes, err := ioutil.ReadAll(manifestYAML.Body)
		if err != nil {
			return fmt.Errorf("Unable to retrieve manifest bytes: %v", err)
		}

		YAMLResourcesBytes, err := SplitYAML(manifestYAMLBytes)
		if err != nil {
			return fmt.Errorf("Unable to split YAML into resources: %v", err)
		}

		for _, YAMLResourceBytes := range YAMLResourcesBytes {
			err := createk8sResourceOnCluster(mapper, dynamicClient, YAMLResourceBytes)
			if err != nil {
				return fmt.Errorf("Error creating resource on cluster from manifest: %v", err)
			}
		}
	}

	return nil

}

// func main() {
// 	config, _ := clientcmd.BuildConfigFromFlags("", "/home/jrao/.kube/config")
// 	err := applyManifestsToCluster(manifestURLs, config)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }
