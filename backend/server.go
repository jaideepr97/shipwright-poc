package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	shipwright "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	buildClient "github.com/shipwright-io/build/pkg/client/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	kubeConfigPath      = filepath.Join(homedir.HomeDir(), ".kube", "config")
	dockerServer        = "https://index.docker.io/v1/"
	quayServer          = "quay.io"
	secretName          = "image-registry-secret"
	contextDir          string
	imageRegistryServer string
	serverPort          = 8085
)

func formHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	imageRegistry := r.FormValue("image-registry")
	username := r.FormValue("username")
	password := r.FormValue("password")
	email := r.FormValue("email")
	repoURL := r.FormValue("repo-url")
	contextDir = r.FormValue("context-dir")
	repoName := repoURL[strings.LastIndex(repoURL, "/")+1:]

	if imageRegistry == "docker" {
		imageRegistryServer = dockerServer
	} else {
		imageRegistryServer = quayServer
	}

	config, _ := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	k8sClient, _ := kubernetes.NewForConfig(config)
	buildClient, _ := buildClient.NewForConfig(config)

	if _, err := k8sClient.CoreV1().Secrets("default").Get(context.TODO(), secretName, v1.GetOptions{}); err == nil {
		err := k8sClient.CoreV1().Secrets("default").Delete(context.TODO(), secretName, v1.DeleteOptions{})
		if err != nil {
			log.Fatal(err)
		}
	}
	dockerSecret := createDockerSecret(username, password, email, imageRegistryServer)
	_, err := k8sClient.CoreV1().Secrets("default").Create(context.TODO(), dockerSecret, v1.CreateOptions{})
	if err != nil {
		log.Fatal(err)
	}

	if _, err := buildClient.ShipwrightV1alpha1().Builds("default").Get(context.TODO(), fmt.Sprintf("%v-build", repoName), v1.GetOptions{}); err == nil {
		err := buildClient.ShipwrightV1alpha1().Builds("default").Delete(context.TODO(), fmt.Sprintf("%v-build", repoName), v1.DeleteOptions{})
		if err != nil {
			log.Fatal(err)
		}
	}
	buildObj := createBuild(imageRegistry, repoURL, username, repoName, secretName, contextDir)
	_, err = buildClient.ShipwrightV1alpha1().Builds("default").Create(context.TODO(), buildObj, v1.CreateOptions{})
	if err != nil {
		log.Fatal(err)
	}

	if _, err := buildClient.ShipwrightV1alpha1().BuildRuns("default").Get(context.TODO(), fmt.Sprintf("%v-buildrun", repoName), v1.GetOptions{}); err == nil {
		err := buildClient.ShipwrightV1alpha1().BuildRuns("default").Delete(context.TODO(), fmt.Sprintf("%v-buildrun", repoName), v1.DeleteOptions{})
		if err != nil {
			log.Fatal(err)
		}
	}
	buildRunObj := createBuildRun(repoName)
	_, err = buildClient.ShipwrightV1alpha1().BuildRuns("default").Create(context.TODO(), buildRunObj, v1.CreateOptions{})
	if err != nil {
		log.Fatal(err)
	}

	err = wait.Poll(time.Second*4, time.Second*180, func() (bool, error) {
		existingBuildRun, err := buildClient.ShipwrightV1alpha1().BuildRuns("default").Get(context.TODO(), buildRunObj.Name, v1.GetOptions{})
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

	prepareCluster()

	fileServer := http.FileServer(http.Dir("./static"))
	http.Handle("/", fileServer)
	http.HandleFunc("/form", formHandler)

	fmt.Printf("Starting server at port %d\n", serverPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", serverPort), nil); err != nil {
		log.Fatal(err)
	}

}
