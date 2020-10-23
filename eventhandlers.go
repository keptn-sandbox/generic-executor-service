package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	keptn "github.com/keptn/go-utils/pkg/lib"
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

const EXECUTESTATUS_NOACTIONFOUND = 0
const EXECUTESTATUS_SUCCESSFULL = 1
const EXECUTESTATUS_FAILED = 2
const EXECUTESTATUS_ACTIONFOUND = 3

//
// if any of the passed files exist either executes the bash or the http request
// The return status depends on the success of the executed script or HTTP Request. If the script fails or if the HTTP call returns a status code >= 300 the call is considered failed
//
func executeScriptOrHTTP(myKeptn *keptn.Keptn, keptnEvent BaseKeptnEvent, data interface{}, bashFilenames, httpFilenames []string, args []string, executeIfExists bool) (int, error) {

	returnStatus := EXECUTESTATUS_NOACTIONFOUND
	var returnError error

	// lets start with the bashfiles
	for _, bashFilename := range bashFilenames {
		resource, err := getKeptnResource(myKeptn, bashFilename)
		if resource != "" && err == nil {
			myKeptn.Logger.Debug("Found script " + bashFilename)

			// if we are just here to see whether we found something to execute then lets return this status
			if !executeIfExists {
				return EXECUTESTATUS_ACTIONFOUND, nil
			}

			var output string
			output, err = executeCommandWithKeptnContext(resource, args, keptnEvent, nil)
			if err != nil {
				myKeptn.Logger.Error(fmt.Sprintf("Error executing script: %s", err.Error()))
				returnStatus = EXECUTESTATUS_FAILED
				returnError = err
			} else {
				myKeptn.Logger.Info("Script output: " + output)
				returnStatus = EXECUTESTATUS_SUCCESSFULL
			}
		} else {
			myKeptn.Logger.Debug("No script found at " + bashFilename)
		}
	}

	// now we iterate through the http files
	for _, httpFilename := range httpFilenames {
		var parsedRequest genericHttpRequest
		resource, err := getKeptnResource(myKeptn, httpFilename)
		if resource != "" && err == nil {
			myKeptn.Logger.Debug("Found http request " + httpFilename)

			// if we are just here to see whether we found something to execute then lets return this status
			if !executeIfExists {
				return EXECUTESTATUS_ACTIONFOUND, nil
			}

			parsedRequest, err = parseHttpRequestFromHttpTextFile(keptnEvent, httpFilename)
			if err == nil {
				statusCode, body, requestError := executeGenericHttpRequest(parsedRequest)
				if requestError != nil {
					myKeptn.Logger.Error(fmt.Sprintf("Error: %s", requestError.Error()))
					returnStatus = EXECUTESTATUS_FAILED
					returnError = requestError
				} else {
					if statusCode <= 299 {
						returnStatus = EXECUTESTATUS_SUCCESSFULL
					} else {
						returnStatus = EXECUTESTATUS_FAILED
						returnError = fmt.Errorf("HTTP Call returned status code: %d", statusCode)
					}
					myKeptn.Logger.Info(fmt.Sprintf("%d - %s", statusCode, body))
				}
			} else {
				myKeptn.Logger.Error(fmt.Sprintf("Error: %s", err.Error()))
				returnStatus = EXECUTESTATUS_FAILED
			}
		} else {
			myKeptn.Logger.Debug("No http file found at " + httpFilename)
		}
	}

	return returnStatus, returnError
}

//
// Handles an incoming event
//
func _executeScriptOrHttEventHandler(myKeptn *keptn.Keptn, keptnEvent BaseKeptnEvent, data interface{}, filePrefix string, executeIfExists bool) (int, error) {
	// we allow different files to be specified by the end user
	bashFilenames := []string{
		GenericScriptFolderBase + "all.events.sh",
		GenericScriptFolderBase + filePrefix + ".sh",
		GenericScriptFolderBase + "all.events.py",
		GenericScriptFolderBase + filePrefix + ".py",
	}

	httpFilenames := []string{
		GenericScriptFolderBase + "all.events.http",
		GenericScriptFolderBase + filePrefix + ".http",
	}

	// First - lets store the event as a json file on the filesystem
	eventJsonPath := "./tmpdata/"
	eventJsonFileName := eventJsonPath + keptnEvent.context + ".event.json"
	dataAsJson, err := json.Marshal(data)
	if err == nil {

		if _, err := os.Stat(eventJsonFileName); os.IsNotExist(err) {
			os.MkdirAll(eventJsonPath, 0700)
		}

		file, err := os.Create(eventJsonFileName)
		if err == nil {
			file.Write(dataAsJson)
			defer file.Close()
		} else {
			myKeptn.Logger.Error(fmt.Sprintf("Couldnt write %s: %s", eventJsonFileName, err.Error()))
		}
	} else {
		myKeptn.Logger.Error(fmt.Sprintf("Couldnt marshal incoming event to JSON string: %s", err.Error()))
	}

	// ass the event filename as argument
	args := []string{eventJsonFileName}

	// now lets execute these scripts!
	status, err := executeScriptOrHTTP(myKeptn, keptnEvent, data, bashFilenames, httpFilenames, args, executeIfExists)
	if err != nil {
		myKeptn.Logger.Error(fmt.Sprintf("Error: %s", err.Error()))
	}

	return status, err
}

