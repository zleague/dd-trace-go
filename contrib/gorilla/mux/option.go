// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016 Datadog, Inc.

package mux

import (
	"math"
	"net/http"

	"github.com/zleague/dd-trace-go/ddtrace"
	"github.com/zleague/dd-trace-go/ddtrace/ext"
	"github.com/zleague/dd-trace-go/ddtrace/tracer"
	"github.com/zleague/dd-trace-go/internal"
	"github.com/zleague/dd-trace-go/internal/globalconfig"
)

type routerConfig struct {
	serviceName   string
	spanOpts      []ddtrace.StartSpanOption // additional span options to be applied
	finishOpts    []ddtrace.FinishOption    // span finish options to be applied
	analyticsRate float64
	resourceNamer func(*Router, *http.Request) string
	ignoreRequest func(*http.Request) bool
	headerTags    bool
	queryParams   bool
}

// RouterOption represents an option that can be passed to NewRouter.
type RouterOption func(*routerConfig)

func newConfig(opts []RouterOption) *routerConfig {
	cfg := new(routerConfig)
	defaults(cfg)
	for _, fn := range opts {
		fn(cfg)
	}
	if !math.IsNaN(cfg.analyticsRate) {
		cfg.spanOpts = append(cfg.spanOpts, tracer.Tag(ext.EventSampleRate, cfg.analyticsRate))
	}
	return cfg
}

func defaults(cfg *routerConfig) {
	if internal.BoolEnv("DD_TRACE_MUX_ANALYTICS_ENABLED", false) {
		cfg.analyticsRate = 1.0
	} else {
		cfg.analyticsRate = globalconfig.AnalyticsRate()
	}
	cfg.serviceName = "mux.router"
	if svc := globalconfig.ServiceName(); svc != "" {
		cfg.serviceName = svc
	}
	cfg.resourceNamer = defaultResourceNamer
	cfg.ignoreRequest = func(_ *http.Request) bool { return false }
}

// WithIgnoreRequest holds the function to use for determining if the
// incoming HTTP request tracing should be skipped.
func WithIgnoreRequest(f func(*http.Request) bool) RouterOption {
	return func(cfg *routerConfig) {
		cfg.ignoreRequest = f
	}
}

// WithServiceName sets the given service name for the router.
func WithServiceName(name string) RouterOption {
	return func(cfg *routerConfig) {
		cfg.serviceName = name
	}
}

// WithSpanOptions applies the given set of options to the spans started
// by the router.
func WithSpanOptions(opts ...ddtrace.StartSpanOption) RouterOption {
	return func(cfg *routerConfig) {
		cfg.spanOpts = opts
	}
}

// NoDebugStack prevents stack traces from being attached to spans finishing
// with an error. This is useful in situations where errors are frequent and
// performance is critical.
func NoDebugStack() RouterOption {
	return func(cfg *routerConfig) {
		cfg.finishOpts = append(cfg.finishOpts, tracer.NoDebugStack())
	}
}

// WithAnalytics enables Trace Analytics for all started spans.
func WithAnalytics(on bool) RouterOption {
	return func(cfg *routerConfig) {
		if on {
			cfg.analyticsRate = 1.0
		} else {
			cfg.analyticsRate = math.NaN()
		}
	}
}

// WithAnalyticsRate sets the sampling rate for Trace Analytics events
// correlated to started spans.
func WithAnalyticsRate(rate float64) RouterOption {
	return func(cfg *routerConfig) {
		if rate >= 0.0 && rate <= 1.0 {
			cfg.analyticsRate = rate
		} else {
			cfg.analyticsRate = math.NaN()
		}
	}
}

// WithResourceNamer specifies a quantizing function which will be used to
// obtain the resource name for a given request.
func WithResourceNamer(namer func(router *Router, req *http.Request) string) RouterOption {
	return func(cfg *routerConfig) {
		cfg.resourceNamer = namer
	}
}

// WithHeaderTags specifies that the integration should attach HTTP request headers as
// tags to spans.
// Warning: using this feature can risk exposing sensitive data such as authorisation tokens
// to Datadog.
func WithHeaderTags() RouterOption {
	return func(cfg *routerConfig) {
		cfg.headerTags = true
	}
}

// WithQueryParams specifies that the integration should attach request query parameters as APM tags.
// Warning: using this feature can risk exposing sensitive data such as authorisation tokens
// to Datadog.
func WithQueryParams() RouterOption {
	return func(cfg *routerConfig) {
		cfg.queryParams = true
	}
}
