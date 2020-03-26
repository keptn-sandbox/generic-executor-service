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
The *generic-executor-service* by default handles all Keptn events and then searches for either .sh or .http files in the stage & service specific Keptn Git repo in the subfolder *generic-executor*. If it doesnt find either all.events.* or event.* file it looks in the stage folder. If nothing is there it looks in the project repo (master branch). 

Here is a sample folder structure in my Git repo for a specific service and stage:
```
[STAGE]/MYSERVICE/genericexecutor
-- all.events.sh              <-- executed for all events for this service
-- configuration.change.sh    <-- executed for configuration.change
-- configuration.change.http  <-- executed for configuration.change

[STAGE]/genericexecutor
-- all.events.sh              <-- executed for all events for all services unless file exists on service level
-- configuration.change.sh    <-- executed for configuration.change for all services unless file exists on service level
-- configuration.change.http  <-- executed for configuration.change for all services unless file exists on service level

[MASTER]/genericexecutor
-- all.events.sh              <-- executed for all events unless file exists on stage service level
-- configuration.change.sh    <-- executed for configuration.change for all services unless file exists on stage or service level
-- configuration.change.http  <-- executed for configuration.change for all services unless file exists on stage or service level
```

The *generic-executor-service* will first execute those files with the name all.events.sh and all.events.http. This gives you the ability to specify one set of action that should be executed for every Keptn event. Good news is that you can also specify these files on a stage or project level. If the *generic-executor-service* doesnt find a file on service level it looks at stage level and then on project. The first that is found will be executed!
After that the *generic-executor-service* will look for a file called KEPTN-EVENT.sh or KEPTN-EVENT.http where KEPTN-EVENT can be one of the following values corresponding to the Keptn events
```
-- configuration.change.*
-- deployment.finished.*
-- tests.finished.*
-- start.evaluation.*
-- evaluation.done.*
-- problem.open.*
```

This gives you full flexiblity to provide a bash and http script for each event or specify a bash and http script that shoudl be executed for all events.

Please have a look at the sample .http and .sh files to see how the *generic-executor-service* is not only calling these scripts or making http calls. The service is also passing Keptn Event specific context data such as PROJECT, SERVICE, LABELS and also ENV-Variables of the *generic-executor-service* pod as variables that you can reference. This gives you a lot of flexibility when writing these scripts.

Here a sample http script that shows you how to call an external webhook with this capability.
The *generic-executor-service* will replace the core Keptn Event values as well as provides each label via $LABEL_LABELNAME and each Environment Variable via $ENV_ENVNAME
```
configuration.change.http:
POST https://webhook.site/YOURHOOKID
Accept: application/json
Cache-Control: no-cache
Content-Type: application/cloudevents+json

{
  "contenttype": "application/json",
  "deploymentstrategy": "blue_green_service",
  "project": "$PROJECT",
  "service": "$SERVICE",
  "stage": "$STAGE",
  "mylabel" : "$LABEL_gitcommit",
  "mytoken" : "$ENV_TESTTOKEN",
  "shkeptncontext": "$CONTEXT",
  "event": "$EVENT",
  "source": "$SOURCE"
}
```

And here a sample bash script that the *generic-executor-service* is calling by setting all the Keptn context, labels and container environment variables as environment variables for this script:
```
all.event.sh:
#!/bin/bash

# This is a script that will be executed by the Keptn Generic Executor Service for ANY event as the filename is called all.events.sh!
# It will be called with a couple of enviornment variables that are filled with Keptn Event Details, Env-Variables from the Service container as well as labels

echo "This is my all.events.sh script"
echo "Context = $CONTEXT"
echo "Project = $PROJECT"
echo "Project = $PROJECT"
echo "Service = $SERVICE"
echo "Stage = $STAGE"
echo "GitCommit = $LABEL_gitcommit"
echo "TestToken = $ENV_TESTTOKEN"

# Here i could do whatever I want with these values, e.g: call an external tool :-)

```

Last but not least - here are all the available placeholders for .http files and env-variables that are passed to your .sh files:
```
// Event Context
$CONTEXT,$EVENT,$SOURCE,$TIMESTRING,$TIMEUTCSTRING,$TIMEUTCMS

// Project Context
$PROJECT,$STAGE,$SERVICE,$DEPLOYMENT,$TESTSTRATEGY
    
// Deployment Finished specific
$DEPLOYMENTURILOCAL,$DEPLOYMENTURIPUBLIC

// Labels will be made available with a $LABEL_ prefix, e.g.:
$LABEL_gitcommit,$LABEL_anotherlabel,$LABEL_xxxx

// Environment variables you pass to the generic-executor-service container in the service.yaml will be available with $ENV_ prefix
$ENV_YOURCUSTOMENV,$ENV_KEPTN_API_TOKEN,$ENV_KEPTN_ENDPOINT,...
```


Enjoy the fun!

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