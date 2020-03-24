package main

import (
	"fmt"

	keptnevents "github.com/keptn/go-utils/pkg/events"
	keptnutils "github.com/keptn/go-utils/pkg/utils"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
)

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
// if the passed files exist either executes the bash or the http request
//
func executeScriptOrHTTP(keptnEvent baseKeptnEvent, logger *keptnutils.Logger, bashFilename, httpFilename string) (bool, error) {
	resource, err := getKeptnResource(keptnEvent, bashFilename, logger)
	if resource != "" && err == nil {
		logger.Info("Found script " + bashFilename)
		_, err = executeCommand(resource, nil, logger)
		if err != nil {
			return false, err
		}
	} else {
		logger.Info("No script found at " + bashFilename)
	}

	var parsedRequest genericHttpRequest
	resource, err = getKeptnResource(keptnEvent, httpFilename, logger)
	if resource != "" && err == nil {
		logger.Info("Found http request " + httpFilename)
		parsedRequest, err = parseHttpRequestFromHttpTextFile(keptnEvent, httpFilename)
		if err == nil {
			statusCode, body, requestError := executeGenericHttpRequest(parsedRequest)
			if requestError != nil {
				logger.Error(fmt.Sprintf("Error: %s", err.Error()))
				return false, err
			} else {
				logger.Info(fmt.Sprintf("%d - %s", statusCode, body))
			}
		} else {
			return false, err
		}
	} else {
		logger.Info("No http found at " + httpFilename)
	}

	return true, nil
}

//
// Handles ConfigurationChangeEventType = "sh.keptn.event.configuration.change"
// TODO: add in your handler code
//
func handleConfigurationChangeEvent(event cloudevents.Event, keptnEvent baseKeptnEvent, data *keptnevents.ConfigurationChangeEventData, logger *keptnutils.Logger) error {
	logger.Info(fmt.Sprintf("Handling Configuration Changed Event: %s", event.Context.GetID()))

	_, err := executeScriptOrHTTP(keptnEvent, logger, "./scripts/configuration.change.sh", "./scripts/configuration.change.http")
	if err != nil {
		logger.Error(fmt.Sprintf("Error: %s", err.Error()))
	}

	return err
}

//
// Handles DeploymentFinishedEventType = "sh.keptn.events.deployment-finished"
// TODO: add in your handler code
//
func handleDeploymentFinishedEvent(event cloudevents.Event, keptnEvent baseKeptnEvent, data *keptnevents.DeploymentFinishedEventData, logger *keptnutils.Logger) error {
	logger.Info(fmt.Sprintf("Handling Deployment Finished Event: %s", event.Context.GetID()))

	return nil
}

//
// Handles TestsFinishedEventType = "sh.keptn.events.tests-finished"
// TODO: add in your handler code
//
func handleTestsFinishedEvent(event cloudevents.Event, keptnEvent baseKeptnEvent, data *keptnevents.TestsFinishedEventData, logger *keptnutils.Logger) error {
	logger.Info(fmt.Sprintf("Handling Tests Finished Event: %s", event.Context.GetID()))

	return nil
}

//
// Handles EvaluationDoneEventType = "sh.keptn.events.evaluation-done"
// TODO: add in your handler code
//
func handleStartEvaluationEvent(event cloudevents.Event, keptnEvent baseKeptnEvent, data *keptnevents.StartEvaluationEventData, logger *keptnutils.Logger) error {
	logger.Info(fmt.Sprintf("Handling Start Evaluation Event: %s", event.Context.GetID()))

	return nil
}

//
// Handles DeploymentFinishedEventType = "sh.keptn.events.deployment-finished"
// TODO: add in your handler code
//
func handleEvaluationDoneEvent(event cloudevents.Event, keptnEvent baseKeptnEvent, data *keptnevents.EvaluationDoneEventData, logger *keptnutils.Logger) error {
	logger.Info(fmt.Sprintf("Handling Evaluation Done Event: %s", event.Context.GetID()))

	return nil
}

//
// Handles ProblemOpenEventType = "sh.keptn.event.problem.open"
// Handles ProblemEventType = "sh.keptn.events.problem"
// TODO: add in your handler code
//
func handleProblemEvent(event cloudevents.Event, keptnEvent baseKeptnEvent, data *keptnevents.ProblemEventData, logger *keptnutils.Logger) error {
	logger.Info(fmt.Sprintf("Handling Problem Event: %s", event.Context.GetID()))

	return nil
}
