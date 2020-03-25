# Generic Executor Service for Keptn

This is a Sandbox Keptn Service that enables generic execution of bash files and HTTP requests for individual Keptn Events 

![GitHub release (latest by date)](https://img.shields.io/github/v/release/grabnerandi/generic-executor-service)
[![Build Status](https://travis-ci.org/grabnerandi/generic-executor-service.svg?branch=master)](https://travis-ci.org/grabnerandi/generic-executor-service)
[![Go Report Card](https://goreportcard.com/badge/github.com/grabnerandi/generic-executor-service)](https://goreportcard.com/report/github.com/grabnerandi/generic-executor-service)

This implements a generic-executor-service for Keptn.

## Compatibility Matrix

| Keptn Version    | [Generic Executor Service for Keptn](https://hub.docker.com/r/grabnerandi/generic-executor-service/tags) |
|:----------------:|:----------------------------------------:|
|       0.6.1      | grabnerandi/generic-executor-service:0.1.0 |

## Installation

The *generic-executor-service* can be installed as a part of [Keptn's uniform](https://keptn.sh).

### Deploy in your Kubernetes cluster

To deploy the current version of the *generic-executor-service* in your Keptn Kubernetes cluster, apply the [`deploy/service.yaml`](deploy/service.yaml) file:

```console
kubectl apply -f deploy/service.yaml
```

This should install the `generic-executor-service` together with a Keptn `distributor` into the `keptn` namespace, which you can verify using

```console
kubectl -n keptn get deployment generic-executor-service -o wide
kubectl -n keptn get pods -l run=generic-executor-service
```

### Up- or Downgrading

Adapt and use the following command in case you want to up- or downgrade your installed version (specified by the `$VERSION` placeholder):

```console
kubectl -n keptn set image deployment/generic-executor-service generic-executor-service=your-username/generic-executor-service:$VERSION --record
```

### Uninstall

To delete a deployed *generic-executor-service*, use the file `deploy/*.yaml` files from this repository and delete the Kubernetes resources:

```console
kubectl delete -f deploy/service.yaml
```

## Development

Be my guest and help me extend this Generic Executor Service for Keptn with new capabilities. 

### Where to start

If you don't care about the details, your first entrypoint is [eventhandlers.go](eventhandlers.go). This is where it handles incoming Keptn events
 
To better understand Keptn CloudEvents, please look at the [Keptn Spec](https://github.com/keptn/spec).
 
If you want to get more insights, please look into [main.go](main.go), [deploy/service.yaml](deploy/service.yaml),
 consult the [Keptn docs](https://keptn.sh/docs/) as well as existing [Keptn Core](https://github.com/keptn/keptn) and
 [Keptn Contrib](https://github.com/keptn-contrib/) services.

### Build yourself

If you want to build this service yourself here is what you need to do

* Build the binary: `go build -ldflags '-linkmode=external' -v -o generic-executor-service`
* Run tests: `go test -race -v ./...`
* Build the docker image: `docker build . -t your-username/generic-executor-service:dev` (Note: Replace `your-username` with your DockerHub account/organization)
* Push the docker image to DockerHub: `docker push your-username/generic-executor-service:dev` (Note: Replace `your-username` with your DockerHub account/organization)
* Deploy the service using `kubectl`: `kubectl apply -f deploy/` (Note: Update the image reference in the service.yaml to point to your docker image on DockerHub)
* Undeploy the service using `kubectl`: `kubectl deploy -f deploy/`
* Watch the deployment using `kubectl`: `kubectl -n keptn get deployment generic-executor-service -o wide`
* Get logs using `kubectl`: `kubectl -n keptn logs deployment/generic-executor-service -f`
* Watch the deployed pods using `kubectl`: `kubectl -n keptn get pods -l run=generic-executor-service`
* Deploy the service using [Skaffold](https://skaffold.dev/): `skaffold run --tail` (Note: please adapt the image name in [skaffold.yaml](skaffold.yaml))

### Testing Cloud Events

We have dummy cloud-events in the form of PostMan Requests in the [test-events/](test-events/) directory.

## License

Please find more information in the [LICENSE](LICENSE) file.