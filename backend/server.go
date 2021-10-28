package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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

func listBuildStrategies(w http.ResponseWriter, r *http.Request) {
	initializeClient()

	type supportedStrategy struct {
		v1.TypeMeta
		Name            string   `json:"name"`
		MandatoryFields []string `json:"mandatory"`
		OptionalFields  []string `json:"optional"`
	}

	potentiallyMandatoryFields := []string{
		"$(build.builder.image)",
		"$(build.dockerfile)",
	}

	strategies, err := shipwrightBuildClient.ShipwrightV1alpha1().ClusterBuildStrategies().List(context.TODO(), v1.ListOptions{})
	var supportedStrategies []supportedStrategy
	for _, s := range strategies.Items {

		inClusterSupportedStratgey := supportedStrategy{
			TypeMeta: s.TypeMeta,
			Name:     s.ObjectMeta.Name,
			MandatoryFields: []string{
				"$(build.source.url)",
			},
			OptionalFields: []string{
				"$(build.source.contextDir)",
				"$(build.source.revision)",
			},
		}
		for _, step := range s.Spec.BuildSteps {
			for _, arg := range step.Args {
				for _, potentiallyMandatoryField := range potentiallyMandatoryFields {
					if strings.Contains(arg, potentiallyMandatoryField) {
						inClusterSupportedStratgey.MandatoryFields = append(inClusterSupportedStratgey.MandatoryFields, potentiallyMandatoryField)
					}
				}
			}
		}

		supportedStrategies = append(supportedStrategies, inClusterSupportedStratgey)
	}

	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(supportedStrategies)
		return
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func buildStatusHandler(w http.ResponseWriter, r *http.Request) {

	paramValues := r.URL.Query()
	buildID := paramValues.Get("name")

	initializeClient()

	existingBuildRun, err := shipwrightBuildClient.ShipwrightV1alpha1().BuildRuns(buildSystemNamespace).Get(context.TODO(), buildID, v1.GetOptions{})
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(existingBuildRun.Status)
		return
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func formHandler(w http.ResponseWriter, r *http.Request) {

	sourceCodeURLParam := "build-source-url"
	contextDirParam := "build-source-contextDir"
	branchParam := "build-source-revision"
	dockerfileParam := "build-dockerfile"

	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	repoURL := r.FormValue(sourceCodeURLParam)
	contextDir := r.FormValue(contextDirParam)
	branch := r.FormValue(branchParam)
	dockerfile := r.FormValue(dockerfileParam)

	initializeClient()

	buildRequestID, _ := uuid.NewV4()
	buildObj := createBuild(buildRequestID.String(), repoURL, contextDir, branch, dockerfile)
	_, err := shipwrightBuildClient.ShipwrightV1alpha1().Builds(buildSystemNamespace).Create(context.TODO(), buildObj, v1.CreateOptions{})
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := shipwrightBuildClient.ShipwrightV1alpha1().BuildRuns(buildSystemNamespace).Get(context.TODO(), fmt.Sprintf("%s", buildRequestID.String()), v1.GetOptions{}); err == nil {
		err := shipwrightBuildClient.ShipwrightV1alpha1().BuildRuns(buildSystemNamespace).Delete(context.TODO(), fmt.Sprintf("%s", buildRequestID.String()), v1.DeleteOptions{})
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	buildRunObj := createBuildRun(buildRequestID.String())
	_, err = shipwrightBuildClient.ShipwrightV1alpha1().BuildRuns(buildSystemNamespace).Create(context.TODO(), buildRunObj, v1.CreateOptions{})
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// return the buildrun id so that it could be polled
	json.NewEncoder(w).Encode(buildRunObj.ObjectMeta)

}

func main() {

	fileServer := http.FileServer(http.Dir("./static"))
	http.Handle("/", fileServer)
	http.HandleFunc("/build", formHandler)
	http.HandleFunc("/buildstatus", buildStatusHandler)
	http.HandleFunc("/buildstrategies", listBuildStrategies)

	fmt.Printf("Starting server at port %d\n", serverPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", serverPort), nil); err != nil {
		log.Fatal(err)
	}

}
