package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	chiprometheus "github.com/toshi0607/chi-prometheus"

	"github.com/riandyrn/otelchi"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

var (
	appName     string = "ledlighter"
	appAddr     string = ":3080"
	metricsAddr string = ":3088"
	appVersion  string = "development"
	buildDate   string

	stateProcessed = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ledlighter_processed_state_actions",
		Help: "Led state update actions",
	}, []string{"ledId", "ledAction", "statusCode"})
)

func prometheusInit() {
	prometheus.MustRegister(stateProcessed)
}

func tracerInit(cfg *Config) trace.Tracer {
	tracer := noop.NewTracerProvider().Tracer("noop")
	if cfg.EnableTraces {
		tracer, err := NewTracer(appName)
		if err != nil {
			log.Fatalf("unable to initialize tracer due: %v", err)
		}
		return tracer
	}
	return tracer
}

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	ledsMap, err := GetLedsMap(cfg)
	if err != nil {
		log.Fatal(err)
	}

	lc := LedClient{Leds: ledsMap}

	tracer := tracerInit(cfg)

	prometheusInit()
	metricsRouter := chi.NewRouter()

	metrics := chiprometheus.New("metricsRouter")
	metrics.MustRegisterDefault()
	metricsRouter.Use(metrics.Handler)
	metricsRouter.Handle("/metrics", promhttp.Handler())

	mainRouter := chi.NewRouter()

	mainRouter.Use(middleware.Heartbeat("/healthz"))
	mainRouter.Use(middleware.Heartbeat("/livez"))
	mainRouter.Use(middleware.Heartbeat("/readyz"))

	mainRouter.Use(middleware.Logger)
	mainRouter.Use(middleware.RequestID)
	mainRouter.Use(middleware.Recoverer)
	mainRouter.Use(otelchi.Middleware(appName, otelchi.WithChiRoutes(mainRouter)))

	mainRouter.Get("/", printRouterInfoWithTrace("Main", tracer))

	apiRouter := chi.NewRouter()
	apiRouter.Get("/", printRouterInfoWithTrace("Api", tracer))

	ledRouter := chi.NewRouter()
	ledRouter.Get("/", printRouterInfoWithTrace("Led", tracer))
	ledRouter.Get("/{ledId}/state/{ledAction}", ledStateUpdateWithTrace(lc, tracer))

	apiRouter.Mount("/led", ledRouter)
	mainRouter.Mount("/api/v1", apiRouter)

	go func() {
		log.Printf("starting metrics for %v on address %v", appName, metricsAddr)
		err = http.ListenAndServe(metricsAddr, metricsRouter)
		if err != nil {
			panic(err)
		}
	}()

	log.Printf("starting app %v (version %v, build date %v) on address %v", appName, appVersion, buildDate, appAddr)
	err = http.ListenAndServe(appAddr, mainRouter)
	if err != nil {
		panic(err)
	}
}
