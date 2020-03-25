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

## Usage

The purpose of the *generic-executor-service" is to allow users to provide either .sh (shell scripts) or .http (HTTP Request) files that will be executed when Keptn sends different events, e.g: you want to execute a specific script when a deployment-finished event is sent.
The *generic-executor-service* by default handles all Keptn events and then searches for either .sh or .http files in the stage & service specific Keptn Git repo in the subfolder *generic-executor*. Here is a sample folder structure in my Git repo for a specific service and stage:
```
/MYSERVICE/genericexecutor
-- all.events.sh
-- configuration.change.sh
-- configuration.change.http
```

The *generic-executor-service* will first execute those files with the name all.events.sh and all.events.http. This gives you the ability to specify one set of action that should be executed for every Keptn event.
After that the *generic-executor-service* will look for a file called KEPTN-EVENT.sh or KEPTN-EVENT.http where KEPTN-EVENT can be one of the following values corresponding to the Keptn events
-- configuration.change.*
-- deployment.finished.*
-- tests.finished.*
-- start.evaluation.*
-- evaluation.done.*
-- problem.open.*

This gives you full flexiblity to provide a bash and http script for each event or specify a bash and http script that shoudl be executed for all events.

Please have a look at the sample .http and .sh files to see how the *generic-executor-service* is not only calling these scripts or making http calls. The service is also passing Keptn Event specific context data such as PROJECT, SERVICE, LABELS and also ENV-Variables of the *generic-executor-service* pod as variables that you can reference. This gives you a lot of flexibility when writing these scripts.

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
* Undeploy the service using `kubectl`: `kubectl apply -f deploy/`
* Watch the deployment using `kubectl`: `kubectl -n keptn get deployment generic-executor-service -o wide`
* Get logs using `kubectl`: `kubectl -n keptn logs deployment/generic-executor-service -f`
* Watch the deployed pods using `kubectl`: `kubectl -n keptn get pods -l run=generic-executor-service`
* Deploy the service using [Skaffold](https://skaffold.dev/): `skaffold run --tail` (Note: please adapt the image name in [skaffold.yaml](skaffold.yaml))

### Testing Cloud Events

We have dummy cloud-events in the form of PostMan Requests in the [test-events/](test-events/) directory.

## License

Please find more information in the [LICENSE](LICENSE) file.