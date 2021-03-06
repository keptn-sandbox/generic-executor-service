package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2" // make sure to use v2 cloudevents here
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

type genericHttpRequest struct {
	method  string
	uri     string
	headers map[string]string
	body    string
}

/**
 * Removes the list of files
 */
func removeFiles(filesToRemove []string) {

	for _, fileName := range filesToRemove {

		err := os.Remove(fileName)
		if err != nil {
			log.Printf("Error removing file: %s", err.Error())
		}
	}
}

/**
 * Retrieves a resource (=file) from the keptn configuration repo and returns its content
 */
func getKeptnResource(myKeptn *keptnv2.Keptn, resource string, uniquePrefix string) (string, error) {
	resourceHandler := myKeptn.ResourceHandler

	// local filesystem?
	if myKeptn.UseLocalFileSystem {
		if _, err := os.Stat(resource); err == nil {
			return resource, nil
		} else {
			return "", err
		}
	}

	// SERVICE-LEVEL: lets try to find it on service level
	requestedResource, err := resourceHandler.GetServiceResource(myKeptn.Event.GetProject(), myKeptn.Event.GetStage(), myKeptn.Event.GetService(), resource)

	if err != nil || requestedResource.ResourceContent == "" {
		// STAGE-LEVEL: not found on service level - lets search one level up on stage level
		requestedResource, err = resourceHandler.GetStageResource(myKeptn.Event.GetProject(), myKeptn.Event.GetStage(), resource)
		if err != nil || requestedResource.ResourceContent == "" {
			// PROJECT-LEVEL: not found on the stage level - lets search one level up on project level
			requestedResource, err = resourceHandler.GetProjectResource(myKeptn.Event.GetProject(), resource)

			if err != nil || requestedResource.ResourceContent == "" {
				return "", err
			}

			myKeptn.Logger.Debug("Found " + resource + " on project level")
		} else {
			myKeptn.Logger.Debug("Found " + resource + " on stage level")
		}
	} else {
		myKeptn.Logger.Debug("Found " + resource + " on service level")
	}

	targetFileName := fmt.Sprintf("%s/%s", uniquePrefix, resource)

	// now store that file on the same directory structure locally
	os.RemoveAll(targetFileName)
	pathArr := strings.Split(targetFileName, "/")
	directory := ""
	for _, pathItem := range pathArr[0 : len(pathArr)-1] {
		directory += pathItem + "/"
	}

	if directory != "" {
		err = os.MkdirAll(directory, os.ModePerm)
		if err != nil {
			return "", err
		}
	}
	resourceFile, err := os.Create(targetFileName)
	if err != nil {
		fmt.Errorf(err.Error())
		return "", err
	}
	defer resourceFile.Close()

	_, err = resourceFile.Write([]byte(requestedResource.ResourceContent))

	if err != nil {
		fmt.Errorf(err.Error())
		return "", err
	}

	return targetFileName, nil
}

//
// Iterates through all elements in the incomingEvent
// First,
// It will replace references in the input string, e.g: ${data.project} will be replaced with the project field
// Second,
// The function also returns an array that can be used to pass env variables, e.g: DATA_PROJECT=project

// $TIMESTRING, $TIMEUTCSTRING, $TIMEUTCMS
// $CONTEXT, $EVENT, $SOURCE
// $PROJECT, $STAGE, $SERVICE
// $DEPLOYMENT, $TESTSTRATEGY
// $DEPLOYMENTURILOCAL, $DEPLOYMENTURIPUBLIC
// $LABEL.XXXX  -> will replace that with a label called XXXX
// $ENV.XXXX    -> will replace that with an env variable called XXXX
// $SECRET.YYYY -> will replace that with the k8s secret called YYYY
//
func manageKeptnPlaceholders(input string, incomingEvent cloudevents.Event) (string, []string) {

	result := input

	// we have three special values that we provide for convenience to better integrate with tools that need a timestamp in UTC and UTC (ms)
	nano := incomingEvent.Time().UTC().UnixNano()
	milli := nano / 1000000
	time := incomingEvent.Time().String()
	timeutc := incomingEvent.Time().UTC().String()
	timeutcms := strconv.FormatInt(milli, 10)

	envArray := []string{
		fmt.Sprintf("TIMESTRING=%s", time),
		fmt.Sprintf("TIMEUTCSTRING=%s", timeutc),
		fmt.Sprintf("TIMEUTCMS=%s", timeutcms),
	}
	result = strings.Replace(result, "${timestring}", time, -1)
	result = strings.Replace(result, "${timeutcstring}", timeutc, -1)
	result = strings.Replace(result, "${timeutcms}", timeutcms, -1)

	// First, we iterate through all our env-variables
	// TODO: Passing our own env variables is mega dangerous, this needs to be adapted when we have proper secret management in Keptn
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		key := strings.ToLower(pair[0])
		// if the key is prefixed with 'secret_', do not handle this environment variable
		if strings.HasPrefix(key, "secret_") {
			continue
		}
		result = strings.Replace(result, "${env."+key+"}", pair[1], -1)

		// also add the env-variable "as-is" to our list of envVariables!
		envArray = append(envArray, env)

		if VerboseLogging {
			log.Printf(env)
		}
	}

	// Second, we iterate through all data elements in our incoming event
	myMap := map[string]interface{}{}
	if err := keptnv2.Decode(incomingEvent, &myMap); err != nil {
		log.Printf("Failed to decode incomingEvent: %s", err.Error())
		return input, envArray
	}
	result, envArray = manageKeptnPlaceholdersRecursively(result, envArray, "", myMap)

	return result, envArray
}

