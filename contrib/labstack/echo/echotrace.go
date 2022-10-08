// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016 Datadog, Inc.

// Package echo provides functions to trace the labstack/echo package (https://github.com/labstack/echo).
package echo

import (
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/zleague/dd-trace-go/contrib/internal/httptrace"
	"github.com/zleague/dd-trace-go/ddtrace"
	"github.com/zleague/dd-trace-go/ddtrace/ext"
	"github.com/zleague/dd-trace-go/ddtrace/tracer"
	"github.com/zleague/dd-trace-go/internal/log"

	"github.com/labstack/echo"
)

// Middleware returns echo middleware which will trace incoming requests.
func Middleware(opts ...Option) echo.MiddlewareFunc {
	cfg := new(config)
	defaults(cfg)
	for _, fn := range opts {
		fn(cfg)
	}
	log.Debug("contrib/labstack/echo: Configuring Middleware: %#v", cfg)
	spanOpts := []ddtrace.StartSpanOption{
		tracer.ServiceName(cfg.serviceName),
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			request := c.Request()
			resource := request.Method + " " + c.Path()
			opts := append(spanOpts, tracer.ResourceName(resource))

			if !math.IsNaN(cfg.analyticsRate) {
				opts = append(opts, tracer.Tag(ext.EventSampleRate, cfg.analyticsRate))
			}

			var finishOpts []tracer.FinishOption
			if cfg.noDebugStack {
				finishOpts = []tracer.FinishOption{tracer.NoDebugStack()}
			}

			span, ctx := httptrace.StartRequestSpan(request, opts...)
			defer func() {
				httptrace.FinishRequestSpan(span, c.Response().Status, finishOpts...)
			}()

			// pass the span through the request context
			c.SetRequest(request.WithContext(ctx))

			// serve the request to the next middleware
			err := next(c)
			if err != nil {
				// It is impossible to determine what the final status code of a request is in echo.
				// This is the best we can do.
				switch err := err.(type) {
				case *echo.HTTPError:
					if cfg.isStatusError(err.Code) {
						// mark 5xx server error
						span.SetTag(ext.Error, err)
					}
					span.SetTag(ext.HTTPCode, strconv.Itoa(err.Code))
				default:
					// Any non-HTTPError errors appear as 5xx errors.
					if cfg.isStatusError(500) {
						span.SetTag(ext.Error, err)
					}
					span.SetTag(ext.HTTPCode, "500")
				}
			} else if status := c.Response().Status; status > 0 {
				if cfg.isStatusError(status) {
					// mark 5xx server error
					span.SetTag(ext.Error, fmt.Errorf("%d: %s", status, http.StatusText(status)))
				}
				span.SetTag(ext.HTTPCode, strconv.Itoa(status))
			} else {
				if cfg.isStatusError(200) {
					// mark 5xx server error
					span.SetTag(ext.Error, fmt.Errorf("%d: %s", 200, http.StatusText(200)))
				}
				span.SetTag(ext.HTTPCode, "200")
			}
			return err
		}
	}
}
