package main

import (
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

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

func printRouterInfoWithTrace(routerName string, tracer trace.Tracer) http.HandlerFunc {
	out := "Router name: " + routerName + " - " + appName + " " + appVersion
	return func(w http.ResponseWriter, r *http.Request) {
		_, span := tracer.Start(r.Context(), routerName)
		defer span.End()
		w.Header().Add("X-Router-Name", routerName)
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(out))
		if err != nil {
			log.Print(err)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
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
		stateProcessed.WithLabelValues(ledId, ledAction, strconv.Itoa(res.StatusCode)).Inc()
		_, err = w.Write([]byte(resBody))
		if err != nil {
			log.Print(err)
		}
	}
}

func ledStateUpdateWithTrace(lc LedClient, tracer trace.Tracer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "ledStateUpdate")
		span.SetName("ledStateUpdateWithTrace")
		defer span.End()

		ledId := chi.URLParam(r, "ledId")
		ledAction := chi.URLParam(r, "ledAction")
		span.SetAttributes(attribute.String("ledId", ledId), attribute.String("ledAction", ledAction))

		if errId := lc.Id(ledId); errId != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			if _, err := w.Write([]byte(errId.Error())); err != nil {
				log.Print(err)
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
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
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
			return
		}

		// res, err := lc.Do()
		res, err := lc.DoWithTrace(ctx, tracer)
		if err != nil {
			log.Print(err)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		defer res.Body.Close()

		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			log.Print(err)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		w.Header().Add("Content-Type", res.Header.Get("Content-Type"))
		w.WriteHeader(res.StatusCode)
		stateProcessed.WithLabelValues(ledId, ledAction, strconv.Itoa(res.StatusCode)).Inc()
		_, err = w.Write([]byte(resBody))
		if err != nil {
			log.Print(err)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
	}
}
