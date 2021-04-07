#!/bin/bash

# This is a script that will be executed by the Keptn Generic Executor Service
# It will be called with a couple of enviornment variables that are filled with Keptn Event Details, Env-Variables from the Service container as well as labels

echo "Deployment Triggered Sample Script"
echo "Project = $DATA_PROJECT"
echo "Service = $DATA_SERVICE"
echo "Stage = $DATA_STAGE"
echo "Image = $DATA_CONFIGURATIONCHANGE_VALUES_IMAGE"
echo "DeploymentStrategy = $DATA_DEPLOYMENT_DEPLOYMENTSTRATEGY"
echo "TestToken = $ENV_TESTTOKEN"

echo "And here the content of the full JSON object"
cat $1