func manageKeptnPlaceholdersRecursively(input string, envArray []string, keyPath string, values map[string]interface{}) (string, []string) {

	for key, value := range values {
		var newKeyPath string
		if keyPath != "" {
			newKeyPath = keyPath + "." + key
		} else {
			newKeyPath = key
		}
		newKeyPathPlaceHolder := "${" + newKeyPath + "}"

		// for env-variables we make them UPPER CASE and replace any . with _
		newKeyPathEnvVariable := strings.ToUpper(strings.ReplaceAll(newKeyPath, ".", "_"))

		switch value.(type) {
		case string:
			// First, replace it in the input string
			stringValue := value.(string)
			input = strings.ReplaceAll(input, newKeyPathPlaceHolder, stringValue)
			if VerboseLogging {
				log.Printf("%s --> %s", newKeyPathPlaceHolder, stringValue)
			}

			// Second, add it to the env array
			envVariableDefinition := fmt.Sprintf("%s=%s", newKeyPathEnvVariable, stringValue)
			envArray = append(envArray, envVariableDefinition)
			if VerboseLogging {
				log.Printf(envVariableDefinition)
			}
		case map[string]interface{}:
			input, envArray = manageKeptnPlaceholdersRecursively(input, envArray, newKeyPath, value.(map[string]interface{}))
		case []interface{}:
			input, envArray = manageKeptnPlaceholdersArrayRecursively(input, envArray, newKeyPath, value.([]interface{}))
		}
	}
	return input, envArray
}

func manageKeptnPlaceholdersArrayRecursively(input string, envArray []string, keyPath string, values []interface{}) (string, []string) {
	if keyPath == "" {
		return input, envArray
	}
	for index, value := range values {
		var newKeyPath string
		newKeyPath = fmt.Sprintf("%s[%d]", keyPath, index)
		newKeyPathPlaceHolder := "${" + newKeyPath + "}"

		// for env-variables we make them UPPER CASE and replace any . with _
		newKeyPathEnvVariable := strings.ToUpper(fmt.Sprintf("%s_%d", strings.ReplaceAll(keyPath, ".", "_"), index))

		switch value.(type) {
		case string:
			// First, replace it in the input string
			stringValue := value.(string)
			input = strings.ReplaceAll(input, newKeyPathPlaceHolder, stringValue)
			if VerboseLogging {
				log.Printf("%s --> %s", newKeyPathPlaceHolder, stringValue)
			}

			// Second, add it to the env array
			envVariableDefinition := fmt.Sprintf("%s=%s", newKeyPathEnvVariable, stringValue)
			envArray = append(envArray, envVariableDefinition)
			if VerboseLogging {
				log.Printf(envVariableDefinition)
			}
		case map[string]interface{}:
			input, envArray = manageKeptnPlaceholdersRecursively(input, envArray, newKeyPath, value.(map[string]interface{}))
		case []interface{}:
			input, envArray = manageKeptnPlaceholdersArrayRecursively(input, envArray, newKeyPath, value.([]interface{}))
		}
	}
	return input, envArray
}

func _nextCleanLine(lines []string, lineIx int, trim bool) (int, string) {
	// sanity check
	lineIx++
	maxLines := len(lines)
	if lineIx < 0 || maxLines <= 0 || lineIx >= maxLines {
		return -1, ""
	}

	line := ""
	for ; lineIx < maxLines; lineIx++ {
		line = lines[lineIx]
		if trim {
			line = strings.Trim(line, " ")
		}
		if strings.HasPrefix(line, "#") {
			continue
		}

		// stop if we found a new line that is not a comment!
		if len(line) >= 0 {
			break
		}
	}

	// have we reached the end of the strings?
	if lineIx >= maxLines {
		return lineIx, ""
	}
	return lineIx, line
}

