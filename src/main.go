package main

import (
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var (
	appName    string = "ledlighter"
	appAddr    string = ":3080"
	appVersion string = "development"
	buildDate  string
)

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

	mainRouter := chi.NewRouter()

	mainRouter.Use(middleware.Heartbeat("/healthz"))
	mainRouter.Use(middleware.Heartbeat("/livez"))
	mainRouter.Use(middleware.Heartbeat("/readyz"))

	mainRouter.Use(middleware.Logger)
	mainRouter.Use(middleware.RequestID)
	mainRouter.Use(middleware.Recoverer)

	mainRouter.Get("/", printRouterInfo("Main"))

	apiRouter := chi.NewRouter()
	apiRouter.Get("/", printRouterInfo("Api"))

	ledRouter := chi.NewRouter()
	ledRouter.Get("/", printRouterInfo("Led"))
	ledRouter.Get("/{ledId}/state/{ledAction}", ledStateUpdate(lc))

	apiRouter.Mount("/led", ledRouter)
	mainRouter.Mount("/api/v1", apiRouter)

	log.Printf("starting app %v (version %v, build date %v) on address %v", appName, appVersion, buildDate, appAddr)
	err = http.ListenAndServe(appAddr, mainRouter)
	if err != nil {
		panic(err)
	}
}

func printRouterInfo(routerName string) http.HandlerFunc {
	out := "Router name: " + routerName + " - " + appName + " " + appVersion
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-Router-Name", routerName)
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(out))
		if err != nil {
			log.Print(err)
		}
	}
}

func ledStateUpdate(lc LedClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ledId := chi.URLParam(r, "ledId")
		ledAction := chi.URLParam(r, "ledAction")

		if errId := lc.Id(ledId); errId != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			if _, err := w.Write([]byte(errId.Error())); err != nil {
				log.Print(err)
			}
			return
		}

		switch ledAction {

		case "on":
			lc.On()

		case "off":
			lc.Off()

		case "toggle":
			lc.Toggle()

		default:
			w.WriteHeader(http.StatusUnprocessableEntity)
			if _, err := w.Write([]byte("unknown led action")); err != nil {
				log.Print(err)
			}
			return
		}

		res, err := lc.Do()
		if err != nil {
			log.Print(err)
		}
		defer res.Body.Close()

		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			log.Print(err)
		}

		w.Header().Add("Content-Type", res.Header.Get("Content-Type"))
		w.WriteHeader(res.StatusCode)
		_, err = w.Write([]byte(resBody))
		if err != nil {
			log.Print(err)
		}
	}
}
