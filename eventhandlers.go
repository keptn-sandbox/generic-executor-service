package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2" // make sure to use v2 cloudevents here
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

// GenericScriptFolderBase Folder in the Keptn GitHub Repo where we expect scripts and http files
const GenericScriptFolderBase = "generic-executor/"

func findAndStoreScriptFile(myKeptn *keptnv2.Keptn, filePrefix string, uniquePrefix string) (string, error) {
	// we allow different files to be specified by the end user - we first look for the more specific ones that include the file name
	allowedFilenames := []string{
		GenericScriptFolderBase + filePrefix + ".sh",
		GenericScriptFolderBase + filePrefix + ".py",
		GenericScriptFolderBase + filePrefix + ".http",
		GenericScriptFolderBase + "all.events.sh",
		GenericScriptFolderBase + "all.events.py",
		GenericScriptFolderBase + "all.events.http",
	}

	// iterate over all files in that order
	for _, filename := range allowedFilenames {
		resourceFilename, err := getKeptnResource(myKeptn, filename, uniquePrefix)

		if resourceFilename != "" && err == nil {
			log.Printf("Found script %s and stored it as %s", filename, resourceFilename)

			return resourceFilename, nil
		} else {
			log.Printf("%s not found: %s", filename, err.Error())
		}
	}

	return "", fmt.Errorf("No file found")
}

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
func executeScriptOrHTTP(scriptFileName string, incomingEvent cloudevents.Event) (string, keptnv2.ResultType, keptnv2.StatusType, error) {

	if strings.HasSuffix(scriptFileName, ".http") {
		// Execute HTTP Test
		parsedRequest, err := parseHttpRequestFromHttpTextFile(scriptFileName, incomingEvent)

		if err != nil {
			return "", keptnv2.ResultFailed, keptnv2.StatusErrored, fmt.Errorf("Failed to parse %s: %s", scriptFileName, err.Error())
		}

		statusCode, body, requestError := executeGenericHttpRequest(parsedRequest)

		if requestError != nil {
			// request errored
			return "", keptnv2.ResultFailed, keptnv2.StatusErrored, requestError
		}
		if statusCode >= 200 && statusCode <= 299 {
			// http status 2xx is suggesting that everything is fine
			return body, keptnv2.ResultPass, keptnv2.StatusSucceeded, nil
		}

		// last but not least: status code != 2xx suggests that something went wrong on the other side
		log.Printf("HTTP Call returned status code %d", statusCode)
		return body, keptnv2.ResultFailed, keptnv2.StatusSucceeded, nil
	}
	// else: execute the script using bash or python

	// store event in file
	eventJSONFileName, err := storeCloudEventInFile(incomingEvent)

	if err != nil {
		return "", keptnv2.ResultFailed, keptnv2.StatusErrored, err
	}
	defer os.Remove(eventJSONFileName)

	var executable string
	var argsToUse []string

	// check if script ends with .py
	if strings.HasSuffix(scriptFileName, ".py") {
		executable = "python3"
		argsToUse = []string{scriptFileName, eventJSONFileName}
	} else if strings.HasSuffix(scriptFileName, ".sh") {
		executable = "bash"
		argsToUse = []string{scriptFileName, eventJSONFileName}
	} else {
		// invalid filename found
		return "", keptnv2.ResultFailed, keptnv2.StatusErrored, fmt.Errorf("Unhandled extension for file %s", scriptFileName)
	}

	// Lets execute it
	output, err := executeCommandWithKeptnContext(executable, argsToUse, incomingEvent, nil)

	if err != nil {
		// return a failed result
		return output, keptnv2.ResultFailed, keptnv2.StatusSucceeded, err
	}

	return output, keptnv2.ResultPass, keptnv2.StatusSucceeded, nil
}

func storeCloudEventInFile(incomingEvent cloudevents.Event) (string, error) {
	// First - lets store the event as a json file on the filesystem as we are passing it to the script as an argument

	eventJSONFileName := fmt.Sprintf("%s.event.json", incomingEvent.ID())

	// marshal incomingEvent
	dataAsJSON, err := json.Marshal(incomingEvent)

	if err != nil {
		log.Printf("Couldn't marshal incoming event to JSON string: %s", err.Error())
		return "", err
	}

	file, err := os.Create(eventJSONFileName)

	file.Write(dataAsJSON)
	defer file.Close()

	return eventJSONFileName, nil
}

// HandleResponsePayload tries to parse the response of a command as json and returns it
func HandleResponsePayload(responsePayload string) (map[string]interface{}, error) {
	// no payload or not json?
	if responsePayload == "" || !strings.HasPrefix(responsePayload, "{") {
		return nil, nil
	}

	// parse response
	parsedResponse := map[string]interface{}{}
	err := json.Unmarshal([]byte(responsePayload), &parsedResponse)
	if err != nil {
		return nil, err
	}

	// Check for error
	if errorValue, errorOk := parsedResponse["error"]; errorOk {
		return nil, errors.New(errorValue.(string))
	}

	return parsedResponse, nil
}

