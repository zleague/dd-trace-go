// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016 Datadog, Inc.

package gorm_test

import (
	"log"

	"github.com/jinzhu/gorm"
	"github.com/lib/pq"

	sqltrace "github.com/zleague/dd-trace-go/contrib/database/sql"
	gormtrace "github.com/zleague/dd-trace-go/contrib/jinzhu/gorm"
)

func ExampleOpen() {
	// Register augments the provided driver with tracing, enabling it to be loaded by gormtrace.Open.
	sqltrace.Register("postgres", &pq.Driver{}, sqltrace.WithServiceName("my-service"))

	// Open the registered driver, allowing all uses of the returned *gorm.DB to be traced.
	db, err := gormtrace.Open("postgres", "postgres://pqgotest:password@localhost/pqgotest?sslmode=disable")
	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}

	user := struct {
		gorm.Model
		Name string
	}{}

	// All calls through gorm.DB are now traced.
	db.Where("name = ?", "jinzhu").First(&user)
}
