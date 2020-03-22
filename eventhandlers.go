package main

import (
	"fmt"

	keptnutils "github.com/keptn/go-utils/pkg/utils"
	keptnevents "github.com/keptn/go-utils/pkg/events"

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
// Handles ConfigurationChangeEventType = "sh.keptn.event.configuration.change"
// TODO: add in your handler code
//
func handleConfigurationChangeEvent(event cloudevents.Event, shkeptncontext string, data *keptnevents.ConfigurationChangeEventData, logger *keptnutils.Logger) error {
	logger.Info(fmt.Sprintf("Handling Configuration Changed Event: %s", event.Context.GetID()));

	resource, err := getKeptnResource(data.Project, data.Service, data.Stage, "./scripts/configuration.change.sh", logger)
	if (resource != "" && err == nil) {
		logger.Info("Found script ./scripts/configuration.change.sh");
		executeCommand(resource, nil, logger);
	} else {
		logger.Info("No script found at ./scripts/configuration.change.sh")
	}

	return nil
}

//
// Handles DeploymentFinishedEventType = "sh.keptn.events.deployment-finished"
// TODO: add in your handler code
//
func handleDeploymentFinishedEvent(event cloudevents.Event, shkeptncontext string, data *keptnevents.DeploymentFinishedEventData, logger *keptnutils.Logger) error {
	logger.Info(fmt.Sprintf("Handling Deployment Finished Event: %s", event.Context.GetID()));

	return nil
}

//
// Handles TestsFinishedEventType = "sh.keptn.events.tests-finished"
// TODO: add in your handler code
//
func handleTestsFinishedEvent(event cloudevents.Event, shkeptncontext string, data *keptnevents.TestsFinishedEventData, logger *keptnutils.Logger) error {
	logger.Info(fmt.Sprintf("Handling Tests Finished Event: %s", event.Context.GetID()));

	return nil
}

//
// Handles EvaluationDoneEventType = "sh.keptn.events.evaluation-done"
// TODO: add in your handler code
//
func handleStartEvaluationEvent(event cloudevents.Event, shkeptncontext string, data *keptnevents.StartEvaluationEventData, logger *keptnutils.Logger) error {
	logger.Info(fmt.Sprintf("Handling Start Evaluation Event: %s", event.Context.GetID()));

	return nil
}

//
// Handles DeploymentFinishedEventType = "sh.keptn.events.deployment-finished"
// TODO: add in your handler code
//
func handleEvaluationDoneEvent(event cloudevents.Event, shkeptncontext string, data *keptnevents.EvaluationDoneEventData, logger *keptnutils.Logger) error {
	logger.Info(fmt.Sprintf("Handling Evaluation Done Event: %s", event.Context.GetID()));

	return nil
}

//
// Handles ProblemOpenEventType = "sh.keptn.event.problem.open"
// Handles ProblemEventType = "sh.keptn.events.problem"
// TODO: add in your handler code
//
func handleProblemEvent(event cloudevents.Event, shkeptncontext string, data *keptnevents.ProblemEventData, logger *keptnutils.Logger) error {
	logger.Info(fmt.Sprintf("Handling Problem Event: %s", event.Context.GetID()));

	return nil
}