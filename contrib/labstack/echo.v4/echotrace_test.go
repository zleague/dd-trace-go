// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016 Datadog, Inc.

package echo

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	pappsec "github.com/zleague/dd-trace-go/appsec"
	"github.com/zleague/dd-trace-go/ddtrace/ext"
	"github.com/zleague/dd-trace-go/ddtrace/mocktracer"
	"github.com/zleague/dd-trace-go/ddtrace/tracer"
	"github.com/zleague/dd-trace-go/internal/appsec"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChildSpan(t *testing.T) {
	assert := assert.New(t)
	mt := mocktracer.Start()
	defer mt.Stop()
	var called, traced bool

	router := echo.New()
	router.Use(Middleware(WithServiceName("foobar")))
	router.GET("/user/:id", func(c echo.Context) error {
		called = true
		_, traced = tracer.SpanFromContext(c.Request().Context())
		return c.NoContent(200)
	})

	r := httptest.NewRequest("GET", "/user/123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	// verify traces look good
	assert.True(called)
	assert.True(traced)
}

func TestTrace200(t *testing.T) {
	assert := assert.New(t)
	mt := mocktracer.Start()
	defer mt.Stop()
	var called, traced bool

	router := echo.New()
	router.Use(Middleware(WithServiceName("foobar"), WithAnalytics(false)))
	router.GET("/user/:id", func(c echo.Context) error {
		called = true
		var span tracer.Span
		span, traced = tracer.SpanFromContext(c.Request().Context())

		// we patch the span on the request context.
		span.SetTag("test.echo", "echony")
		assert.Equal(span.(mocktracer.Span).Tag(ext.ServiceName), "foobar")
		return c.NoContent(200)
	})

	root := tracer.StartSpan("root")
	r := httptest.NewRequest("GET", "/user/123", nil)
	err := tracer.Inject(root.Context(), tracer.HTTPHeadersCarrier(r.Header))
	assert.Nil(err)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	// verify traces look good
	assert.True(called)
	assert.True(traced)

	spans := mt.FinishedSpans()
	assert.Len(spans, 1)

	span := spans[0]
	assert.Equal("http.request", span.OperationName())
	assert.Equal(ext.SpanTypeWeb, span.Tag(ext.SpanType))
	assert.Equal("foobar", span.Tag(ext.ServiceName))
	assert.Equal("echony", span.Tag("test.echo"))
	assert.Contains(span.Tag(ext.ResourceName), "/user/:id")
	assert.Equal("200", span.Tag(ext.HTTPCode))
	assert.Equal("GET", span.Tag(ext.HTTPMethod))
	assert.Equal(root.Context().SpanID(), span.ParentID())

	assert.Equal("http://example.com/user/123", span.Tag(ext.HTTPURL))
}

func TestTraceAnalytics(t *testing.T) {
	assert := assert.New(t)
	mt := mocktracer.Start()
	defer mt.Stop()
	var called, traced bool

	router := echo.New()
	router.Use(Middleware(WithServiceName("foobar"), WithAnalytics(true)))
	router.GET("/user/:id", func(c echo.Context) error {
		called = true
		var span tracer.Span
		span, traced = tracer.SpanFromContext(c.Request().Context())

		// we patch the span on the request context.
		span.SetTag("test.echo", "echony")
		assert.Equal(span.(mocktracer.Span).Tag(ext.ServiceName), "foobar")
		return c.NoContent(200)
	})

	root := tracer.StartSpan("root")
	r := httptest.NewRequest("GET", "/user/123", nil)
	err := tracer.Inject(root.Context(), tracer.HTTPHeadersCarrier(r.Header))
	assert.Nil(err)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	// verify traces look good
	assert.True(called)
	assert.True(traced)

	spans := mt.FinishedSpans()
	assert.Len(spans, 1)

	span := spans[0]
	assert.Equal("http.request", span.OperationName())
	assert.Equal(ext.SpanTypeWeb, span.Tag(ext.SpanType))
	assert.Equal("foobar", span.Tag(ext.ServiceName))
	assert.Equal("echony", span.Tag("test.echo"))
	assert.Contains(span.Tag(ext.ResourceName), "/user/:id")
	assert.Equal("200", span.Tag(ext.HTTPCode))
	assert.Equal("GET", span.Tag(ext.HTTPMethod))
	assert.Equal(1.0, span.Tag(ext.EventSampleRate))
	assert.Equal(root.Context().SpanID(), span.ParentID())

	assert.Equal("http://example.com/user/123", span.Tag(ext.HTTPURL))
}

func TestError(t *testing.T) {
	for _, tt := range []struct {
		isStatusError func(statusCode int) bool
		err           error
		code          string
		handler       func(c echo.Context) error
	}{
		{
			err:  errors.New("oh no"),
			code: "500",
			handler: func(c echo.Context) error {
				return errors.New("oh no")
			},
		},
		{
			err:  echo.NewHTTPError(http.StatusInternalServerError, "my error message"),
			code: "500",
			handler: func(c echo.Context) error {
				return echo.NewHTTPError(http.StatusInternalServerError, "my error message")
			},
		},
		{
			err:  nil,
			code: "400",
			handler: func(c echo.Context) error {
				return echo.NewHTTPError(http.StatusBadRequest, "my error message")
			},
		},
		{
			isStatusError: func(statusCode int) bool { return statusCode >= 400 && statusCode < 500 },
			err:           errors.New("500: Internal Server Error"),
			code:          "500",
			handler: func(c echo.Context) error {
				return errors.New("oh no")
			},
		},
		{
			isStatusError: func(statusCode int) bool { return statusCode >= 400 && statusCode < 500 },
			err:           errors.New("500: Internal Server Error"),
			code:          "500",
			handler: func(c echo.Context) error {
				return echo.NewHTTPError(http.StatusInternalServerError, "my error message")
			},
		},
		{
			isStatusError: func(statusCode int) bool { return statusCode >= 400 },
			err:           echo.NewHTTPError(http.StatusBadRequest, "my error message"),
			code:          "400",
			handler: func(c echo.Context) error {
				return echo.NewHTTPError(http.StatusBadRequest, "my error message")
			},
		},
		{
			isStatusError: func(statusCode int) bool { return statusCode >= 200 },
			err:           fmt.Errorf("201: Created"),
			code:          "201",
			handler: func(c echo.Context) error {
				c.JSON(201, map[string]string{"status": "ok", "type": "test"})
				return nil
			},
		},
		{
			isStatusError: func(statusCode int) bool { return statusCode >= 200 },
			err:           fmt.Errorf("200: OK"),
			code:          "200",
			handler: func(c echo.Context) error {
				// It's not clear if unset (0) status is possible naturally, but we can simulate that situation.
				c.Response().Status = 0
				return nil
			},
		},
	} {
		t.Run("", func(t *testing.T) {
			assert := assert.New(t)
			mt := mocktracer.Start()
			defer mt.Stop()

			router := echo.New()
			opts := []Option{WithServiceName("foobar")}
			if tt.isStatusError != nil {
				opts = append(opts, WithStatusCheck(tt.isStatusError))
			}
			router.Use(Middleware(opts...))
			router.GET("/err", tt.handler)
			r := httptest.NewRequest("GET", "/err", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)

			spans := mt.FinishedSpans()
			assert.Len(spans, 1)
			span := spans[0]
			assert.Equal("http.request", span.OperationName())
			assert.Equal(ext.SpanTypeWeb, span.Tag(ext.SpanType))
			assert.Equal("foobar", span.Tag(ext.ServiceName))
			assert.Contains(span.Tag(ext.ResourceName), "/err")
			assert.Equal(tt.code, span.Tag(ext.HTTPCode))
			assert.Equal("GET", span.Tag(ext.HTTPMethod))
			err := span.Tag(ext.Error)
			if tt.err != nil {
				if !assert.NotNil(err) {
					return
				}
				assert.Equal(tt.err.Error(), err.(error).Error())
			} else {
				assert.Nil(err)
			}
		})
	}
}

func TestGetSpanNotInstrumented(t *testing.T) {
	assert := assert.New(t)
	router := echo.New()
	var called, traced bool

	router.GET("/ping", func(c echo.Context) error {
		// Assert we don't have a span on the context.
		called = true
		_, traced = tracer.SpanFromContext(c.Request().Context())
		return c.NoContent(200)
	})

	r := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)
	assert.True(called)
	assert.False(traced)
}

