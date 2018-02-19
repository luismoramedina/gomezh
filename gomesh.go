package main

import (
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	"io/ioutil"
	"bytes"
	"github.com/opentracing/opentracing-go"
	"fmt"
	"net/http"
)

var tracer opentracing.Tracer

func main() {
	// 1) Create a opentracing.Tracer that does nothing, use a 
	collector := new(zipkin.NopCollector)
	tracer, _ = zipkin.NewTracer(
		zipkin.NewRecorder(collector, false, "127.0.0.1:0", "mesh"))

	http.HandleFunc("/", handleRequest)
	fmt.Println("Listening")
	//   http.ListenAndServe(":8080", nethttp.Middleware(tracer, http.DefaultServeMux))
	http.ListenAndServe(":8080", nil)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	span := tracer.StartSpan("request")
	defer span.Finish()
	fmt.Println(span.Context().(zipkin.SpanContext).TraceID.Low)

	resp, _ := forwardRequest(w, r, span)

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

//https://stackoverflow.com/questions/34724160/go-http-send-incoming-http-request-to-an-other-server-using-client-do
func forwardRequest(w http.ResponseWriter, req *http.Request, span opentracing.Span) (*http.Response, error) {
	// we need to buffer the body if we want to read it here and send it
	// in the request.
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}

	// you can reassign the body if you need to parse it as multipart
	req.Body = ioutil.NopCloser(bytes.NewReader(body))

	// create a new url from the raw RequestURI sent by the client
	url := fmt.Sprintf("%s://%s%s", "http", "localhost:8081", req.RequestURI)

	proxyReq, err := http.NewRequest(req.Method, url, bytes.NewReader(body))

	// We may want to filter some headers, otherwise we could just use a shallow copy
	// proxyReq.Header = req.Header
	proxyReq.Header = make(http.Header)
	for h, val := range req.Header {
		proxyReq.Header[h] = val
	}

	fmt.Printf("spancontext-> %+v\n", span.Context())

	// Transmit the span's TraceContext as HTTP headers on our
	// outbound request.
	tracer.Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(proxyReq.Header))

	httpClient := http.Client{}
	resp, err := httpClient.Do(proxyReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return nil, err
	}

	return resp, nil
}