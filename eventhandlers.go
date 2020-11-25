package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

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

/**
 * executeScriptOrHTTP
 * This method will iterate through the bash and http filenames. If the filename is found it will execute it if executeIfExits==true.
 * If onlyFirstMatch==true this method will stop after it has found the first matching filename. Otherwise it will continue finding more matches. This would for instance allow executing multiple scripts such as mypspecialscript.ph and all.events.sh
 * When calling a bashscript the args are passed as arguments to that script
 * Parameters
 * @bashfilename: filenames that the service tries to locate. This can either be shell scripts python scripts or anything else that is executable in the container
 * @httpFilenames: filenames of http scripts
 * @args: List of command line argumetns passed to the bashfiles
 * @executeIfExists: if true and a script is found it will be executed - otherwise it just returns EXECUTESTATUS_ACTIONFOUND
 * @onlyFirstMatch: if true will only execute the first matching script - otherwise it will keep looking for more matches
 *
 * Return:
 * int: status, e.g: EXECUTESTATUS_XXX
 * string: name of the first matched script
 * string: return payload
 * error: any error that may have occured
 */
// if any of the passed files exist either executes the bash or the http request
// The return status depends on the success of the executed script or HTTP Request. If the script fails or if the HTTP call returns a status code >= 300 the call is considered failed
//
func executeScriptOrHTTP(myKeptn *keptn.Keptn, keptnEvent BaseKeptnEvent, data interface{}, bashFilenames, httpFilenames []string, args []string, executeIfExists bool, onlyFirstMatch bool) (int, string, string, error) {

	returnStatus := EXECUTESTATUS_NOACTIONFOUND
	var returnError error
	executedScript := ""
	returnPayload := ""

	// we will prefix any downloaded file with the keptn context to make sure we dont have any colisions
	uniqueFilePrefix := keptnEvent.context

	// lets start with the bashfiles
	for _, bashFilename := range bashFilenames {
		resource, err := getKeptnResource(myKeptn, bashFilename, uniqueFilePrefix)
		if resource != "" && err == nil {
			myKeptn.Logger.Debug("Found script " + bashFilename)

			executedScript = bashFilename
			// if we are just here to see whether we found something to execute then lets return this status
			if !executeIfExists {
				removeFiles(myKeptn, []string{resource})
				return EXECUTESTATUS_ACTIONFOUND, executedScript, returnPayload, nil
			}

			// if this is a python script our command is actually python3 and we pass the script as an argument
			argsToUse := args
			executable := resource
			if strings.Index(bashFilename, ".py") > 0 {
				argsToUse = append([]string{resource}, args...)
				executable = "python3"
			}

			// Lets execute it
			var output string
			output, err = executeCommandWithKeptnContext(executable, argsToUse, keptnEvent, nil)
			if err != nil {
				myKeptn.Logger.Error(fmt.Sprintf("Error executing script: %s", err.Error()))
				returnStatus = EXECUTESTATUS_FAILED
				returnError = err
			} else {
				myKeptn.Logger.Info("Script output: " + output)
				returnPayload = output
				returnStatus = EXECUTESTATUS_SUCCESSFULL
				returnError = nil
			}

			// lets make sure remove that file again after we executed it
			removeFiles(myKeptn, []string{resource})

			// only first match?
			if onlyFirstMatch {
				return returnStatus, executedScript, returnPayload, returnError
			}

		} else {
			myKeptn.Logger.Debug("No script found at " + bashFilename)
			if err != nil {
				myKeptn.Logger.Debug("err " + err.Error())
			}
		}
	}

	// now we iterate through the http files
	for _, httpFilename := range httpFilenames {
		var parsedRequest genericHttpRequest
		resource, err := getKeptnResource(myKeptn, httpFilename, uniqueFilePrefix)
		if resource != "" && err == nil {
			myKeptn.Logger.Debug("Found http request " + httpFilename)

			executedScript = httpFilename

			// if we are just here to see whether we found something to execute then lets return this status
			if !executeIfExists {
				removeFiles(myKeptn, []string{resource})
				return EXECUTESTATUS_ACTIONFOUND, executedScript, returnPayload, nil
			}

			// Lets execute it
			parsedRequest, err = parseHttpRequestFromHttpTextFile(keptnEvent, httpFilename)
			if err == nil {
				statusCode, body, requestError := executeGenericHttpRequest(parsedRequest)
				if requestError != nil {
					myKeptn.Logger.Error(fmt.Sprintf("Error: %s", requestError.Error()))
					returnStatus = EXECUTESTATUS_FAILED
					returnError = requestError
				} else {
					returnPayload = body
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

			// lets make sure remove that file again after we executed it
			removeFiles(myKeptn, []string{resource})

			// only first match?
			if onlyFirstMatch {
				return returnStatus, executedScript, returnPayload, returnError
			}
		} else {
			myKeptn.Logger.Debug("No http file found at " + httpFilename)
		}
	}

	return returnStatus, executedScript, returnPayload, returnError
}

/*
* Please see documentation of executeScriptOrHTTP as this method calls executeScriptOrHTTP with a list of files it should look for in the config repo
* Return:
* int: status, e.g: EXECUTESTATUS_XXX
* string: name of the first matched script
* string: return payload
* error: any error that may have occured
 */
func _executeScriptOrHttEventHandler(myKeptn *keptn.Keptn, keptnEvent BaseKeptnEvent, data interface{}, filePrefix string, executeIfExists bool, onlyFirstMatch bool) (int, string, string, error) {
	// we allow different files to be specified by the end user - we first look for the more specific ones that include the file name
	bashFilenames := []string{
		GenericScriptFolderBase + filePrefix + ".sh",
		GenericScriptFolderBase + filePrefix + ".py",
		GenericScriptFolderBase + "all.events.sh",
		GenericScriptFolderBase + "all.events.py",
	}

	httpFilenames := []string{
		GenericScriptFolderBase + filePrefix + ".http",
		GenericScriptFolderBase + "all.events.http",
	}

	//
	// First - lets store the event as a json file on the filesystem as we are passing it to the script as an argument
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

	// pass the event filename as argument
	args := []string{eventJsonFileName}

	// now lets execute these scripts!
	status, scriptName, responsePayload, err := executeScriptOrHTTP(myKeptn, keptnEvent, data, bashFilenames, httpFilenames, args, executeIfExists, onlyFirstMatch)
	if err != nil {
		myKeptn.Logger.Error(fmt.Sprintf("Error: %s", err.Error()))
	}

	// delete the temp data file
	os.Remove(eventJsonFileName)

	// cleanup the scriptName - remove the GenericScriptFolderBase
	scriptName = strings.TrimPrefix(scriptName, GenericScriptFolderBase)

	return status, scriptName, responsePayload, err
}

/**
 * Analyzes the response payload and e.g: sends a keptn event, raise an error, ...
 * TODO: define what payload can contain
 */
func handleScriptResponsePayload(myKeptn *keptn.Keptn, keptnEvent BaseKeptnEvent, responsePayload string) error {

	return nil
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
	// log.Printf("Handling Configure Monitoring Event: %s", incomingEvent.Context.GetID())

	keptnEvent := initializeKeptnBaseEvent(incomingEvent, data.Project, "", data.Service, map[string]string{})

	eventName := "monitoring.configure"
	_, _, payload, err := _executeScriptOrHttEventHandler(myKeptn, keptnEvent, data, eventName, true, false)
	if err != nil {
		return err
	}

	err = handleScriptResponsePayload(myKeptn, keptnEvent, payload)
	return err
}

//
// Handles ConfigurationChangeEventType = "sh.keptn.event.configuration.change"
// TODO: add in your handler code
//
func HandleConfigurationChangeEvent(myKeptn *keptn.Keptn, incomingEvent cloudevents.Event, data *keptn.ConfigurationChangeEventData) error {
	// log.Printf("Handling Configuration Changed Event: %s", incomingEvent.Context.GetID())

	keptnEvent := initializeKeptnBaseEvent(incomingEvent, data.Project, data.Stage, data.Service, data.Labels)

	eventName := "configuration.change"
	_, _, payload, err := _executeScriptOrHttEventHandler(myKeptn, keptnEvent, data, eventName, true, false)
	if err != nil {
		return err
	}

	err = handleScriptResponsePayload(myKeptn, keptnEvent, payload)
	return err
}

//
// Handles DeploymentFinishedEventType = "sh.keptn.events.deployment-finished"
// TODO: add in your handler code
//
func HandleDeploymentFinishedEvent(myKeptn *keptn.Keptn, incomingEvent cloudevents.Event, data *keptn.DeploymentFinishedEventData) error {
	// log.Printf("Handling Deployment Finished Event: %s", incomingEvent.Context.GetID())

	keptnEvent := initializeKeptnBaseEvent(incomingEvent, data.Project, data.Stage, data.Service, data.Labels)
	keptnEvent.deployment = data.DeploymentStrategy
	keptnEvent.testStrategy = data.TestStrategy
	keptnEvent.deploymentURILocal = data.DeploymentURILocal
	keptnEvent.deploymentURIPublic = data.DeploymentURIPublic
	keptnEvent.image = data.Image
	keptnEvent.tag = data.Tag

	eventName := "deployment.finished"
	_, _, payload, err := _executeScriptOrHttEventHandler(myKeptn, keptnEvent, data, eventName, true, false)
	if err != nil {
		return err
	}

	err = handleScriptResponsePayload(myKeptn, keptnEvent, payload)
	return err
}

//
// Handles TestsFinishedEventType = "sh.keptn.events.tests-finished"
// TODO: add in your handler code
//
func HandleTestsFinishedEvent(myKeptn *keptn.Keptn, incomingEvent cloudevents.Event, data *keptn.TestsFinishedEventData) error {
	// log.Printf("Handling Tests Finished Event: %s", incomingEvent.Context.GetID())

	keptnEvent := initializeKeptnBaseEvent(incomingEvent, data.Project, data.Stage, data.Service, data.Labels)
	keptnEvent.deployment = data.DeploymentStrategy
	keptnEvent.testStrategy = data.TestStrategy
	keptnEvent.testStart = data.Start
	keptnEvent.testEnd = data.End

	eventName := "tests.finished"
	_, _, payload, err := _executeScriptOrHttEventHandler(myKeptn, keptnEvent, data, eventName, true, false)
	if err != nil {
		return err
	}

	err = handleScriptResponsePayload(myKeptn, keptnEvent, payload)
	return err
}

//
// Handles EvaluationDoneEventType = "sh.keptn.event.start-evaluation"
// TODO: add in your handler code
//
func HandleStartEvaluationEvent(myKeptn *keptn.Keptn, incomingEvent cloudevents.Event, data *keptn.StartEvaluationEventData) error {
	// log.Printf("Handling Start Evaluation Event: %s", incomingEvent.Context.GetID())

	keptnEvent := initializeKeptnBaseEvent(incomingEvent, data.Project, data.Stage, data.Service, data.Labels)
	keptnEvent.deployment = data.DeploymentStrategy
	keptnEvent.testStrategy = data.TestStrategy
	keptnEvent.evaluationStart = data.Start
	keptnEvent.evaluationEnd = data.End

	eventName := "start.evaluation"
	_, _, payload, err := _executeScriptOrHttEventHandler(myKeptn, keptnEvent, data, eventName, true, false)
	if err != nil {
		return err
	}

	err = handleScriptResponsePayload(myKeptn, keptnEvent, payload)
	return err
}

//
// Handles DeploymentFinishedEventType = "sh.keptn.events.evaluation-done"
// TODO: add in your handler code
//
func HandleEvaluationDoneEvent(myKeptn *keptn.Keptn, incomingEvent cloudevents.Event, data *keptn.EvaluationDoneEventData) error {
	// log.Printf("Handling Evaluation Done Event: %s", incomingEvent.Context.GetID())

	keptnEvent := initializeKeptnBaseEvent(incomingEvent, data.Project, data.Stage, data.Service, data.Labels)
	keptnEvent.deployment = data.DeploymentStrategy
	keptnEvent.testStrategy = data.TestStrategy
	keptnEvent.evaluationResult = data.Result

	eventName := "evaluation.done"
	_, _, payload, err := _executeScriptOrHttEventHandler(myKeptn, keptnEvent, data, eventName, true, false)
	if err != nil {
		return err
	}

	err = handleScriptResponsePayload(myKeptn, keptnEvent, payload)
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
	// log.Printf("Handling Problem Event: %s", incomingEvent.Context.GetID())

	// Deprecated since Keptn 0.7.0 - use the HandleActionTriggeredEvent instead

	return nil
}

//
// Handles ActionTriggeredEventType = "sh.keptn.event.action.triggered"
// TODO: add in your handler code
//
func HandleActionTriggeredEvent(myKeptn *keptn.Keptn, incomingEvent cloudevents.Event, data *keptn.ActionTriggeredEventData) error {
	// log.Printf("Handling Action Triggered Event: %s", incomingEvent.Context.GetID())

	keptnEvent := initializeKeptnBaseEvent(incomingEvent, data.Project, data.Stage, data.Service, data.Labels)
	keptnEvent.action = data.Action.Action
	keptnEvent.problemState = data.Problem.State
	keptnEvent.problemID = data.Problem.ProblemID
	keptnEvent.problemTitle = data.Problem.ProblemTitle
	keptnEvent.pid = data.Problem.PID
	keptnEvent.problemURL = data.Problem.ProblemURL

	// Get remediation values from data.
	keptnEvent.remediationValues = make(map[string]string)
	values, ok := data.Action.Value.(map[string]interface{})
	if ok {
		for keyValue, valueValue := range values {
			keptnEvent.remediationValues[keyValue] = fmt.Sprintf("%v", valueValue)
			log.Println(keyValue + ": " + keptnEvent.remediationValues[keyValue])
		}
	}

	eventName := "action.triggered." + data.Action.Action
	status, scriptName, payload, err := _executeScriptOrHttEventHandler(myKeptn, keptnEvent, data, eventName, false, true)

	// lets see if we have found an action - if so - lets notify keptn - then execute it - and the report when we are done
	if status == EXECUTESTATUS_ACTIONFOUND {

		// Step 1: Send an ActionStartedEvent
		actionStartedEventData := &keptn.ActionStartedEventData{}
		err := incomingEvent.DataAs(actionStartedEventData)
		if err != nil {
			log.Printf("Got Data Error: %s", err.Error())
			return err
		}

		// making sure we pass Problem URL, Action and Executor Script as labels
		if data.Labels == nil {
			data.Labels = make(map[string]string)
		}
		data.Labels["Problem URL"] = keptnEvent.problemURL
		if scriptName != "" {
			data.Labels[data.Action.Action] = scriptName
		}

		err = myKeptn.SendActionStartedEvent(&incomingEvent, data.Labels, "generic-executor-service")
		if err != nil {
			log.Printf("Got Error From SendActionStartedEvent: %s", err.Error())
			return err
		}

		// Step 2: Lets execute the Script
		var actionResult keptn.ActionResult
		status, scriptName, payload, err = _executeScriptOrHttEventHandler(myKeptn, keptnEvent, data, eventName, true, false)
		if status == EXECUTESTATUS_SUCCESSFULL {
			log.Printf("Successful execution of action")
			actionResult.Result = keptn.ActionResultPass
			actionResult.Status = keptn.ActionStatusSucceeded
		} else {
			log.Printf("Execution of action failed")
			actionResult.Result = keptn.ActionResultFailed
			actionResult.Status = keptn.ActionStatusErrored
		}

		// we allow the script to pass back the actual result, e.g: the output
		if payload != "" {
			actionResult.Status = keptn.ActionStatusType(payload)
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
