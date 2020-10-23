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
	"strings"

	api "github.com/keptn/go-utils/pkg/api/utils"
	keptn "github.com/keptn/go-utils/pkg/lib"
)

/**
 * Structs
 */

type BaseKeptnEvent struct {
	// context, source and eventid
	context   string
	source    string
	event     string
	time      string
	timeutc   string
	timeutcms string

	// project & deployment specific
	project      string
	stage        string
	service      string
	deployment   string
	testStrategy string

	// test specific
	testStart string
	testEnd   string

	// evaluation specific
	evaluationStart  string
	evaluationEnd    string
	evaluationResult string

	// only filled for deployment events
	tag                 string
	image               string
	deploymentURILocal  string
	deploymentURIPublic string

	// action & problem specific
	action       string
	problemState string
	problemID    string
	problemTitle string
	pid          string
	problemURL   string

	labels map[string]string
}

type genericHttpRequest struct {
	method  string
	uri     string
	headers map[string]string
	body    string
}

var KeptnOptions = keptn.KeptnOpts{}

/**
 * Retrieves a resource (=file) from the keptn configuration repo and stores it in the local file system
 */
func getKeptnResource(myKeptn *keptn.Keptn, resource string) (string, error) {

	// if we run in a runlocal mode we are just getting the file from the local disk
	if KeptnOptions.UseLocalFileSystem {
		return _getKeptnResourceFromLocal(resource)
	}

	resourceHandler := api.NewResourceHandler(KeptnOptions.ConfigurationServiceURL)

	// SERVICE-LEVEL: lets try to find it on service level
	requestedResource, err := resourceHandler.GetServiceResource(myKeptn.KeptnBase.Project, myKeptn.KeptnBase.Stage, myKeptn.KeptnBase.Service, resource)
	if err != nil || requestedResource.ResourceContent == "" {
		// STAGE-LEVEL: not found on service level - lets search one level up on stage level
		requestedResource, err = resourceHandler.GetStageResource(myKeptn.KeptnBase.Project, myKeptn.KeptnBase.Stage, resource)
		if err != nil || requestedResource.ResourceContent == "" {
			// PROJECT-LEVEL: not found on the stage level - lets search one level up on project level
			requestedResource, err = resourceHandler.GetProjectResource(myKeptn.KeptnBase.Project, resource)

			if err != nil || requestedResource.ResourceContent == "" {
				myKeptn.Logger.Debug(fmt.Sprintf("Keptn Resource not found: %s/%s/%s/%s - %s", myKeptn.KeptnBase.Project, myKeptn.KeptnBase.Stage, myKeptn.KeptnBase.Service, resource, err))
				return "", err
			}

			myKeptn.Logger.Debug("Found " + resource + " on project level")
		} else {
			myKeptn.Logger.Debug("Found " + resource + " on stage level")
		}
	} else {
		myKeptn.Logger.Debug("Found " + resource + " on service level")
	}

	// now store that file on the same directory structure locally
	os.RemoveAll(resource)
	pathArr := strings.Split(resource, "/")
	directory := ""
	for _, pathItem := range pathArr[0 : len(pathArr)-1] {
		directory += pathItem + "/"
	}

	err = os.MkdirAll(directory, os.ModePerm)
	if err != nil {
		return "", err
	}
	resourceFile, err := os.Create(resource)
	if err != nil {
		myKeptn.Logger.Error(err.Error())
		return "", err
	}
	defer resourceFile.Close()

	_, err = resourceFile.Write([]byte(requestedResource.ResourceContent))

	if err != nil {
		myKeptn.Logger.Error(err.Error())
		return "", err
	}

	// if the downloaded file is a shell script we also change the permissions
	if strings.HasSuffix(resource, ".sh") {
		os.Chmod(resource, 0777)
	}

	return resource, nil
}

/**
 * Retrieves a resource (=file) from the local file system. Basically checks if the file is available and if so returns it
 */
func _getKeptnResourceFromLocal(resource string) (string, error) {
	if _, err := os.Stat(resource); err == nil {
		return resource, nil
	} else {
		return "", err
	}
}

