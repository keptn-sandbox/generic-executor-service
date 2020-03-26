package main

import (
	"fmt"

	keptnevents "github.com/keptn/go-utils/pkg/events"
	keptnutils "github.com/keptn/go-utils/pkg/utils"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
)

// GenericScriptFolderBase Folder in the Keptn GitHub Repo where we expect scripts and http files
const GenericScriptFolderBase = "generic-executor/"

/**
* Here are all the handler functions for the individual events
  -> "sh.keptn.event.configuration.change"
  -> "sh.keptn.events.deployment-finished"
  -> "sh.keptn.events.tests-finished"
  -> "sh.keptn.event.start-evaluation"
  -> "sh.keptn.events.evaluation-done"
  -> "sh.keptn.event.problem.open"
  -> "sh.keptn.events.problem"
*/
//
// if any of the passed files exist either executes the bash or the http request
//
func executeScriptOrHTTP(keptnEvent baseKeptnEvent, logger *keptnutils.Logger, bashFilenames, httpFilenames []string) (bool, error) {

	success := true

	// lets start with the bashfiles
	for _, bashFilename := range bashFilenames {
		resource, err := getKeptnResource(keptnEvent, bashFilename, logger)
		if resource != "" && err == nil {
			logger.Debug("Found script " + bashFilename)
			var output string
			output, err = executeCommandWithKeptnContext(resource, nil, keptnEvent, nil, logger)
			if err != nil {
				logger.Error(fmt.Sprintf("Error executing script: %s", err.Error()))
				success = false
			} else {
				logger.Info("Script output: " + output)
			}
		} else {
			logger.Debug("No script found at " + bashFilename)
		}
	}

	// now we iterate through the http files
	for _, httpFilename := range httpFilenames {
		var parsedRequest genericHttpRequest
		resource, err := getKeptnResource(keptnEvent, httpFilename, logger)
		if resource != "" && err == nil {
			logger.Debug("Found http request " + httpFilename)
			parsedRequest, err = parseHttpRequestFromHttpTextFile(keptnEvent, httpFilename)
			if err == nil {
				statusCode, body, requestError := executeGenericHttpRequest(parsedRequest)
				if requestError != nil {
					logger.Error(fmt.Sprintf("Error: %s", requestError.Error()))
					success = false
				} else {
					logger.Info(fmt.Sprintf("%d - %s", statusCode, body))
				}
			} else {
				logger.Error(fmt.Sprintf("Error: %s", err.Error()))
				success = false
			}
		} else {
			logger.Debug("No http file found at " + httpFilename)
		}
	}

	return success, nil
}

//
// Handles an incoming event
//
func _executeScriptOrHttEventHandler(event cloudevents.Event, keptnEvent baseKeptnEvent, data interface{}, filePrefix string, logger *keptnutils.Logger) error {
	// we allow different files to be specified by the end user
	bashFilenames := []string{
		GenericScriptFolderBase + "all.events.sh",
		GenericScriptFolderBase + filePrefix + ".sh",
	}

	httpFilenames := []string{
		GenericScriptFolderBase + "all.events.http",
		GenericScriptFolderBase + filePrefix + ".http",
	}

	// now lets execute these scripts!
	_, err := executeScriptOrHTTP(keptnEvent, logger, bashFilenames, httpFilenames)
	if err != nil {
		logger.Error(fmt.Sprintf("Error: %s", err.Error()))
	}

	return err
}

//
// Handles ConfigurationChangeEventType = "sh.keptn.event.configuration.change"
// TODO: add in your handler code
//
func handleConfigurationChangeEvent(event cloudevents.Event, keptnEvent baseKeptnEvent, data *keptnevents.ConfigurationChangeEventData, logger *keptnutils.Logger) error {
	logger.Info(fmt.Sprintf("Handling Configuration Changed Event: %s", event.Context.GetID()))
	keptnEvent.event = "configuration.change"

	return _executeScriptOrHttEventHandler(event, keptnEvent, data, keptnEvent.event, logger)
}

//
// Handles DeploymentFinishedEventType = "sh.keptn.events.deployment-finished"
// TODO: add in your handler code
//
func handleDeploymentFinishedEvent(event cloudevents.Event, keptnEvent baseKeptnEvent, data *keptnevents.DeploymentFinishedEventData, logger *keptnutils.Logger) error {
	logger.Info(fmt.Sprintf("Handling Deployment Finished Event: %s", event.Context.GetID()))
	keptnEvent.event = "deployment.finished"

	return _executeScriptOrHttEventHandler(event, keptnEvent, data, keptnEvent.event, logger)
}

//
// Handles TestsFinishedEventType = "sh.keptn.events.tests-finished"
// TODO: add in your handler code
//
func handleTestsFinishedEvent(event cloudevents.Event, keptnEvent baseKeptnEvent, data *keptnevents.TestsFinishedEventData, logger *keptnutils.Logger) error {
	logger.Info(fmt.Sprintf("Handling Tests Finished Event: %s", event.Context.GetID()))
	keptnEvent.event = "tests.finished"

	return _executeScriptOrHttEventHandler(event, keptnEvent, data, keptnEvent.event, logger)
}

//
// Handles EvaluationDoneEventType = "sh.keptn.events.start-evaluation"
// TODO: add in your handler code
//
func handleStartEvaluationEvent(event cloudevents.Event, keptnEvent baseKeptnEvent, data *keptnevents.StartEvaluationEventData, logger *keptnutils.Logger) error {
	logger.Info(fmt.Sprintf("Handling Start Evaluation Event: %s", event.Context.GetID()))
	keptnEvent.event = "start.evaluation"

	return _executeScriptOrHttEventHandler(event, keptnEvent, data, keptnEvent.event, logger)
}

//
// Handles DeploymentFinishedEventType = "sh.keptn.events.evaluation-done"
// TODO: add in your handler code
//
func handleEvaluationDoneEvent(event cloudevents.Event, keptnEvent baseKeptnEvent, data *keptnevents.EvaluationDoneEventData, logger *keptnutils.Logger) error {
	logger.Info(fmt.Sprintf("Handling Evaluation Done Event: %s", event.Context.GetID()))
	keptnEvent.event = "evaluation.done"

	return _executeScriptOrHttEventHandler(event, keptnEvent, data, keptnEvent.event, logger)
}

//
// Handles ProblemOpenEventType = "sh.keptn.event.problem.open"
// Handles ProblemEventType = "sh.keptn.events.problem"
// TODO: add in your handler code
//
func handleProblemEvent(event cloudevents.Event, keptnEvent baseKeptnEvent, data *keptnevents.ProblemEventData, logger *keptnutils.Logger) error {
	logger.Info(fmt.Sprintf("Handling Problem Open Event: %s", event.Context.GetID()))
	keptnEvent.event = "problem.open"

	return _executeScriptOrHttEventHandler(event, keptnEvent, data, keptnEvent.event, logger)
}