func TestNoDebugStack(t *testing.T) {
	assert := assert.New(t)
	mt := mocktracer.Start()
	defer mt.Stop()
	var called, traced bool

	// setup
	router := echo.New()
	router.Use(Middleware(NoDebugStack()))
	wantErr := errors.New("oh no")

	// a handler with an error and make the requests
	router.GET("/err", func(c echo.Context) error {
		_, traced = tracer.SpanFromContext(c.Request().Context())
		called = true

		err := wantErr
		c.Error(err)
		return err
	})
	r := httptest.NewRequest("GET", "/err", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	// verify the error is correct and the stacktrace is disabled
	assert.True(called)
	assert.True(traced)

	spans := mt.FinishedSpans()
	assert.Len(spans, 1)

	span := spans[0]
	assert.Equal(wantErr.Error(), span.Tag(ext.Error).(error).Error())
	assert.Equal("<debug stack disabled>", span.Tag(ext.ErrorStack))
}

func TestIgnoreRequestFunc(t *testing.T) {
	assert := assert.New(t)
	mt := mocktracer.Start()
	defer mt.Stop()
	var called, traced bool

	// setup
	ignoreRequestFunc := func(c echo.Context) bool {
		return true
	}
	router := echo.New()
	router.Use(Middleware(WithIgnoreRequest(ignoreRequestFunc)))

	// a handler with an error and make the requests
	router.GET("/err", func(c echo.Context) error {
		_, traced = tracer.SpanFromContext(c.Request().Context())
		called = true
		return nil
	})
	r := httptest.NewRequest("GET", "/err", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	// verify the error is correct and the stacktrace is disabled
	assert.True(called)
	assert.False(traced)

	spans := mt.FinishedSpans()
	assert.Len(spans, 0)
}

func TestAppSec(t *testing.T) {
	appsec.Start()
	defer appsec.Stop()

	if !appsec.Enabled() {
		t.Skip("appsec disabled")
	}

	// Start and trace an HTTP server
	e := echo.New()
	e.Use(Middleware())

	// Add some testing routes
	e.POST("/path0.0/:myPathParam0/path0.1/:myPathParam1/path0.2/:myPathParam2/path0.3/*myPathParam3", func(c echo.Context) error {
		return c.String(200, "Hello World!\n")
	})
	e.POST("/", func(c echo.Context) error {
		return c.String(200, "Hello World!\n")
	})
	e.POST("/body", func(c echo.Context) error {
		pappsec.MonitorParsedHTTPBody(c.Request().Context(), "$globals")
		return c.String(200, "Hello Body!\n")
	})
	srv := httptest.NewServer(e)
	defer srv.Close()

	// Test an LFI attack via path parameters
	t.Run("request-uri", func(t *testing.T) {
		mt := mocktracer.Start()
		defer mt.Stop()
		// Send an LFI attack (according to appsec rule id crs-930-110)
		req, err := http.NewRequest("POST", srv.URL+"/../../../secret.txt", nil)
		if err != nil {
			panic(err)
		}
		res, err := srv.Client().Do(req)
		require.NoError(t, err)
		// Check that the server behaved as intended (no 301 but 404 directly)
		require.Equal(t, http.StatusNotFound, res.StatusCode)
		// The span should contain the security event
		finished := mt.FinishedSpans()
		require.Len(t, finished, 1)
		event := finished[0].Tag("_dd.appsec.json").(string)
		require.NotNil(t, event)
		require.True(t, strings.Contains(event, "crs-930-110"))
		require.True(t, strings.Contains(event, "server.request.uri.raw"))
	})

	// Test a security scanner attack via path parameters
	t.Run("path-params", func(t *testing.T) {
		t.Run("regular", func(t *testing.T) {
			mt := mocktracer.Start()
			defer mt.Stop()
			// Send a security scanner attack (according to appsec rule id crs-913-120)
			req, err := http.NewRequest("POST", srv.URL+"/path0.0/param0/path0.1/param1/path0.2/appscan_fingerprint/path0.3/param3", nil)
			if err != nil {
				panic(err)
			}
			res, err := srv.Client().Do(req)
			require.NoError(t, err)
			// Check that the handler was properly called
			b, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			require.Equal(t, "Hello World!\n", string(b))
			require.Equal(t, http.StatusOK, res.StatusCode)
			// The span should contain the security event
			finished := mt.FinishedSpans()
			require.Len(t, finished, 1)
			event := finished[0].Tag("_dd.appsec.json").(string)
			require.NotNil(t, event)
			require.True(t, strings.Contains(event, "crs-913-120"))
			require.True(t, strings.Contains(event, "myPathParam2"))
			require.True(t, strings.Contains(event, "server.request.path_params"))
		})

		t.Run("wildcard", func(t *testing.T) {
			mt := mocktracer.Start()
			defer mt.Stop()
			// Send a security scanner attack (according to appsec rule id crs-913-120)
			req, err := http.NewRequest("POST", srv.URL+"/path0.0/param0/path0.1/param1/path0.2/param2/path0.3/appscan_fingerprint", nil)
			if err != nil {
				panic(err)
			}
			res, err := srv.Client().Do(req)
			require.NoError(t, err)
			// Check that the handler was properly called
			b, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			require.Equal(t, "Hello World!\n", string(b))
			require.Equal(t, http.StatusOK, res.StatusCode)
			// The span should contain the security event
			finished := mt.FinishedSpans()
			require.Len(t, finished, 1)
			event := finished[0].Tag("_dd.appsec.json").(string)
			require.NotNil(t, event)
			require.True(t, strings.Contains(event, "crs-913-120"))
			// Wildcards are not named in echo
			require.False(t, strings.Contains(event, "myPathParam3"))
			require.True(t, strings.Contains(event, "server.request.path_params"))
		})
	})

	t.Run("response-status", func(t *testing.T) {
		mt := mocktracer.Start()
		defer mt.Stop()

		req, err := http.NewRequest("POST", srv.URL+"/etc/", nil)
		if err != nil {
			panic(err)
		}
		res, err := srv.Client().Do(req)
		require.NoError(t, err)
		require.Equal(t, 404, res.StatusCode)

		finished := mt.FinishedSpans()
		require.Len(t, finished, 1)
		event := finished[0].Tag("_dd.appsec.json").(string)
		require.NotNil(t, event)
		require.True(t, strings.Contains(event, "server.response.status"))
		require.True(t, strings.Contains(event, "nfd-000-001"))
	})

	// Test a PHP injection attack via request parsed body
	t.Run("SDK-body", func(t *testing.T) {
		mt := mocktracer.Start()
		defer mt.Stop()

		req, err := http.NewRequest("POST", srv.URL+"/body", nil)
		if err != nil {
			panic(err)
		}
		res, err := srv.Client().Do(req)
		require.NoError(t, err)
		// Check that the handler was properly called
		b, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		require.Equal(t, "Hello Body!\n", string(b))

		finished := mt.FinishedSpans()
		require.Len(t, finished, 1)
		event := finished[0].Tag("_dd.appsec.json")
		require.NotNil(t, event)
		require.True(t, strings.Contains(event.(string), "crs-933-130"))
	})
}
