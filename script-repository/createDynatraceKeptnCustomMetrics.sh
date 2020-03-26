#!/bin/bash

# Usage:
# Assumes DT_TENANT and DT_API_TOKEN environment variables are set
# ./createDynatraceKeptnCustomMetrics.sh

if [[ -z "$DT_TENANT" || -z "$DT_API_TOKEN" ]]; then
  echo "DT_TENANT & DT_API_TOKEN MUST BE SET!!"
  exit 1
fi

####################################################################################################################
## createCustomMetric(METRICKEY, METRICNAME)
####################################################################################################################
# Example: createCustomMetric("custom:keptn.events.configuration.change", "Keptn Configuration Change Events")
function createCustomMetric() {
    METRICKEY=$1
    METRICNAME=$2
  
    PAYLOAD='{
        "displayName" : "'$METRICNAME'",
        "unit" : "Count",
	        "dimensions": [
            "project",
            "stage",
            "service"
        ],
        "types": [
            "Event"
    	    ]
        }'    

    echo "Creating Custom Metric $METRICNAME($METRICKEY)"
    echo "PUT https://$DT_TENANT/api/v1/timeseries/$METRICKEY"
    echo "$PAYLOAD"

    curl -X PUT \
        "https://$DT_TENANT/api/v1/timeseries/$METRICKEY" \
        -H 'accept: application/json; charset=utf-8' \
        -H "Authorization: Api-Token $DT_API_TOKEN" \
        -H 'Content-Type: application/json; charset=utf-8' \
        -d "$PAYLOAD" \
        -o curloutput.txt

    cat curloutput.txt
}

# now lets create metrics for each event
createCustomMetric "custom:keptn.events.configuration.change" "Keptn Configuration Change Events"
createCustomMetric "custom:keptn.events.deployment.finished" "Keptn Deployment Finished Events"
createCustomMetric "custom:keptn.events.tests.finished" "Keptn Tests Finished Events"
createCustomMetric "custom:keptn.events.start.evaluation" "Keptn Start Evaluation Events"
createCustomMetric "custom:keptn.events.evaluation.done" "Keptn Evaluation Done Events"
createCustomMetric "custom:keptn.events.problem.open" "Keptn Problem Open Events"
