[![CircleCI](https://circleci.com/gh/DataDog/dd-trace-go/tree/v1.svg?style=svg)](https://circleci.com/gh/DataDog/dd-trace-go/tree/v1)
[![Godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://pkg.go.dev/github.com/zleague/dd-trace-go)
[![codecov](https://codecov.io/gh/DataDog/dd-trace-go/branch/v1/graph/badge.svg?token=jGG20Xhv8i)](https://codecov.io/gh/DataDog/dd-trace-go)

### Datadog Client Libraries for Go

This repository contains Go packages for the client-side components of the Datadog product suite for Application Performance Monitoring, Continuous Profiling and Application Security Monitoring of Go applications.

- [Datadog Application Performance Monitoring (APM)](https://docs.datadoghq.com/tracing/): Trace requests as they flow across web servers, databases and microservices so that developers have great visiblity into bottlenecks and troublesome requests.  
The package [`github.com/zleague/dd-trace-go/ddtrace/tracer`](https://pkg.go.dev/github.com/zleague/dd-trace-go/ddtrace/tracer) allows you to trace any piece of your Go code, and commonly used Go libraries can be automatically traced thanks to our out-of-the-box integrations which can be found in the package [`github.com/zleague/dd-trace-go/ddtrace/contrib`](https://pkg.go.dev/github.com/zleague/dd-trace-go/contrib).

- [Datadog Go Continuous Profiler](https://docs.datadoghq.com/profiler/): Continuously profile your Go apps to find CPU, memory, and synchronization bottlenecks, broken down by function name, and line number, to significantly reduce end-user latency and infrastructure costs.  
The package [`github.com/zleague/dd-trace-go/profiler`](https://pkg.go.dev/github.com/zleague/dd-trace-go/profiler) allows you to periodically collect and send Go profiles to the Datadog API.

- [Datadog Application Security Monitoring (ASM)](https://docs.datadoghq.com/security_platform/application_security/): Get observability into your Go apps attacks that aim to exploit code-level vulnerabilities, such as Server-Side-Request-Forgery (SSRF), SQL injection, Log4Shell, and Reflected Cross-Site-Scripting (XSS). ASM is not a standalone Go package and is transparently integrated into the APM tracer. Simply [enable it](https://docs.datadoghq.com/security_platform/application_security/setup_and_configure/?code-lang=go) to benefit from the security monitoring.

### Installing

This module contains many packages, but most users should probably install the two packages below:

```bash
go get github.com/zleague/dd-trace-go/ddtrace/tracer
go get github.com/zleague/dd-trace-go/profiler
```

Additionally there are many [contrib](./contrib) packages that can be installed to automatically instrument and trace commonly used Go libraries such as [net/http](https://pkg.go.dev/github.com/zleague/dd-trace-go/contrib/net/http), [gorilla/mux](https://pkg.go.dev/github.com/zleague/dd-trace-go/contrib/gorilla/mux) or [database/sql](https://pkg.go.dev/github.com/zleague/dd-trace-go/contrib/database/sql):

```
go get github.com/zleague/dd-trace-go/contrib/gorilla/mux
```

If you installed more packages than you intended, you can use `go mod tidy` to remove any unused packages.

### Documentation

 - [APM Tracing API](https://pkg.go.dev/github.com/zleague/dd-trace-go/ddtrace)
 - [APM Tracing Go Applications](https://docs.datadoghq.com/tracing/setup/go/)
 - [Continuous Go Profiler](https://docs.datadoghq.com/tracing/profiler/enabling/go)
 - [Application Security Monitoring](https://docs.datadoghq.com/security_platform/application_security/setup_and_configure/?code-lang=go)
 - If you are migrating from an older version of the tracer (e.g. 0.6.x) you may also find the [migration document](MIGRATING.md) we've put together helpful.

### Support Policy

Datadog APM for Go is built upon dependencies defined in specific versions of the host operating system, Go releases, and the Datadog Agent/API. For Go the two latest releases are [GA](#support-ga) supported and the version before that is in [Maintenance](#support-maintenance). We do make efforts to support older releases, but generally these releases are considered [Legacy](#support-legacy). This library only officially supports [first class ports](https://github.com/golang/go/wiki/PortingPolicy#first-class-ports) of Go.

| **Level**                                              | **Support provided**                                                                                                                                                         |
|--------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| <span id="support-ga">General Availability (GA)</span> | Full implementation of all features. Full support for new features, bug & security fixes.                                                                                    |
| <span id="support-maintenance">Maintenance</span>      | Full implementation of existing features. May receive new features. Support for bug & security fixes only.                                                                   |
| <span id="support-legacy">Legacy</span>                | Legacy implementation. May have limited function, but no maintenance provided. [Contact our customer support team for special requests.](https://www.datadoghq.com/support/) |

### Supported Versions
<!-- NOTE: When updating the below section ensure you update the minimum supported version listed in the public docs here: https://docs.datadoghq.com/tracing/setup_overview/setup/go/?tab=containers#compatibility-requirements -->
| **Go Version** | **Support level**                   |
|----------------|-------------------------------------|
| 1.19           | [GA](#support-ga)                   |
| 1.18           | [GA](#support-ga)                   |
| 1.17           | [Maintenance](#support-maintenance) |
| 1.16           | [Legacy](#support-legacy)           |

* Datadog's Trace Agent >= 5.21.1


#### Package Versioning

A **Minor** version change will be released whenever a new version of Go is released. At that time the newest version of Go is added to [GA](#support-ga), the second oldest supported version moved to [Maintenance](#support-maintenance) and the oldest previously supported version dropped to [Legacy](#support-legacy).
**For example**:
For a dd-trace-go version 1.37.*

| Go Version | Support                             |
|------------|-------------------------------------|
| 1.18       | [GA](#support-ga)                   |
| 1.17       | [GA](#support-ga)                   |
| 1.16       | [Maintenance](#support-maintenance) |

Then after Go 1.19 is released there will be a new dd-trace-go version 1.38.0 with support:

| Go Version | Support                             |
|------------|-------------------------------------|
| 1.19       | [GA](#support-ga)                   |
| 1.18       | [GA](#support-ga)                   |
| 1.17       | [Maintenance](#support-maintenance) |
| 1.16       | [Legacy](#support-legacy)           |

### Contributing

Before considering contributions to the project, please take a moment to read our brief [contribution guidelines](CONTRIBUTING.md).

### Testing

Tests can be run locally using the Go toolset. The grpc.v12 integration will fail (and this is normal), because it covers for deprecated methods. In the CI environment
we vendor this version of the library inside the integration. Under normal circumstances this is not something that we want to do, because users using this integration
might be running versions different from the vendored one, creating hard to debug conflicts.

To run integration tests locally, you should set the `INTEGRATION` environment variable. The dependencies of the integration tests are best run via Docker. To get an
idea about the versions and the set-up take a look at our [CI config](./.circleci/config.yml).

The best way to run the entire test suite is using the [CircleCI CLI](https://circleci.com/docs/2.0/local-cli/). In order to run
jobs locally, you'll first need to convert the Circle CI configuration to a format accepted by the `circleci` cli tool:
  * `circleci config process .circleci/config.yml > process.yml` (from the repository root)

Once you have a converted `process.yml`, simply run `circleci local execute -c process.yml --job <job-name>`. 
Note that you might have to increase the resources dedicated to Docker to around 4GB.
