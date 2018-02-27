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
	"github.com/luismoramedina/gomezh/sidecar"
	"time"
	"os"
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

	egressUrl := r.URL
	log.Printf("egressUrl -> %+v\n", egressUrl)
	newUrl := egressUrl.Path
	newUrl = strings.Replace(newUrl, "/mesh/", "", 1)
	split := strings.SplitN(newUrl, "/", 2)
	service := split[0]
	path := split[1]
	resp, err := forwardRequest(w, r, service, path)
	defer s.showElapsed(traceId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		log.Println(err)
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

func forwardRequest(w http.ResponseWriter, req *http.Request, service string, path string) (*http.Response, error) {
	log.Println("Forwarding request")
	// we need to buffer the body if we want to read it here and send it
	// in the request.
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}

	var url string
	if os.Getenv("ENV") == "local" {
		url = fmt.Sprintf("%s://%s%s/%s", "http", "localhost", ":8083", path)//, local env
	} else {
		// create a new url from the raw RequestURI sent by the client
		url = fmt.Sprintf("%s://%s%s/%s", "http", service, ":8080", path)
	}


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