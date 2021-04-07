#!/bin/bash

# This script could execute a test and then return a json with start and endtime which will be passed to Keptn's "test" data structure in the .finished event
jsonResponse="{
    \"start\" : \"$TIMEUTCMS\",
    \"end\" : \"$TIMEUTCMS\"
    }"

# If you echo anything else than the JSON it will be considered regular response 
echo "hallo - we just executed a test against $DATA_DEPLOYMENT_DEPLOYMENTURISPUBLIC_0"
echo "Results are written to $ID.finished.event.json"
echo $jsonResponse > "$ID.finished.event.json"