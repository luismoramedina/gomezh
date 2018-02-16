package main

import "net/http"
import "fmt"
import "github.com/opentracing/opentracing-go"
import zipkin "github.com/openzipkin/zipkin-go-opentracing"

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
	sp := tracer.StartSpan("request")
	spc := sp.Context().(zipkin.SpanContext)
	fmt.Printf("%+v\n", spc)
	fmt.Println(spc.TraceID.Low)
	defer sp.Finish()

	w.Write([]byte(fmt.Sprintf("Hello, World! %v", spc.TraceID.Low)))
}

/*
func makeSomeRequest(ctx context.Context) {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		httpClient := &http.Client{}
		httpReq, _ := http.NewRequest("GET", "http://myservice/", nil)

		// Transmit the span's TraceContext as HTTP headers on our
		// outbound request.
		opentracing.GlobalTracer().Inject(
			span.Context(),
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(httpReq.Header))

		resp, err := httpClient.Do(httpReq)
	}
}
*/
