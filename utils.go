package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"

	shipwright "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var (
	strategyKind shipwright.BuildStrategyKind = shipwright.ClusterBuildStrategyKind
)

func TypeMeta(kind, apiVersion string) v1.TypeMeta {
	return v1.TypeMeta{
		Kind:       kind,
		APIVersion: apiVersion,
	}
}

func ObjectMeta(n types.NamespacedName, opts ...objectMetaFunc) v1.ObjectMeta {
	om := v1.ObjectMeta{
		Namespace: n.Namespace,
		Name:      n.Name,
	}
	for _, o := range opts {
		o(&om)
	}
	return om

}

func prepareCluster() {
	kubectlPath, err := exec.LookPath("kubectl")
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command(kubectlPath, "apply", "--filename", "https://storage.googleapis.com/tekton-releases/pipeline/previous/v0.25.0/release.yaml")
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	cmd = exec.Command(kubectlPath, "apply", "--filename", "https://github.com/shipwright-io/build/releases/download/v0.5.1/release.yaml")
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	cmd = exec.Command(kubectlPath, "apply", "--filename", "https://github.com/shipwright-io/build/releases/download/nightly/default_strategies.yaml")
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

// handleDockerCfgJSONContent serializes a ~/.docker/config.json file
func handleDockerCfgJSONContent(username, password, email, server string) ([]byte, error) {
	dockerConfigAuth := DockerConfigEntry{
		Username: username,
		Password: password,
		Email:    email,
		Auth:     encodeDockerConfigFieldAuth(username, password),
	}
	dockerConfigJSON := DockerConfigJSON{
		Auths: map[string]DockerConfigEntry{server: dockerConfigAuth},
	}

	return json.Marshal(dockerConfigJSON)
}

// encodeDockerConfigFieldAuth returns base64 encoding of the username and password string
func encodeDockerConfigFieldAuth(username, password string) string {
	fieldValue := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(fieldValue))
}

func createDockerSecret(username, password, email, server string) *corev1.Secret {

	dockerConfigJSONContent, err := handleDockerCfgJSONContent(username, password, email, server)
	if err != nil {
		log.Fatal(err)
	}

	dockerSecret := &corev1.Secret{
		TypeMeta:   TypeMeta("Secret", "v1"),
		ObjectMeta: ObjectMeta(types.NamespacedName{Namespace: "default", Name: secretName}),
		Type:       corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			corev1.DockerConfigJsonKey: dockerConfigJSONContent,
		},
	}

	return dockerSecret
}

func createBuild(imageRegistry, repoURL, username, repoName, secretName, contextDir string) *shipwright.Build {

	build := &shipwright.Build{
		TypeMeta:   TypeMeta("Build", "shipwright.io/v1alpha1"),
		ObjectMeta: ObjectMeta(types.NamespacedName{Namespace: "", Name: fmt.Sprintf("%v-build", repoName)}),
		Spec: shipwright.BuildSpec{
			Source: shipwright.Source{
				URL:        repoURL,
				ContextDir: &contextDir,
			},
			Strategy: &shipwright.Strategy{
				Name: "buildpacks-v3",
				Kind: &strategyKind,
			},
			Output: shipwright.Image{
				Image: fmt.Sprintf("%s.io/%s/%v:latest", imageRegistry, username, repoName),
				Credentials: &corev1.LocalObjectReference{
					Name: secretName,
				},
			},
		},
	}

	return build
}

func createBuildRun(repoName string) *shipwright.BuildRun {

	buildRun := &shipwright.BuildRun{
		TypeMeta: TypeMeta("BuildRun", "shipwright.io/v1alpha1"),
		// ObjectMeta: v1.ObjectMeta{GenerateName: fmt.Sprintf("%v-buildrun-", repoName)},
		ObjectMeta: ObjectMeta(types.NamespacedName{Namespace: "default", Name: fmt.Sprintf("%v-buildrun", repoName)}),
		Spec: shipwright.BuildRunSpec{
			BuildRef: &shipwright.BuildRef{
				Name: fmt.Sprintf("%v-build", repoName),
			},
		},
	}

	return buildRun
}
