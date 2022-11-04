package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/cloudevents/sdk-go/observability/opencensus/v2/client"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/protocol"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"

	"go.uber.org/zap"
	"knative.dev/pkg/tracing"
	"knative.dev/pkg/tracing/config"
)

// Events holds the historical events
type Events struct {
	mu     sync.Mutex
	Events []cloudevents.Event `json:"cloudEvents"`
}

var logger *zap.Logger
var events Events = Events{
	Events: make([]event.Event, 0),
}

const eventsLimit = 100
const serviceName = "ce-event-passthrough"

// AddEvent adds an event to the list of events
func (e *Events) AddEvent(event cloudevents.Event) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(e.Events) > eventsLimit {
		e.Events = append(e.Events[1:], event)
	} else {
		e.Events = append(e.Events, event)
	}
}

// display prints the given Event in a human-readable format.
func display(event cloudevents.Event) (*event.Event, protocol.Result) {
	logger.Debug("evt_rcv", zap.Any("ce_evt", event))
	fmt.Printf("\n🚀  received cloudevents.Event\n%s", event)
	go events.AddEvent(event)
	return rebuildEvent(event)
}

func rebuildEvent(event cloudevents.Event) (*event.Event, error) {
	var v interface{}
	clone := event.Clone()
	err := json.Unmarshal(event.Data(), &v)
	err = clone.SetData(cloudevents.ApplicationJSON, v)
	return &clone, err
}

func run(ctx context.Context) {
	c, err := client.NewClientHTTP(
		[]cehttp.Option{cehttp.WithMiddleware(ceMiddleware)}, nil,
	)
	if err != nil {
		logger.Fatal("ce_client_err", zap.Error(err))
	}
	conf, err := config.JSONToTracingConfig(os.Getenv("K_CONFIG_TRACING"))
	if err != nil {
		logger.Info("failed_to_create_tracing_cfg", zap.Error(err))
	}
	tracer, err := tracing.SetupPublishingWithStaticConfig(logger.Sugar(), serviceName, conf)
	if err != nil {
		logger.Fatal("tracing_err", zap.Error(err))
	}
	defer tracer.Shutdown(context.Background())

	if err := c.StartReceiver(ctx, display); err != nil {
		logger.Fatal("ce_rcv_err", zap.Error(err))
	}
}

// HTTP path of the health endpoint used for probing the service.
const healthzPath = "/healthz"

// HTTP path of the last event endpoint used for retrieving the newest received event.
const lastEventPath = "/eventz"

func healthz(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func eventz(w http.ResponseWriter) {
	b, err := json.Marshal(events.Events)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

// ceMiddleware is a cehttp.Middleware which exposes health/event history endpoints.
func ceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.RequestURI {
		case healthzPath:
			healthz(w)
			break
		case lastEventPath:
			eventz(w)
			break
		default:
			next.ServeHTTP(w, req)
			break
		}
	})
}

func main() {
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		panic(err)
	}
	run(context.Background())
}
