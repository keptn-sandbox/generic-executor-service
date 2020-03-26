# Repository of useful generic executor scripts
This folder is meant to share useful scripts that people can reuse.
If you have a useful script just do a PR and I am happy to add your script to this folder and give it a quick description in this readme

| Script        | Explanation    |
| ------------- |-------------   | 
| send.dynatrace.metric.http | Creates a custom metric in Dynatrace to measure the number of events processed by Keptn | 


## send.dynatrace.metric.http

This one will send custom metrics to dynatrace for every event that Keptn processes. It does it by sending custom metrics for each event type, e.g: configuration.changed, tests.finished ... - the metrics are also sent with the key keptn dimensions project, stage & service.
In order for this script to work you first have to create these custom metrics in Dynatrace. You can do this by calling the script createDynatraceKeptnCustomMetrics.sh like this:
```
export DT_API_TOKEN=YOURAPITOKEN
export DT_TENANT=YOURTENANTURL
./createDynatraceKeptnCustomMetrics.sh
```

Now you can use send.dynatrace.metric.http which will automatically send custom metric values to Dynatrace for each event.
In Dynatrace you will see a new custom Device called Keptn. If you want to give your Keptn a more specific name you can change that in the send.dynatrace.metric.http by changing DisplayName to e.g: "My Keptn Installation on ABC"

To get all events I suggest to add it as an all.events.http - like this:
```
keptn add-resource --project=PROJECTNAME --stage=STAGE --service=SERVICENAME --resource=./send.dynatrace.metric.sh --resourceUri=generic-executor/all.events.http
``` 