/**
* Here are all the handler functions for the individual event
  See https://github.com/keptn/spec/blob/0.1.3/cloudevents.md for details on the payload

  -> "sh.keptn.event.configuration.change"
  -> "sh.keptn.events.deployment-finished"
  -> "sh.keptn.events.tests-finished"
  -> "sh.keptn.event.start-evaluation"
  -> "sh.keptn.events.evaluation-done"
  -> "sh.keptn.event.problem.open"
	-> "sh.keptn.events.problem"
	-> "sh.keptn.event.action.triggered"
*/

/**
 * Initalizes KeptnBaseEvent from event object and passed values
 */
func initializeKeptnBaseEvent(incomingEvent cloudevents.Event, project, stage, service string, labels map[string]string) BaseKeptnEvent {
	var shkeptncontext string
	incomingEvent.Context.ExtensionAs("shkeptncontext", &shkeptncontext)

	nano := incomingEvent.Time().UTC().UnixNano()
	milli := nano / 1000000

	// create a base Keptn Event
	keptnEvent := BaseKeptnEvent{
		event:     incomingEvent.Type(),
		source:    incomingEvent.Source(),
		context:   shkeptncontext,
		time:      incomingEvent.Time().String(),
		timeutc:   incomingEvent.Time().UTC().String(),
		timeutcms: strconv.FormatInt(milli, 10),
	}

	keptnEvent.project = project
	keptnEvent.stage = stage
	keptnEvent.service = service
	keptnEvent.labels = labels

	return keptnEvent
}

// Handles ConfigureMonitoringEventType = "sh.keptn.event.monitoring.configure"
func HandleConfigureMonitoringEvent(myKeptn *keptn.Keptn, incomingEvent cloudevents.Event, data *keptn.ConfigureMonitoringEventData) error {
	log.Printf("Handling Configure Monitoring Event: %s", incomingEvent.Context.GetID())

	keptnEvent := initializeKeptnBaseEvent(incomingEvent, data.Project, "", data.Service, map[string]string{})

	eventName := "monitoring.configure"
	_, err := _executeScriptOrHttEventHandler(myKeptn, keptnEvent, data, eventName, true)
	return err
}

//
// Handles ConfigurationChangeEventType = "sh.keptn.event.configuration.change"
// TODO: add in your handler code
//
func HandleConfigurationChangeEvent(myKeptn *keptn.Keptn, incomingEvent cloudevents.Event, data *keptn.ConfigurationChangeEventData) error {
	log.Printf("Handling Configuration Changed Event: %s", incomingEvent.Context.GetID())

	keptnEvent := initializeKeptnBaseEvent(incomingEvent, data.Project, data.Stage, data.Service, data.Labels)

	eventName := "configuration.change"
	_, err := _executeScriptOrHttEventHandler(myKeptn, keptnEvent, data, eventName, true)
	return err
}

//
// Handles DeploymentFinishedEventType = "sh.keptn.events.deployment-finished"
// TODO: add in your handler code
//
func HandleDeploymentFinishedEvent(myKeptn *keptn.Keptn, incomingEvent cloudevents.Event, data *keptn.DeploymentFinishedEventData) error {
	log.Printf("Handling Deployment Finished Event: %s", incomingEvent.Context.GetID())

	keptnEvent := initializeKeptnBaseEvent(incomingEvent, data.Project, data.Stage, data.Service, data.Labels)
	keptnEvent.deployment = data.DeploymentStrategy
	keptnEvent.testStrategy = data.TestStrategy
	keptnEvent.deploymentURILocal = data.DeploymentURILocal
	keptnEvent.deploymentURIPublic = data.DeploymentURIPublic
	keptnEvent.image = data.Image
	keptnEvent.tag = data.Tag

	eventName := "deployment.finished"
	_, err := _executeScriptOrHttEventHandler(myKeptn, keptnEvent, data, eventName, true)
	return err
}

//
// Handles TestsFinishedEventType = "sh.keptn.events.tests-finished"
// TODO: add in your handler code
//
func HandleTestsFinishedEvent(myKeptn *keptn.Keptn, incomingEvent cloudevents.Event, data *keptn.TestsFinishedEventData) error {
	log.Printf("Handling Tests Finished Event: %s", incomingEvent.Context.GetID())

	keptnEvent := initializeKeptnBaseEvent(incomingEvent, data.Project, data.Stage, data.Service, data.Labels)
	keptnEvent.deployment = data.DeploymentStrategy
	keptnEvent.testStrategy = data.TestStrategy
	keptnEvent.testStart = data.Start
	keptnEvent.testEnd = data.End

	eventName := "tests.finished"
	_, err := _executeScriptOrHttEventHandler(myKeptn, keptnEvent, data, eventName, true)
	return err
}

