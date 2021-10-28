## Shipwright Build Service

## List available build strategies

### Request

```
curl -X GET \
  http://localhost:8085/buildstrategies \
  -H 'cache-control: no-cache'
```

### Response

```
[
    {
        "kind": "ClusterBuildStrategy",
        "apiVersion": "shipwright.io/v1alpha1",
        "name": "buildah",
        "mandatory": [
            "$(build.source.url)",
            "$(build.dockerfile)"
        ],
        "optional": [
            "$(build.source.contextDir)",
            "$(build.source.revision)"
        ]
    },
    {
        "kind": "ClusterBuildStrategy",
        "apiVersion": "shipwright.io/v1alpha1",
        "name": "buildkit",
        "mandatory": [
            "$(build.source.url)",
            "$(build.dockerfile)"
        ],
        "optional": [
            "$(build.source.contextDir)",
            "$(build.source.revision)"
        ]
    },
    {
        "kind": "ClusterBuildStrategy",
        "apiVersion": "shipwright.io/v1alpha1",
        "name": "buildpacks-v3",
        "mandatory": [
            "$(build.source.url)"
        ],
        "optional": [
            "$(build.source.contextDir)",
            "$(build.source.revision)"
        ]
    },
    {
        "kind": "ClusterBuildStrategy",
        "apiVersion": "shipwright.io/v1alpha1",
        "name": "kaniko",
        "mandatory": [
            "$(build.source.url)",
            "$(build.dockerfile)"
        ],
        "optional": [
            "$(build.source.contextDir)",
            "$(build.source.revision)"
        ]
    },
    {
        "kind": "ClusterBuildStrategy",
        "apiVersion": "shipwright.io/v1alpha1",
        "name": "ko",
        "mandatory": [
            "$(build.source.url)"
        ],
        "optional": [
            "$(build.source.contextDir)",
            "$(build.source.revision)"
        ]
    },
    {
        "kind": "ClusterBuildStrategy",
        "apiVersion": "shipwright.io/v1alpha1",
        "name": "source-to-image",
        "mandatory": [
            "$(build.source.url)"
        ],
        "optional": [
            "$(build.source.contextDir)",
            "$(build.source.revision)"
        ]
    },
]
```


## Starting a new Build

### Request

```
curl -X POST \
  http://localhost:8085/form \
  -H 'cache-control: no-cache' \
  -H 'content-type: application/x-www-form-urlencoded' \
  -X POST -d "build-source-url=https://github.com/sbose78/sample-nodejs&build-source-contextDir=source-build"
```

### Response

```
{
    "name": "cf44fa8c-66f0-4663-7b2b-3863a4f47a73",
    "namespace": "shipwright-tenant",
    "creationTimestamp": null
}
```


##  Fetching build status by name

### Request

```
curl -X GET \
  'http://localhost:8085/buildstatus?name=cf44fa8c-66f0-4663-7b2b-3863a4f47a73' \
  -H 'cache-control: no-cache' \
  -H 'content-type: application/x-www-form-urlencoded'
```


### Response

```
{
    "conditions": [
        {
            "type": "Succeeded",
            "status": "False", # <----- Should be True when the build succeeds.
            "lastTransitionTime": "2021-10-28T17:02:16Z",
            "reason": "CouldntGetTask",
            "message": "....."
        }
    ],
    "latestTaskRunRef": "cf44fa8c-66f0-4663-7b2b-3863a4f47a73-srlfq",
    "startTime": "2021-10-28T17:02:16Z",
    "completionTime": "2021-10-28T17:02:16Z",
    "buildSpec": {
        "source": {
            "url": "https://github.com/sbose78/sample-nodejs",
            "revision": "",
            "contextDir": "source-build"
        },
        "strategy": {
            "name": "buildpacks-v3",
            "kind": "ClusterBuildStrategy"
        },
        "dockerfile": "",
        "output": {
            "image": "docker.io/sbose78/generated:cf44fa8c-66f0-4663-7b2b-3863a4f47a73", #<--- System-generated image reference
            "credentials": {
                "name": "my-docker-credentials"
            }
        }
    }
}
```

