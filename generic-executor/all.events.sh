#!/bin/bash

# This is a script that will be executed by the Keptn Generic Executor Service for ANY event as the filename is called all.events.sh!
# It will be called with a couple of enviornment variables that are filled with Keptn Event Details, Env-Variables from the Service container as well as labels

echo "This is my all.events.sh script"
echo "Context = ${SHKEPTNCONTEXT}"

# Here i could do whatever I want