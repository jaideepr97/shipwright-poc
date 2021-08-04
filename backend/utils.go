package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"os/exec"

	shipwright "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	goyaml "github.com/go-yaml/yaml"
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

func SplitYAML(resources []byte) ([][]byte, error) {
	dec := goyaml.NewDecoder(bytes.NewReader(resources))

	var res [][]byte
	for {
		var value interface{}
		err := dec.Decode(&value)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		valueBytes, err := goyaml.Marshal(value)
		if err != nil {
			return nil, err
		}
		res = append(res, valueBytes)
	}
	return res, nil
}