// GenericCloudEventsHandler handles all cloud-events by looking up a script-file and executing it
func GenericCloudEventsHandler(myKeptn *keptnv2.Keptn, incomingEvent cloudevents.Event, data interface{}) error {
	log.Printf("Handling %s Event: %s", incomingEvent.Type(), incomingEvent.Context.GetID())
	log.Printf("CloudEvent %T: %v", data, data)

	// check if the status type is either 'triggered', 'started', or 'finished'
	split := strings.Split(incomingEvent.Type(), ".")

	if len(split) < 3 {
		return fmt.Errorf("Failed to split event of type %s", incomingEvent.Type())
	}

	// split incoming event by dots and separate it into statusType, taskSequencename and stageName
	statusType := split[len(split)-1]
	taskName := split[len(split)-2]
	log.Printf("task=%s,status=%s", taskName, statusType)

	// list of all filenames we want to check
	eventNamesToExecute := []string{}

	// Special Handling for action.triggered - we need to extract the action name because we will be looking for files called action.triggered.actionname.XX
	if incomingEvent.Type() == keptnv2.GetTriggeredEventType(keptnv2.ActionTaskName) {
		actionTriggeredEvent := &keptnv2.ActionTriggeredEventData{}
		err := incomingEvent.DataAs(actionTriggeredEvent)
		if err != nil {
			return fmt.Errorf("Failed to parse action triggered event %s", incomingEvent.Type())
		}

		eventNamesToExecute = append(eventNamesToExecute, fmt.Sprintf("%s.%s.%s", taskName, statusType, actionTriggeredEvent.Action.Action))
	}

	// by default we always check for taskName.statusType, e.g: test.triggered
	eventNamesToExecute = append(eventNamesToExecute, fmt.Sprintf("%s.%s", taskName, statusType))

	// prefix for storing filenames
	uniquePrefix := incomingEvent.Context.GetID()

	// if this is a triggered event we will be sending out start & finished events - otherwise not as we assume we just handle these events for notification purposes
	sendStartFinishedEvents := true
	_, eventErr := keptnv2.GetEventTypeForTriggeredEvent(incomingEvent.Type(), "")
	if eventErr != nil {
		sendStartFinishedEvents = false
		log.Printf("Not sending start/finished event as %s is not a triggered event!", incomingEvent.Type())
	}

	// now we iterate through all eventName we want to look for scripts for
	for _, eventName := range eventNamesToExecute {

		// Check if a suitable script/... exists; exit if not
		scriptFileName, err := findAndStoreScriptFile(myKeptn, eventName, uniquePrefix)

		if err != nil {
			// not found -> ignore this event
			log.Printf("Ignoring event %s as no suitable file was found", eventName)
			continue
			// return err
		}

		// Script exists -> Send task.started event in case we are handling a triggered event
		if sendStartFinishedEvents {
			_, err = myKeptn.SendTaskStartedEvent(&keptnv2.EventData{
				Message: fmt.Sprintf("Found script %s", scriptFileName),
			}, ServiceName)

			if err != nil {
				log.Printf("Failed to send task.started event: %s", err.Error())
				return err
			}
		}

		// Finally Executing the Script
		log.Printf("Executing %s", scriptFileName)
		response, result, status, err := executeScriptOrHTTP(scriptFileName, incomingEvent)

		if err != nil {

			log.Printf("Script execution failed: %s", err.Error())

			if sendStartFinishedEvents {
				// script execution failed - send finished event
				_, err = myKeptn.SendTaskFinishedEvent(&keptnv2.EventData{
					Status:  status,
					Result:  result,
					Message: fmt.Sprintf("Failed to execute %s: %s", scriptFileName, err.Error()),
				}, ServiceName)
			}

			return err
		} else {
			log.Printf("Script execution successful: %s, %s", result, status)
			if VerboseLogging {
				log.Printf(response)
			}
		}

		// parse the response
		responseJSON, err := HandleResponsePayload(response)

		if err != nil {
			// failed to parse response payload
			return handleError(myKeptn, err)
		}

		if sendStartFinishedEvents {
			// finally send a task.finished event
			responseCloudEvent := &keptnv2.EventData{
				Status:  status,
				Result:  result,
				Message: response,
			}

			if responseJSON != nil {
				log.Printf("Script returned JSON properties for finished event: %v", responseJSON)

				// convert the event to a map[string]interface{} to set the result of the operation as a property of the outgoing event
				responseEventMap := map[string]interface{}{}
				if err := keptnv2.Decode(responseCloudEvent, responseEventMap); err != nil {
					return handleError(myKeptn, err)
				}

				// set the responseJSON to e.g: "test" when handling the test task
				responseEventMap[taskName] = responseJSON

				if err := keptnv2.Decode(responseEventMap, responseCloudEvent); err != nil {
					return handleError(myKeptn, err)
				}
			}

			_, err = myKeptn.SendTaskFinishedEvent(responseCloudEvent, ServiceName)
		}

	} // eventNamesToExecute

	log.Printf("Done executing scripts!")

	return nil
}

func handleError(myKeptn *keptnv2.Keptn, err error) error {

	log.Printf("handleError: %s", err.Error())

	_, err = myKeptn.SendTaskFinishedEvent(&keptnv2.EventData{
		Status:  keptnv2.StatusSucceeded,
		Result:  keptnv2.ResultWarning,
		Message: fmt.Sprintf("Failed to parse response: %s", err.Error()),
	}, ServiceName)

	return err
}
