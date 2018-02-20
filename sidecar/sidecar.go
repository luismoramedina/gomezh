package sidecar

import (
	"github.com/opentracing/opentracing-go"
)

type Sidecar struct {
	Tracer opentracing.Tracer
	Auths map[uint64]string
}