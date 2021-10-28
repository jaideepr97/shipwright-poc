package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	uuid "github.com/nu7hatch/gouuid"
	shipwright "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	buildClient "github.com/shipwright-io/build/pkg/client/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	kubeConfigPath       = filepath.Join(homedir.HomeDir(), ".kube", "config")
	quayUsername         = "sbose78"
	imageRepo            = "generated"
	secretName           = "my-docker-credentials"
	contextDir           string
	imageRegistryServer  = "docker.io"
	serverPort           = 8085
	buildSystemNamespace = "shipwright-tenant"
)

func formHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	repoURL := r.FormValue("repo-url")
	contextDir = r.FormValue("context-dir")

	config, _ := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	buildClient, _ := buildClient.NewForConfig(config)

	buildRequestID, _ := uuid.NewV4()
	fmt.Println("found build")

	if _, err := buildClient.ShipwrightV1alpha1().Builds(buildSystemNamespace).Get(context.TODO(), fmt.Sprintf("%s", buildRequestID.String()), v1.GetOptions{}); err == nil {
		err := buildClient.ShipwrightV1alpha1().Builds(buildSystemNamespace).Delete(context.TODO(), fmt.Sprintf("%s", buildRequestID.String()), v1.DeleteOptions{})
		if err != nil {
			log.Fatal(err)
		}
	}
	buildObj := createBuild(buildRequestID.String(), repoURL, contextDir)
	_, err := buildClient.ShipwrightV1alpha1().Builds(buildSystemNamespace).Create(context.TODO(), buildObj, v1.CreateOptions{})
	if err != nil {
		log.Fatal(err)
	}

	if _, err := buildClient.ShipwrightV1alpha1().BuildRuns(buildSystemNamespace).Get(context.TODO(), fmt.Sprintf("%s", buildRequestID.String()), v1.GetOptions{}); err == nil {
		err := buildClient.ShipwrightV1alpha1().BuildRuns(buildSystemNamespace).Delete(context.TODO(), fmt.Sprintf("%s", buildRequestID.String()), v1.DeleteOptions{})
		if err != nil {
			log.Fatal(err)
		}
	}
	buildRunObj := createBuildRun(buildRequestID.String())
	_, err = buildClient.ShipwrightV1alpha1().BuildRuns(buildSystemNamespace).Create(context.TODO(), buildRunObj, v1.CreateOptions{})
	if err != nil {
		log.Fatal(err)
	}

	err = wait.Poll(time.Second*4, time.Second*180, func() (bool, error) {
		existingBuildRun, err := buildClient.ShipwrightV1alpha1().BuildRuns(buildSystemNamespace).Get(context.TODO(), buildRunObj.Name, v1.GetOptions{})
		if err != nil {
			// log.Fatalf("Error retrieving buildrun: %v", err)
			return false, err
		}
		for _, condition := range existingBuildRun.Status.Conditions {
			if condition.Type == shipwright.Succeeded && condition.Status == corev1.ConditionUnknown {
				fmt.Fprintf(w, "Building...")
				return false, nil
			} else if condition.Type == shipwright.Succeeded && condition.Status == corev1.ConditionFalse {
				fmt.Fprintf(w, "Build failed")
				return true, nil
			} else if condition.Type == shipwright.Succeeded && condition.Status == corev1.ConditionTrue {
				fmt.Fprintf(w, "Build successful!")
				break
			}
		}

		return true, nil
	})

}

func main() {

	fileServer := http.FileServer(http.Dir("./static"))
	http.Handle("/", fileServer)
	http.HandleFunc("/form", formHandler)

	fmt.Printf("Starting server at port %d\n", serverPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", serverPort), nil); err != nil {
		log.Fatal(err)
	}

}
