package main

import (
	"fmt"

	shipwright "github.com/shipwright-io/build/pkg/apis/build/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func createDockerSecret(username, password, email, server string) (*corev1.Secret, error) {

	dockerConfigJSONContent, err := handleDockerCfgJSONContent(username, password, email, server)
	if err != nil {
		return nil, fmt.Errorf("error creating docker config json content : %v", err)
	}

	dockerSecret := &corev1.Secret{
		TypeMeta:   TypeMeta("Secret", "v1"),
		ObjectMeta: ObjectMeta(types.NamespacedName{Namespace: buildSystemNamespace, Name: secretName}),
		Type:       corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			corev1.DockerConfigJsonKey: dockerConfigJSONContent,
		},
	}

	return dockerSecret, nil
}

func createBuild(repoName string, repoURL, contextDir string) *shipwright.Build {

	build := &shipwright.Build{
		TypeMeta:   TypeMeta("Build", "shipwright.io/v1alpha1"),
		ObjectMeta: ObjectMeta(types.NamespacedName{Namespace: "", Name: fmt.Sprintf("%v", repoName)}),
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
				Image: fmt.Sprintf("%s/%s/%s:%s", imageRegistryServer, quayUsername, imageRepo, imageRepo),
				Credentials: &corev1.LocalObjectReference{
					Name: secretName,
				},
			},
		},
	}

	return build
}

func createBuildRun(buildName string) *shipwright.BuildRun {

	buildRun := &shipwright.BuildRun{
		TypeMeta: TypeMeta("BuildRun", "shipwright.io/v1alpha1"),
		// ObjectMeta: v1.ObjectMeta{GenerateName: fmt.Sprintf("%v-buildrun-", repoName)},
		ObjectMeta: ObjectMeta(types.NamespacedName{Namespace: buildSystemNamespace, Name: fmt.Sprintf("%s", buildName)}),
		Spec: shipwright.BuildRunSpec{
			BuildRef: &shipwright.BuildRef{
				Name: fmt.Sprintf("%s", buildName),
			},
		},
	}

	return buildRun
}