//
// Handles EvaluationDoneEventType = "sh.keptn.event.start-evaluation"
// TODO: add in your handler code
//
func HandleStartEvaluationEvent(myKeptn *keptn.Keptn, incomingEvent cloudevents.Event, data *keptn.StartEvaluationEventData) error {
	log.Printf("Handling Start Evaluation Event: %s", incomingEvent.Context.GetID())

	keptnEvent := initializeKeptnBaseEvent(incomingEvent, data.Project, data.Stage, data.Service, data.Labels)
	keptnEvent.deployment = data.DeploymentStrategy
	keptnEvent.testStrategy = data.TestStrategy
	keptnEvent.evaluationStart = data.Start
	keptnEvent.evaluationEnd = data.End

	eventName := "start.evaluation"
	_, err := _executeScriptOrHttEventHandler(myKeptn, keptnEvent, data, eventName, true)
	return err
}

//
// Handles DeploymentFinishedEventType = "sh.keptn.events.evaluation-done"
// TODO: add in your handler code
//
func HandleEvaluationDoneEvent(myKeptn *keptn.Keptn, incomingEvent cloudevents.Event, data *keptn.EvaluationDoneEventData) error {
	log.Printf("Handling Evaluation Done Event: %s", incomingEvent.Context.GetID())

	keptnEvent := initializeKeptnBaseEvent(incomingEvent, data.Project, data.Stage, data.Service, data.Labels)
	keptnEvent.deployment = data.DeploymentStrategy
	keptnEvent.testStrategy = data.TestStrategy
	keptnEvent.evaluationResult = data.Result

	eventName := "evaluation.done"
	_, err := _executeScriptOrHttEventHandler(myKeptn, keptnEvent, data, eventName, true)
	return err
}

//
// Handles InternalGetSLIEventType = "sh.keptn.internal.event.get-sli"
// TODO: add in your handler code
//
func HandleInternalGetSLIEvent(myKeptn *keptn.Keptn, incomingEvent cloudevents.Event, data *keptn.InternalGetSLIEventData) error {
	// log.Printf("Handling Internal Get SLI Event: %s", incomingEvent.Context.GetID())

	return nil
}

//
// Handles ProblemOpenEventType = "sh.keptn.event.problem.open"
// Handles ProblemEventType = "sh.keptn.events.problem"
// TODO: add in your handler code
//
func HandleProblemEvent(myKeptn *keptn.Keptn, incomingEvent cloudevents.Event, data *keptn.ProblemEventData) error {
	log.Printf("Handling Problem Event: %s", incomingEvent.Context.GetID())

	// Deprecated since Keptn 0.7.0 - use the HandleActionTriggeredEvent instead

	return nil
}

//
// Handles ActionTriggeredEventType = "sh.keptn.event.action.triggered"
// TODO: add in your handler code
//
func HandleActionTriggeredEvent(myKeptn *keptn.Keptn, incomingEvent cloudevents.Event, data *keptn.ActionTriggeredEventData) error {
	log.Printf("Handling Action Triggered Event: %s", incomingEvent.Context.GetID())

	keptnEvent := initializeKeptnBaseEvent(incomingEvent, data.Project, data.Stage, data.Service, data.Labels)
	keptnEvent.action = data.Action.Action
	keptnEvent.problemState = data.Problem.State
	keptnEvent.problemID = data.Problem.ProblemID
	keptnEvent.problemTitle = data.Problem.ProblemTitle
	keptnEvent.pid = data.Problem.PID
	keptnEvent.problemURL = data.Problem.ProblemURL

	eventName := "action.triggered." + data.Action.Action
	status, err := _executeScriptOrHttEventHandler(myKeptn, keptnEvent, data, eventName, false)

	// lets see if we have found an action - if so - lets notify keptn - then execute it - and the report when we are done
	if status == EXECUTESTATUS_ACTIONFOUND {

		// Step 1: Send an ActionStartedEvent
		actionStartedEventData := &keptn.ActionStartedEventData{}
		err := incomingEvent.DataAs(actionStartedEventData)
		if err != nil {
			log.Printf("Got Data Error: %s", err.Error())
			return err
		}

		err = myKeptn.SendActionStartedEvent(&incomingEvent, data.Labels, "generic-executor-service")
		if err != nil {
			log.Printf("Got Error From SendActionStartedEvent: %s", err.Error())
			return err
		}

		// Step 2: Lets execute the Script
		var actionResult keptn.ActionResult
		status, err = _executeScriptOrHttEventHandler(myKeptn, keptnEvent, data, eventName, true)
		if status == EXECUTESTATUS_SUCCESSFULL {
			log.Printf("Successful execution of action")
			actionResult.Result = keptn.ActionResultPass
			actionResult.Status = keptn.ActionStatusSucceeded
		} else {
			log.Printf("Execution of action failed")
			actionResult.Result = keptn.ActionResultFailed
			actionResult.Status = keptn.ActionStatusErrored
		}

		// STep 3: Send an action finished

		myKeptn.SendActionFinishedEvent(&incomingEvent, actionResult, data.Labels, "generic-executor-service")
		if err != nil {
			log.Printf("Got Error From SendActionFinishedEvent: %s", err.Error())
			return err
		}
	}

	return err
}
