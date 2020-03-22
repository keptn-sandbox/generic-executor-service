package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
	"log"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/client"
	cloudeventshttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/google/uuid"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"

	configutils "github.com/keptn/go-utils/pkg/configuration-service/utils"
	keptnutils "github.com/keptn/go-utils/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	keptnevents "github.com/keptn/go-utils/pkg/events"
)

/**
 * Helper Functions
 */
 func getConfigurationServiceURL() string {
	if os.Getenv("env") == "production" {
		return "configuration-service:8080"
	}
	return "localhost:8080"
}

/**
 * Retrieves a resource (=file) from the keptn configuration repo and stores it in the local file system
 */
func getKeptnResource(project, stage, service, resource string, logger *keptnutils.Logger) (string, error) {

	// if we run in a runlocal mode we are just getting the file from the local disk
	if(runlocal) {
		return _getKeptnResourceFromLocal(project, stage, service, resource, logger);
	} 

	// get it from keptn
	resourceHandler := configutils.NewResourceHandler(getConfigurationServiceURL())
	requestedResource, err := resourceHandler.GetServiceResource(project, stage, service, resource)

	// return Nil in case resource couldnt be retrieved
	if err != nil || requestedResource.ResourceContent == "" {
		logger.Debug(fmt.Sprintf("Keptn Resource not found: %s - %s", resource, err))
		return "", err
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
		logger.Error(err.Error())
		return "", err
	}
	defer resourceFile.Close()

	_, err = resourceFile.Write([]byte(requestedResource.ResourceContent))

	if err != nil {
		logger.Error(err.Error())
		return "", err
	}

	return resource, nil
}

/**
 * Retrieves a resource (=file) from the local file system. Basically checks if the file is available and if so returns it
 */
func _getKeptnResourceFromLocal(project string, stage string, service string, resource string, logger *keptnutils.Logger) (string, error) {
	if _, err := os.Stat(resource); err == nil {
		return resource, nil  
	} else {
		return "", err
	}	
}

/**
 * Returns the Keptn Domain, e.g: keptn.yourdomain.com
 */
