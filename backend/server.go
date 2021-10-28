package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	uuid "github.com/nu7hatch/gouuid"
	buildClient "github.com/shipwright-io/build/pkg/client/clientset/versioned"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	kubeConfigPath        = filepath.Join(homedir.HomeDir(), ".kube", "config")
	quayUsername          = "sbose78"
	imageRepo             = "generated"
	secretName            = "my-docker-credentials"
	contextDir            string
	imageRegistryServer   = "docker.io"
	serverPort            = 8085
	buildSystemNamespace  = "shipwright-tenant"
	config                *rest.Config
	shipwrightBuildClient *buildClient.Clientset
)

func initializeClient() {
	var err error

	if shipwrightBuildClient == nil {
		if config == nil {
			config, err = rest.InClusterConfig()
			if os.Getenv("DEVMODE") == "true" {
				config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
			}
			if err != nil {
				panic(err.Error())
			}
		}
		shipwrightBuildClient, _ = buildClient.NewForConfig(config)
	}
}

func buildStatusHandler(w http.ResponseWriter, r *http.Request) {
	paramValues := r.URL.Query()
	buildID := paramValues.Get("id")

	initializeClient()

	existingBuildRun, err := shipwrightBuildClient.ShipwrightV1alpha1().BuildRuns(buildSystemNamespace).Get(context.TODO(), buildID, v1.GetOptions{})
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(existingBuildRun.Status)
	}
}

func formHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	repoURL := r.FormValue("repo-url")
	contextDir = r.FormValue("context-dir")

	initializeClient()

	buildRequestID, _ := uuid.NewV4()

	if _, err := shipwrightBuildClient.ShipwrightV1alpha1().Builds(buildSystemNamespace).Get(context.TODO(), fmt.Sprintf("%s", buildRequestID.String()), v1.GetOptions{}); err == nil {
		err := shipwrightBuildClient.ShipwrightV1alpha1().Builds(buildSystemNamespace).Delete(context.TODO(), fmt.Sprintf("%s", buildRequestID.String()), v1.DeleteOptions{})
		if err != nil {
			log.Fatal(err)
		}
	}
	buildObj := createBuild(buildRequestID.String(), repoURL, contextDir)
	_, err := shipwrightBuildClient.ShipwrightV1alpha1().Builds(buildSystemNamespace).Create(context.TODO(), buildObj, v1.CreateOptions{})
	if err != nil {
		log.Fatal(err)
	}

	if _, err := shipwrightBuildClient.ShipwrightV1alpha1().BuildRuns(buildSystemNamespace).Get(context.TODO(), fmt.Sprintf("%s", buildRequestID.String()), v1.GetOptions{}); err == nil {
		err := shipwrightBuildClient.ShipwrightV1alpha1().BuildRuns(buildSystemNamespace).Delete(context.TODO(), fmt.Sprintf("%s", buildRequestID.String()), v1.DeleteOptions{})
		if err != nil {
			log.Fatal(err)
		}
	}
	buildRunObj := createBuildRun(buildRequestID.String())
	_, err = shipwrightBuildClient.ShipwrightV1alpha1().BuildRuns(buildSystemNamespace).Create(context.TODO(), buildRunObj, v1.CreateOptions{})
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")

	// return the buildrun id so that it could be polled
	json.NewEncoder(w).Encode(buildRunObj.ObjectMeta)

}

func main() {

	fileServer := http.FileServer(http.Dir("./static"))
	http.Handle("/", fileServer)
	http.HandleFunc("/form", formHandler)
	http.HandleFunc("/buildstatus", buildStatusHandler)

	fmt.Printf("Starting server at port %d\n", serverPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", serverPort), nil); err != nil {
		log.Fatal(err)
	}

}
