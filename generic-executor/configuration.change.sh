#!/bin/bash

# This is a script that will be executed by the Keptn Generic Executor Service
# It will be called with a couple of enviornment variables that are filled with Keptn Event Details, Env-Variables from the Service container as well as labels

echo "Configuration Change Sample Script"
echo "Project = $PROJECT"
echo "Service = $SERVICE"
echo "Stage = $STAGE"
echo "GitCommit = $LABEL_gitcommit"
echo "TestToken = $ENV_TESTTOKEN"