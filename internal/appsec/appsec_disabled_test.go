// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016 Datadog, Inc.

//go:build !appsec
// +build !appsec

package appsec_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zleague/dd-trace-go/ddtrace/tracer"
	"github.com/zleague/dd-trace-go/internal/appsec"

	"github.com/stretchr/testify/require"
)

func TestEnabled(t *testing.T) {
	enabledStr := os.Getenv("DD_APPSEC_ENABLED")
	if enabledStr != "" {
		defer os.Setenv("DD_APPSEC_ENABLED", enabledStr)
	}
	// AppSec should be always disabled
	require.False(t, appsec.Enabled())
	tracer.Start()
	assert.False(t, appsec.Enabled())
	tracer.Stop()
	assert.False(t, appsec.Enabled())
	os.Setenv("DD_APPSEC_ENABLED", "true")
	require.False(t, appsec.Enabled())
	tracer.Start()
	assert.False(t, appsec.Enabled())
	tracer.Stop()
	assert.False(t, appsec.Enabled())

}
