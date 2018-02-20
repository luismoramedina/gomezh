package egress

import (
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	"net/http"
	"github.com/opentracing/opentracing-go"
	"log"
	"fmt"
	"strings"
	"io/ioutil"
	"bytes"
)

var Tracer opentracing.Tracer
var Auths map[uint64]string

func EgressHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting Egress request")
	wireContext, err := Tracer.Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(r.Header))
	if err != nil {
		http.Error(w, "", http.StatusServiceUnavailable)
		return
	}
	traceId := wireContext.(zipkin.SpanContext).TraceID.Low
	log.Println(traceId)

	authorization := Auths[traceId]
	log.Println(authorization)

	r.Header.Add("Authorization", authorization)

	egressUrl := r.URL
	log.Printf("egressUrl -> %+v\n", egressUrl)
	newUrl := egressUrl.Path
	newUrl = strings.Replace(newUrl, "/mesh/", "", 1)
	split := strings.SplitN(newUrl, "/", 2)
	service := split[0]
	path := split[1]
	resp, err := forwardRequest(w, r, service, path)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	for header, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}

	w.Write(resBody)
}

func forwardRequest(w http.ResponseWriter, req *http.Request, service string, path string) (*http.Response, error) {
	log.Println("Forwarding request")
	// we need to buffer the body if we want to read it here and send it
	// in the request.
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}

//	url := fmt.Sprintf("%s://%s%s/%s", "http", "localhost", ":8083", path), local env
	// create a new url from the raw RequestURI sent by the client
	url := fmt.Sprintf("%s://%s%s/%s", "http", service, ":8080", path)


	proxyReq, err := http.NewRequest(req.Method, url, bytes.NewReader(body))

	// We may want to filter some headers, otherwise we could just use a shallow copy
	proxyReq.Header = req.Header

	httpClient := http.Client{}
	resp, err := httpClient.Do(proxyReq)
	if err != nil {
		return nil, err
	}

	return resp, nil
}