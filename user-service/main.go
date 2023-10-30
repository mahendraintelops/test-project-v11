package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/mahendraintelops/test-project-v11/user-service/config"
	restcontrollers "github.com/mahendraintelops/test-project-v11/user-service/pkg/rest/server/controllers"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sinhashubham95/go-actuator"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"os"
)

var (
	serviceName  = os.Getenv("SERVICE_NAME")
	collectorURL = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	insecure     = os.Getenv("INSECURE_MODE")
)

func main() {

	// rest server configuration
	router := gin.Default()
	var restTraceProvider *sdktrace.TracerProvider
	if len(serviceName) > 0 && len(collectorURL) > 0 {
		// add opentel
		restTraceProvider = config.InitRestTracer(serviceName, collectorURL, insecure)
		router.Use(otelgin.Middleware(serviceName))
	}
	defer func() {
		if restTraceProvider != nil {
			if err := restTraceProvider.Shutdown(context.Background()); err != nil {
				log.Printf("Error shutting down tracer provider: %v", err)
			}
		}
	}()
	// add actuator
	addActuator(router)
	// add prometheus
	addPrometheus(router)

	userController, err := restcontrollers.NewUserController()
	if err != nil {
		log.Errorf("error occurred: %v", err)
		os.Exit(1)
	}

	v1 := router.Group("/v1")
	{

		v1.POST("/users", userController.CreateUser)

		v1.GET("/users", userController.ListUsers)

		v1.GET("/users/:id", userController.FetchUser)

	}

	Port := ":3444"
	log.Println("Server started")
	if err = router.Run(Port); err != nil {
		log.Errorf("error occurred: %v", err)
		os.Exit(1)
	}

}

func prometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func addPrometheus(router *gin.Engine) {
	router.GET("/metrics", prometheusHandler())
}

func addActuator(router *gin.Engine) {
	actuatorHandler := actuator.GetActuatorHandler(&actuator.Config{Endpoints: []int{
		actuator.Env,
		actuator.Info,
		actuator.Metrics,
		actuator.Ping,
		// actuator.Shutdown,
		actuator.ThreadDump,
	},
		Env:     "dev",
		Name:    "user-service",
		Port:    3444,
		Version: "0.0.1",
	})
	ginActuatorHandler := func(ctx *gin.Context) {
		actuatorHandler(ctx.Writer, ctx.Request)
	}
	router.GET("/actuator/*endpoint", ginActuatorHandler)
}

func init() {
	// Log as JSON instead of the default ASCII formatter.
	// log.SetFormatter(&log.JSONFormatter{})
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)
	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)
}