//
// Parses .http raw file content and returns HTTP METHOD, URI, HEADERS, BODY
//
func parseHttpRequestFromHttpTextFile(httpfile string, incomingEvent cloudevents.Event) (genericHttpRequest, error) {
	var returnRequest genericHttpRequest

	content, err := ioutil.ReadFile(httpfile)
	if err != nil {
		return returnRequest, err
	}

	return parseHttpRequestFromString(string(content), incomingEvent)
}

//
// Parses .http string content and returns HTTP METHOD, URI, HEADERS, BODY
//
func parseHttpRequestFromString(rawContent string, incomingEvent cloudevents.Event) (genericHttpRequest, error) {
	var returnRequest genericHttpRequest

	// lets first replace all Keptn related placeholders
	rawContent, _ = manageKeptnPlaceholders(rawContent, incomingEvent)

	// lets get each line
	lines := strings.Split(rawContent, "\n")

	//
	// lets find the first clean line - must be the HTTP Method and URI, e.g: GET http://myuri
	lineIx, line := _nextCleanLine(lines, -1, true)
	if lineIx < 0 {
		return returnRequest, errors.New("No HTTP Method or URI Found")
	}

	lineSplits := strings.Split(line, " ")
	if len(lineSplits) == 1 {
		// only provides URI
		returnRequest.method = "GET"
		returnRequest.uri = lineSplits[0]
	} else {
		// provides method and URI
		returnRequest.method = lineSplits[0]
		returnRequest.uri = lineSplits[1]
	}

	//
	// now lets iterate through the next lines as they should all be headers until we end up with a blank line or EOF
	returnRequest.headers = make(map[string]string)
	lineIx, line = _nextCleanLine(lines, lineIx, true)
	for (lineIx > 0) && (len(line) > 0) {
		lineSplits = strings.Split(line, ":")
		if len(lineSplits) < 2 {
			break
		}
		headerName := strings.Trim(lineSplits[0], " ")
		headerValue := strings.Trim(lineSplits[1], " ")
		returnRequest.headers[headerName] = headerValue
		lineIx, line = _nextCleanLine(lines, lineIx, true)
	}

	//
	// if we still have content it must be the request body
	returnRequest.body = ""
	lineIx, line = _nextCleanLine(lines, lineIx, false)
	for lineIx > 0 && len(line) > 0 {
		returnRequest.body += line
		returnRequest.body += "\n"
		lineIx, line = _nextCleanLine(lines, lineIx, false)
	}

	return returnRequest, nil
}

//
// Sends a generic HTTP Request
//
func executeGenericHttpRequest(request genericHttpRequest) (int, string, error) {
	client := http.Client{}

	// define the request
	log.Println(request.method, request.uri, request.uri, request.body)
	req, err := http.NewRequest(request.method, request.uri, bytes.NewBufferString(request.body))

	if err != nil {
		return -1, "", err
	}

	// add the headers
	for key, value := range request.headers {
		req.Header.Add(key, value)
	}

	// execute
	resp, err := client.Do(req)
	if err != nil {
		return -1, "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	return resp.StatusCode, string(body), err
}

//
// Executes the commands by adding data from the incomingEvent as Env-Variables
//
func executeCommandWithKeptnContext(command string, args []string, incomingEvent cloudevents.Event, directory *string) (string, error) {
	// lets first replace all Keptn related placeholders
	_, envVars := manageKeptnPlaceholders("", incomingEvent)

	return executeCommand(command, args, envVars, directory)
}

//
// Executes a command, e.g: ls -l; ./yourscript.sh
// Also sets the enviornment variables passed
//
func executeCommand(command string, args []string, envs []string, directory *string) (string, error) {
	cmd := exec.Command(command, args...)
	if directory != nil {
		cmd.Dir = *directory
	}

	if VerboseLogging {
		log.Printf("About to execute: %s with %s", command, args)

		for _, envVariable := range envs {
			log.Printf(envVariable)
		}
	}

	// pass environment variables
	cmd.Env = envs

	// Execute Command
	out, err := cmd.CombinedOutput()
	if err != nil {
		errMessage := fmt.Sprintf("Error executing command %s %s: %s\n%s", command, strings.Join(args, " "), err.Error(), string(out))
		log.Printf(errMessage)
		err = fmt.Errorf(errMessage)

		return "", err
	} else {
		if VerboseLogging {
			log.Printf("Script executed successful")
			log.Printf("%s", string(out))
		}
	}

	return string(out), nil
}
