package main

import (
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	"io/ioutil"
	"bytes"
	"github.com/opentracing/opentracing-go"
	"fmt"
	"net/http"
	"log"
	"github.com/luismoramedina/gomesh/egress"
)

var tracer opentracing.Tracer
var auths map[uint64]string

func main() {
	auths = make(map[uint64]string)

	// 1) Create a opentracing.Tracer that does nothing, use a 
	collector := new(zipkin.NopCollector)
	tracer, _ = zipkin.NewTracer(
		zipkin.NewRecorder(collector, false, "127.0.0.1:0", "mesh"))

	egress.Tracer = tracer
	egress.Auths = auths
	ingressHandler := http.HandlerFunc(ingressHandler)
	egressHandler := http.HandlerFunc(egress.EgressHandler)
	log.Printf("Listening ingress %s, egress %s", ":8080" ,":8082")
	go http.ListenAndServe(":8082", egressHandler)
	http.ListenAndServe(":8080", securityMiddleware(ingressHandler))
}

func ingressHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting ingress request")
	span := tracer.StartSpan("request")
	defer span.Finish()
	reqId := span.Context().(zipkin.SpanContext).TraceID.Low
	log.Println(reqId)

	auths[reqId] = r.Header.Get("Authorization")

	log.Printf("auths -> %+v\n", auths)

	resp, err := forwardRequest(w, r, span)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	defer resp.Body.Close()

	for header, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}

	w.Write(resBody)

}
func securityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Executing security middleware")
		if ((len(r.Header.Get("Authorization"))) == 0) {
			http.Error(w, "no authorization header found", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

//https://stackoverflow.com/questions/34724160/go-http-send-incoming-http-request-to-an-other-server-using-client-do
func forwardRequest(w http.ResponseWriter, req *http.Request, span opentracing.Span) (*http.Response, error) {
	// we need to buffer the body if we want to read it here and send it
	// in the request.
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}

	// create a new url from the raw RequestURI sent by the client
	url := fmt.Sprintf("%s://%s%s", "http", "localhost:8081", req.RequestURI)

	proxyReq, err := http.NewRequest(req.Method, url, bytes.NewReader(body))

	// We may want to filter some headers, otherwise we could just use a shallow copy
	proxyReq.Header = req.Header

	log.Printf("spancontext-> %+v\n", span.Context())

	// Transmit the span's TraceContext as HTTP headers on our
	// outbound request.
	tracer.Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(proxyReq.Header))

	httpClient := http.Client{}
	resp, err := httpClient.Do(proxyReq)
	if err != nil {
		return nil, err
	}

	return resp, nil
}