func getKeptnDomain() (string, error) {

	api, err := keptnutils.GetKubeAPI(true)
	if err != nil {
		return "", err
	}

	cm, err := api.ConfigMaps("keptn").Get("keptn-domain", metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	return string(cm.Data["app_domain"]), nil
}

//
// Executes a command, e.g: ls -l
//
func executeCommand(command string, args []string, logger *keptnutils.Logger) (bool, error) {
	res, err := keptnutils.ExecuteCommand(command, args)
	
	logger.Info(res)
	if err != nil {
		logger.Error(err.Error())
		return false, err
	}

	logger.Debug("Successfull executed command: " + command);

	return true, nil
}

//
// Sends a ConfigurationChangeEventType = "sh.keptn.event.configuration.change"
//
func sendConfigurationChangeEvent(shkeptncontext string, incomingEvent *cloudevents.Event, project, service, stage string, labels map[string]string, logger *keptnutils.Logger) error {
	source, _ := url.Parse("jenkins-service")
	contentType := "application/json"

	configurationChangeData := keptnevents.ConfigurationChangeEventData{}
	
	// if we have an incoming event we pre-populate data
	if incomingEvent != nil {
		incomingEvent.DataAs(&configurationChangeData)
	}

	if project != "" 	{ configurationChangeData.Project = project }
	if service != "" 	{ configurationChangeData.Service = service }
	if stage != "" 		{ configurationChangeData.Stage = stage }
	if labels != nil 	{ configurationChangeData.Labels = labels }

	event := cloudevents.Event{
		Context: cloudevents.EventContextV02{
			ID:          uuid.New().String(),
			Time:        &types.Timestamp{Time: time.Now()},
			Type:        keptnevents.ConfigurationChangeEventType,
			Source:      types.URLRef{URL: *source},
			ContentType: &contentType,
			Extensions:  map[string]interface{}{"shkeptncontext": shkeptncontext},
		}.AsV02(),
		Data: configurationChangeData,
	}

	if (logger != nil) {
		logger.Info(fmt.Sprintf("%s", event))
	} else {
		log.Println(fmt.Sprintf("%s", event));
	}

	return sendCloudNativeEvent(event)
}

//
// Sends a DeploymentFinishedEventType = "sh.keptn.events.deployment-finished"
//
 func sendDeploymentFinishedEvent(shkeptncontext string, incomingEvent *cloudevents.Event, project, service, stage, teststrategy, deploymentstrategy, image, tag, deploymentURILocal, deploymentURIPublic string, labels map[string]string, logger *keptnutils.Logger) error {
	source, _ := url.Parse("jenkins-service")
	contentType := "application/json"

	deploymentFinishedData := keptnevents.DeploymentFinishedEventData{}
	
	// if we have an incoming event we pre-populate data
	if incomingEvent != nil {
		incomingEvent.DataAs(&deploymentFinishedData)
	}

	if project != "" 			{ deploymentFinishedData.Project = project }
	if service != "" 			{ deploymentFinishedData.Service = service	}
	if stage != ""   			{ deploymentFinishedData.Stage = stage	}
	if teststrategy != "" 		{ deploymentFinishedData.TestStrategy = teststrategy }
	if deploymentstrategy != "" { deploymentFinishedData.DeploymentStrategy = deploymentstrategy }
	if image != ""   			{ deploymentFinishedData.Image = image }
	if tag != ""   				{ deploymentFinishedData.Tag = tag }
	
	if labels != nil 			{ deploymentFinishedData.Labels = labels }

	event := cloudevents.Event{
		Context: cloudevents.EventContextV02{
			ID:          uuid.New().String(),
			Time:        &types.Timestamp{Time: time.Now()},
			Type:        keptnevents.DeploymentFinishedEventType,
			Source:      types.URLRef{URL: *source},
			ContentType: &contentType,
			Extensions:  map[string]interface{}{"shkeptncontext": shkeptncontext},
		}.AsV02(),
		Data: deploymentFinishedData,
	}

	if (logger != nil) {
		logger.Info(fmt.Sprintf("%s", event))
	} else {
		log.Println(fmt.Sprintf("%s", event));
	}

	return sendCloudNativeEvent(event)

}

//
// Sends a TestsFinishedEventType = "sh.keptn.events.tests-finished"
//
func sendTestsFinishedEvent(shkeptncontext string, incomingEvent *cloudevents.Event, project, service, stage, teststrategy, deploymentstrategy string, startedAt, finishedAt time.Time, result string, labels map[string]string, logger *keptnutils.Logger) error {
	source, _ := url.Parse("jenkins-service")
	contentType := "application/json"

	testFinishedData := keptnevents.TestsFinishedEventData{}

	// if we have an incoming event we pre-populate data
	if incomingEvent != nil {
		incomingEvent.DataAs(&testFinishedData)
	}

	if project != "" 			{ testFinishedData.Project = project }
	if service != "" 			{ testFinishedData.Service = service	}
	if stage != ""   			{ testFinishedData.Stage = stage	}
	if teststrategy != "" 		{ testFinishedData.TestStrategy = teststrategy }
	if deploymentstrategy != "" { testFinishedData.DeploymentStrategy = deploymentstrategy }
	
	if labels != nil 			{ testFinishedData.Labels = labels }

	// fill in timestamps
	testFinishedData.Start = startedAt.Format(time.RFC3339)
	testFinishedData.End = time.Now().Format(time.RFC3339)

	// set test result
	testFinishedData.Result = result

	event := cloudevents.Event{
		Context: cloudevents.EventContextV02{
			ID:          uuid.New().String(),
			Time:        &types.Timestamp{Time: time.Now()},
			Type:        keptnevents.TestsFinishedEventType,
			Source:      types.URLRef{URL: *source},
			ContentType: &contentType,
			Extensions:  map[string]interface{}{"shkeptncontext": shkeptncontext},
		}.AsV02(),
		Data: testFinishedData,
	}

	if (logger != nil) {
		logger.Info(fmt.Sprintf("%s", event))
	} else {
		log.Println(fmt.Sprintf("%s", event));
	}

	return sendCloudNativeEvent(event)
}

//
// Sends a Cloud Native event to the endpoint configured in the env-variable EVENTBROKER, e.g: http://event-broker.keptn.svc.cluster.local/keptn
//
func sendCloudNativeEvent(event cloudevents.Event) error {
	endPoint, err := getServiceEndpoint(eventbroker)
	if err != nil {
		return errors.New("Failed to retrieve endpoint of eventbroker. %s" + err.Error())
	}

	if endPoint.Host == "" {
		return errors.New("Host of eventbroker not set")
	}

	transport, err := cloudeventshttp.New(
		cloudeventshttp.WithTarget(endPoint.String()),
		cloudeventshttp.WithEncoding(cloudeventshttp.StructuredV02),
	)
	if err != nil {
		return errors.New("Failed to create transport:" + err.Error())
	}

	c, err := client.New(transport)
	if err != nil {
		return errors.New("Failed to create HTTP client:" + err.Error())
	}

	if _, err := c.Send(context.Background(), event); err != nil {
		return errors.New("Failed to send cloudevent:, " + err.Error())
	}
	return nil
}

//
// getServiceEndpoint gets an endpoint stored in an environment variable and sets http as default scheme
//
func getServiceEndpoint(service string) (url.URL, error) {
	url, err := url.Parse(os.Getenv(service))
	if err != nil {
		return *url, fmt.Errorf("Failed to retrieve value from ENVIRONMENT_VARIABLE: %s", service)
	}

	if url.Scheme == "" {
		url.Scheme = "http"
	}

	return *url, nil
}