//
// replaces $ placeholders with actual values
// $TIMESTRING, $TIMEUTCSTRING, $TIMEUTCMS
// $CONTEXT, $EVENT, $SOURCE
// $PROJECT, $STAGE, $SERVICE
// $DEPLOYMENT, $TESTSTRATEGY
// $DEPLOYMENTURILOCAL, $DEPLOYMENTURIPUBLIC
// $LABEL.XXXX  -> will replace that with a label called XXXX
// $ENV.XXXX    -> will replace that with an env variable called XXXX
// $SECRET.YYYY -> will replace that with the k8s secret called YYYY
//
func replaceKeptnPlaceholders(input string, keptnEvent BaseKeptnEvent) string {
	result := input

	// first we do the regular keptn values
	result = strings.Replace(result, "$TIMESTRING", keptnEvent.time, -1)
	result = strings.Replace(result, "$TIMEUTCSTRING", keptnEvent.timeutc, -1)
	result = strings.Replace(result, "$TIMEUTCMS", keptnEvent.timeutcms, -1)

	result = strings.Replace(result, "$CONTEXT", keptnEvent.context, -1)
	result = strings.Replace(result, "$EVENT", keptnEvent.event, -1)
	result = strings.Replace(result, "$SOURCE", keptnEvent.source, -1)

	result = strings.Replace(result, "$PROJECT", keptnEvent.project, -1)
	result = strings.Replace(result, "$STAGE", keptnEvent.stage, -1)
	result = strings.Replace(result, "$SERVICE", keptnEvent.service, -1)
	result = strings.Replace(result, "$DEPLOYMENT", keptnEvent.deployment, -1)
	result = strings.Replace(result, "$TESTSTRATEGY", keptnEvent.testStrategy, -1)

	result = strings.Replace(result, "$DEPLOYMENTURILOCAL", keptnEvent.deploymentURILocal, -1)
	result = strings.Replace(result, "$DEPLOYMENTURIPUBLIC", keptnEvent.deploymentURIPublic, -1)

	result = strings.Replace(result, "$ACTION", keptnEvent.action, -1)

	result = strings.Replace(result, "$PROBLEMID", keptnEvent.problemID, -1)
	result = strings.Replace(result, "$PROBLEMSTATE", keptnEvent.problemState, -1)
	result = strings.Replace(result, "$PID", keptnEvent.pid, -1)
	result = strings.Replace(result, "$PROBLEMTITLE", keptnEvent.problemTitle, -1)
	result = strings.Replace(result, "$PROBLEMURL", keptnEvent.problemURL, -1)

	// now we do the labels
	for key, value := range keptnEvent.labels {
		result = strings.Replace(result, "$LABEL_"+key, value, -1)
	}

	// now we do all environment variables
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		result = strings.Replace(result, "$ENV_"+pair[0], pair[1], -1)
	}

	// TODO: iterate through k8s secrets!

	return result
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
func parseHttpRequestFromHttpTextFile(keptnEvent BaseKeptnEvent, httpfile string) (genericHttpRequest, error) {
	var returnRequest genericHttpRequest

	content, err := ioutil.ReadFile(httpfile)
	if err != nil {
		return returnRequest, err
	}

	return parseHttpRequestFromString(string(content), keptnEvent)
}

//
// Parses .http string content and returns HTTP METHOD, URI, HEADERS, BODY
//
func parseHttpRequestFromString(rawContent string, keptnEvent BaseKeptnEvent) (genericHttpRequest, error) {
	var returnRequest genericHttpRequest

	// lets first replace all Keptn related placeholders
	rawContent = replaceKeptnPlaceholders(rawContent, keptnEvent)

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

func executeCommandWithKeptnContext(command string, args []string, keptnEvent BaseKeptnEvent, directory *string) (string, error) {
	// build the map of environment variables

	// first we build our core keptn values
	keptnEnvs := []string{
		"TIMESTRING=" + keptnEvent.time,
		"TIMEUTCSTRING=" + keptnEvent.timeutc,
		"TIMEUTCMS=" + keptnEvent.timeutcms,
		"CONTEXT=" + keptnEvent.context,
		"EVENT=" + keptnEvent.event,
		"SOURCE=" + keptnEvent.source,
		"PROJECT=" + keptnEvent.project,
		"SERVICE=" + keptnEvent.service,
		"STAGE=" + keptnEvent.stage,
		"DEPLOYMENT=" + keptnEvent.deployment,
		"TESTSTRATEGY=" + keptnEvent.testStrategy,
		"DEPLOYMENTURILOCAL=" + keptnEvent.deploymentURILocal,
		"DEPLOYMENTURIPUBLIC=" + keptnEvent.deploymentURIPublic,
		"ACTION=" + keptnEvent.action,
		"PROBLEMID=" + keptnEvent.problemID,
		"PROBLEMSTATE=" + keptnEvent.problemState,
		"PID=" + keptnEvent.pid,
		"PROBLEMTITLE=" + keptnEvent.problemTitle,
		"PROBLEMURL=" + keptnEvent.problemURL,
	}

	// we combine the environment variables of our running process with all those with labels
	// those from our local process are prefixed with ENV_ , e.g: ENV_processenv=abcd
	// those coming from labels are prefixed with LABEL_, e.g: LABEL_mylabel=abcd
	localEnvs := os.Environ()
	commandEnvs := make([]string, len(keptnEnvs)+len(localEnvs)+len(keptnEvent.labels))
	var envIx = 0
	for _, env := range keptnEnvs {
		commandEnvs[envIx] = env
		envIx++
	}
	for _, env := range localEnvs {
		commandEnvs[envIx] = "ENV_" + env
		envIx++
	}
	for key, value := range keptnEvent.labels {
		commandEnvs[envIx] = "LABEL_" + key + "=" + value
		envIx++
	}

	return executeCommand(command, args, commandEnvs, directory)
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

	// pass environment variables
	cmd.Env = envs

	// Execute Command
	out, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("Error executing command %s %s: %s\n%s", command, strings.Join(args, " "), err.Error(), string(out))
		return "", err
	}

	return string(out), nil
}
