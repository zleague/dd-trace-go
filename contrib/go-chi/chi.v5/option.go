// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016 Datadog, Inc.

package chi

import (
	"math"
	"net/http"

	"github.com/zleague/dd-trace-go/ddtrace"
	"github.com/zleague/dd-trace-go/internal"
	"github.com/zleague/dd-trace-go/internal/globalconfig"
)

type config struct {
	serviceName   string
	spanOpts      []ddtrace.StartSpanOption // additional span options to be applied
	analyticsRate float64
	isStatusError func(statusCode int) bool
	ignoreRequest func(r *http.Request) bool
}

// Option represents an option that can be passed to NewRouter.
type Option func(*config)

func defaults(cfg *config) {
	cfg.serviceName = "chi.router"
	if svc := globalconfig.ServiceName(); svc != "" {
		cfg.serviceName = svc
	}
	if internal.BoolEnv("DD_TRACE_CHI_ANALYTICS_ENABLED", false) {
		cfg.analyticsRate = 1.0
	} else {
		cfg.analyticsRate = globalconfig.AnalyticsRate()
	}
	cfg.isStatusError = isServerError
	cfg.ignoreRequest = func(_ *http.Request) bool { return false }
}

// WithServiceName sets the given service name for the router.
func WithServiceName(name string) Option {
	return func(cfg *config) {
		cfg.serviceName = name
	}
}

// WithSpanOptions applies the given set of options to the spans started
// by the router.
func WithSpanOptions(opts ...ddtrace.StartSpanOption) Option {
	return func(cfg *config) {
		cfg.spanOpts = opts
	}
}

// WithAnalytics enables Trace Analytics for all started spans.
func WithAnalytics(on bool) Option {
	return func(cfg *config) {
		if on {
			cfg.analyticsRate = 1.0
		} else {
			cfg.analyticsRate = math.NaN()
		}
	}
}

// WithAnalyticsRate sets the sampling rate for Trace Analytics events
// correlated to started spans.
func WithAnalyticsRate(rate float64) Option {
	return func(cfg *config) {
		if rate >= 0.0 && rate <= 1.0 {
			cfg.analyticsRate = rate
		} else {
			cfg.analyticsRate = math.NaN()
		}
	}
}

// WithStatusCheck specifies a function fn which reports whether the passed
// statusCode should be considered an error.
func WithStatusCheck(fn func(statusCode int) bool) Option {
	return func(cfg *config) {
		cfg.isStatusError = fn
	}
}

func isServerError(statusCode int) bool {
	return statusCode >= 500 && statusCode < 600
}

// WithIgnoreRequest specifies a function to use for determining if the
// incoming HTTP request tracing should be skipped.
func WithIgnoreRequest(fn func(r *http.Request) bool) Option {
	return func(cfg *config) {
		cfg.ignoreRequest = fn
	}
}
