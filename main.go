package main

import (
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	"github.com/opentracing/opentracing-go"
	"net/http"
	"log"
	"github.com/luismoramedina/gomesh/egress"
	"github.com/luismoramedina/gomesh/sidecar"
	"github.com/luismoramedina/gomesh/ingress"
)

var tracer opentracing.Tracer
var auths map[uint64]string

type IngressController struct {
	sidecar.Sidecar
}

func main() {
	auths = make(map[uint64]string)

	// 1) Create a opentracing.Tracer that does nothing, use a
	collector := new(zipkin.NopCollector)
	tracer, _ = zipkin.NewTracer(
		zipkin.NewRecorder(collector, false, "127.0.0.1:0", "mesh"))

	egressController := egress.EgressController{Sidecar: sidecar.Sidecar{Tracer: tracer, Auths: auths}}
	ingressController := ingress.IngressController{Sidecar: sidecar.Sidecar{Tracer: tracer, Auths: auths}}

	ingressHandler := http.HandlerFunc(ingressController.Handler)
	egressHandler := http.HandlerFunc(egressController.Handler)
	log.Printf("Listening ingress %s, egress %s", ":8080" ,":8082")
	go http.ListenAndServe(":8082", egressHandler)
	http.ListenAndServe(":8080", ingress.SecurityMiddleware(ingressHandler))
}