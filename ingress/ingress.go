package ingress

import (
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	"io/ioutil"
	"bytes"
	"github.com/opentracing/opentracing-go"
	"fmt"
	"net/http"
	"log"
	"github.com/luismoramedina/gomesh/sidecar"
	myjwt "github.com/luismoramedina/gomesh/jwt"
	"time"
	"github.com/dgrijalva/jwt-go"
	"crypto/rsa"
	"os"
	"encoding/json"
)

var rsaPublicKey *rsa.PublicKey

const defaultPKey string = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA4f5wg5l2hKsTeNem/V41
fGnJm6gOdrj8ym3rFkEU/wT8RDtnSgFEZOQpHEgQ7JL38xUfU0Y3g6aYw9QT0hJ7
mCpz9Er5qLaMXJwZxzHzAahlfA0icqabvJOMvQtzD6uQv6wPEyZtDTWiQi9AXwBp
HssPnpYGIn20ZZuNlX2BrClciHhCPUIIZOQn/MmqTD31jSyjoQoV7MhhMTATKJx2
XrHhR+1DcKJzQBSTAGnpYVaqpsARap+nwRipr3nUTuxyGohBTSmjJ2usSeQXHI3b
ODIRe1AuTyHceAbewn8b462yEWKARdpd9AjQW5SIVPfdsz5B6GlYQ5LdYKtznTuy
7wIDAQAB
-----END PUBLIC KEY-----`

var jwtValidator myjwt.JwtValidator

func init() {
	log.Println("init ingress, loading public key")
	var e error
	rsaPublicKey, e = jwt.ParseRSAPublicKeyFromPEM([]byte(os.Getenv("PUBLIC_KEY")))
	if e != nil {
		rsaPublicKey, e = jwt.ParseRSAPublicKeyFromPEM([]byte(defaultPKey))
		log.Println("Default key loaded")
	}
	jwtValidator = myjwt.JwtValidator{PublicKey: rsaPublicKey}
}

type IngressController struct {
	sidecar.Sidecar
}

func (s IngressController) Handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	log.Println("Starting ingress request")
	span := s.Tracer.StartSpan("request")
	defer span.Finish()
	reqId := span.Context().(zipkin.SpanContext).TraceID.Low
	log.Printf("Traceid: %d", reqId)

	s.Times.Put(reqId, start)
	secContext := sidecar.SecurityContext{}
	secContext.Token = r.Header.Get("Authorization")
	secContext.PlainContext = r.Header.Get("Plain-Authorization")
	s.Auths.Put(reqId, secContext)
	defer s.showElapsed(reqId, start)

	resp, err := s.forwardRequest(w, r, span)
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

	w.WriteHeader(resp.StatusCode)
	w.Write(resBody)

}
func SecurityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Executing security middleware")
		ok, claims := jwtValidator.IsValidCredential(r.Header.Get("Authorization"));
		if !ok {
			http.Error(w, "not valid credentials or no credentials", http.StatusForbidden)
			return
		}
		plainSecContext, err := json.Marshal(claims)
		if err == nil {
			r.Header.Add("Plain-authorization", string(plainSecContext))
		}
		next.ServeHTTP(w, r)
	})
}

func (s IngressController) showElapsed(traceId uint64, start time.Time) {
	s.Times.Delete(traceId)
	elapsed := time.Now().Sub(start)
	log.Printf("[TIME] request %d -> %f", traceId, elapsed.Seconds())
}

//https://stackoverflow.com/questions/34724160/go-http-send-incoming-http-request-to-an-other-server-using-client-do
func (s IngressController) forwardRequest(w http.ResponseWriter, req *http.Request, span opentracing.Span) (*http.Response, error) {
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
	s.Tracer.Inject(
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