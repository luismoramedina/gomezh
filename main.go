package main

import (
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	"github.com/opentracing/opentracing-go"
	"net/http"
	"log"
	"github.com/luismoramedina/gomesh/egress"
	"github.com/luismoramedina/gomesh/sidecar"
	"github.com/luismoramedina/gomesh/ingress"
	"time"
)


func main() {
	var tracer opentracing.Tracer
	authMap := sidecar.NewAuthsMap()
	timeMap := sidecar.NewTimeMap()

	// 1) Create a opentracing.Tracer that does nothing, use a
	collector := new(zipkin.NopCollector)
	tracer, _ = zipkin.NewTracer(
		zipkin.NewRecorder(collector, false, "127.0.0.1:0", "mesh"))

	egressController := egress.EgressController{
		Sidecar: sidecar.Sidecar{
			Tracer: tracer, Auths: authMap, Times: timeMap}}
	ingressController := ingress.IngressController{
		Sidecar: sidecar.Sidecar{
			Tracer: tracer, Auths: authMap, Times: timeMap}}

	ingressHandler := http.HandlerFunc(ingressController.Handler)
	egressHandler := http.HandlerFunc(egressController.Handler)
	log.Printf("Listening ingress %s, egress %s", ":8080" ,":8082")

	ingressServer := &http.Server{
		Addr: ":8080",
		Handler: ingress.SecurityMiddleware(ingressHandler),
		ReadTimeout: 5 * time.Second,
		WriteTimeout: 5 * time.Second}

	egressServer := &http.Server{
		Addr: ":8082",
		Handler: egressHandler,
		ReadTimeout: 5 * time.Second,
		WriteTimeout: 5 * time.Second}

	go egressServer.ListenAndServe()
	ingressServer.ListenAndServe()
}