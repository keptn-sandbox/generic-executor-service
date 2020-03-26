package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/client"
	cloudeventshttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"

	"github.com/kelseyhightower/envconfig"

	keptnevents "github.com/keptn/go-utils/pkg/events"
	keptnutils "github.com/keptn/go-utils/pkg/utils"
)

const eventbroker = "EVENTBROKER"

var runlocal = (os.Getenv("env") == "runlocal")

type envConfig struct {
	// Port on which to listen for cloudevents
	Port int    `envconfig:"RCV_PORT" default:"8080"`
	Path string `envconfig:"RCV_PATH" default:"/"`
}

/**
 * This method gets called when a new event is received from the Keptn Event Distributor
 * Depending on the Event Type will call the specific event handler functions, e.g: handleDeploymentFinishedEvent
 */
func processKeptnCloudEvent(ctx context.Context, event cloudevents.Event) error {
	var shkeptncontext string
	event.Context.ExtensionAs("shkeptncontext", &shkeptncontext)

	nano := event.Time().UTC().UnixNano()
	milli := nano / 1000000

	// create a base Keptn Event
	keptnEvent := baseKeptnEvent{
		event:     event.Type(),
		source:    event.Source(),
		context:   shkeptncontext,
		time:      event.Time().String(),
		timeutc:   event.Time().UTC().String(),
		timeutcms: strconv.FormatInt(milli, 10),
	}

	logger := keptnutils.NewLogger(shkeptncontext, event.Context.GetID(), "jenkins-service")
	logger.Info(fmt.Sprintf("gotEvent(%s): %s - %s", event.Type(), shkeptncontext, event.Context.GetID()))

	logger.Info(fmt.Sprintf("%s -- %s -- %s", keptnEvent.time, keptnEvent.timeutc, keptnEvent.timeutcms))

	// ********************************************
	// Lets test on each possible Event Type and call the respective handler function
	// ********************************************
	if keptnEvent.event == keptnevents.ConfigurationChangeEventType {
		logger.Info("Got Configuration Change Event")

		data := &keptnevents.ConfigurationChangeEventData{}
		err := event.DataAs(data)
		if err != nil {
			logger.Error(fmt.Sprintf("Got Data Error: %s", err.Error()))
			return err
		} else {
			keptnEvent.project = data.Project
			keptnEvent.stage = data.Stage
			keptnEvent.service = data.Service
			keptnEvent.labels = data.Labels
			return handleConfigurationChangeEvent(event, keptnEvent, data, logger)
		}
		return nil
	}
	if keptnEvent.event == keptnevents.DeploymentFinishedEventType {
		logger.Info("Got Deployment Finished Event")

		data := &keptnevents.DeploymentFinishedEventData{}
		err := event.DataAs(data)
		if err != nil {
			logger.Error(fmt.Sprintf("Got Data Error: %s", err.Error()))
			return err
		} else {
			keptnEvent.project = data.Project
			keptnEvent.stage = data.Stage
			keptnEvent.service = data.Service
			keptnEvent.labels = data.Labels
			return handleDeploymentFinishedEvent(event, keptnEvent, data, logger)
		}
		return nil
	}
	if keptnEvent.event == keptnevents.TestsFinishedEventType {
		logger.Info("Got Tests Finished Change Event")

		data := &keptnevents.TestsFinishedEventData{}
		err := event.DataAs(data)
		if err != nil {
			logger.Error(fmt.Sprintf("Got Data Error: %s", err.Error()))
			return err
		} else {
			keptnEvent.project = data.Project
			keptnEvent.stage = data.Stage
			keptnEvent.service = data.Service
			keptnEvent.labels = data.Labels
			return handleTestsFinishedEvent(event, keptnEvent, data, logger)
		}
		return nil
	}
	if keptnEvent.event == keptnevents.StartEvaluationEventType {
		logger.Info("Got Start Evaluation Event")

		data := &keptnevents.StartEvaluationEventData{}
		err := event.DataAs(data)
		if err != nil {
			logger.Error(fmt.Sprintf("Got Data Error: %s", err.Error()))
			return err
		} else {
			keptnEvent.project = data.Project
			keptnEvent.stage = data.Stage
			keptnEvent.service = data.Service
			keptnEvent.labels = data.Labels
			return handleStartEvaluationEvent(event, keptnEvent, data, logger)
		}
		return nil
	}
	if keptnEvent.event == keptnevents.EvaluationDoneEventType {
		logger.Info("Got Evaluation Done Event")

		data := &keptnevents.EvaluationDoneEventData{}
		err := event.DataAs(data)
		if err != nil {
			logger.Error(fmt.Sprintf("Got Data Error: %s", err.Error()))
			return err
		} else {
			keptnEvent.project = data.Project
			keptnEvent.stage = data.Stage
			keptnEvent.service = data.Service
			keptnEvent.labels = data.Labels
			return handleEvaluationDoneEvent(event, keptnEvent, data, logger)
		}
		return nil
	}
	if keptnEvent.event == keptnevents.ProblemOpenEventType || keptnEvent.event == keptnevents.ProblemEventType {
		logger.Info("Got Problem Event")

		data := &keptnevents.ProblemEventData{}
		err := event.DataAs(data)
		if err != nil {
			logger.Error(fmt.Sprintf("Got Data Error: %s", err.Error()))
			return err
		} else {
			keptnEvent.project = data.Project
			keptnEvent.stage = data.Stage
			keptnEvent.service = data.Service
			keptnEvent.labels = data.Labels
			return handleProblemEvent(event, keptnEvent, data, logger)
		}
		return nil
	}

	// Unkonwn Keptn Event -> Throw Error!
	var errorMsg string
	errorMsg = fmt.Sprintf("Received unexpected keptn event: %s", event.Type())
	logger.Error(errorMsg)
	return errors.New(errorMsg)
}

/**
 * Usage: ./main [test] [*|testname]
 * no args: starts listening for cloudnative events on localhost:port/path
 * test:    will send test cloud events to localhost:port/path
 * test *:  will send all availalbe test events
 * test keptneventtype: will only send a test event for that keptn event type
 *
 * Environment Variables
 * env=runlocal   -> will fetch resources from local drive instead of configuration service
 */
func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Fatalf("Failed to process env var: %s", err)
	}

	if runlocal {
		log.Println("env=runlocal: Running with local filesystem to fetch resources")
		log.Println("Also setting the env variable TESTTOKEN for testing purposes")

		// set some test env variables
		os.Setenv("TESTTOKEN", "MYTESTTOKENVALUE")
	}

	os.Exit(_main(os.Args[1:], env))
}

/**
 * Opens up a listener on localhost:port/path and passes incoming requets to gotEvent
 */
func _main(args []string, env envConfig) int {

	ctx := context.Background()

	t, err := cloudeventshttp.New(
		cloudeventshttp.WithPort(env.Port),
		cloudeventshttp.WithPath(env.Path),
	)

	logger := keptnutils.NewLogger("Startup", "Init", "jenkins-service")
	logger.Info(fmt.Sprintf("Port = %d; Path=%s", env.Port, env.Path))

	if err != nil {
		log.Fatalf("failed to create transport, %v", err)
	}
	c, err := client.New(t)
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	log.Fatalf("failed to start receiver: %s", c.StartReceiver(ctx, processKeptnCloudEvent))

	return 0
}
