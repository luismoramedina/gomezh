package egress

import (
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	"net/http"
	"github.com/opentracing/opentracing-go"
	"log"
	"io/ioutil"
	"github.com/luismoramedina/gomezh/sidecar"
	"time"
	"bytes"
	"fmt"
)

type EgressController struct {
	sidecar.Sidecar
}

func (s EgressController) Handler(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting Egress request")
	wireContext, err := s.Tracer.Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(r.Header))
	if err != nil {
		http.Error(w, "", http.StatusServiceUnavailable)
		return
	}
	traceId := wireContext.(zipkin.SpanContext).TraceID.Low
	log.Printf("Trace id: %x", traceId)

	secContext := sidecar.SecurityContext(s.Auths.Get(traceId))
	defer s.Auths.Delete(traceId)

	log.Printf("Injecting authorization: %s", secContext.Token)

	r.Header.Add("Authorization", secContext.Token)

	resp, err := forwardRequest(w, r)
	defer s.showElapsed(traceId)

	if err != nil {
		log.Println(err)
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

func (s EgressController) showElapsed(traceId uint64) {
	start := s.Times.Get(traceId)
	s.Times.Delete(traceId)
	elapsed := time.Now().Sub(start)
	log.Printf("[TIME-ie] request %x -> %f", traceId, elapsed.Seconds())
}

func forwardRequest(w http.ResponseWriter, req *http.Request) (*http.Response, error) {
	log.Println("Forwarding request")

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}

	// create a new url from the raw RequestURI sent by the client
	url := fmt.Sprintf("%s://%s%s", "http", req.Host, req.RequestURI)

	log.Printf("egress Forwarding -> %s\n", url)
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