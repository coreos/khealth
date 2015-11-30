# khealth
=====

[![Docker Image on Quay.io](https://quay.io/repository/coreos/khealth/status "Docker Image on Quay.io")](https://quay.io/repository/coreos/khealth)

*khealth*  is a Kubernetes cluster monitoring suite. Its Routines exercise Kubernetes subsystems and send events to Collectors. Collectors collate these events to compute current cluster state. Cluster status is available from Collectors over a simple HTTP API, which is served on a cluster nodeport in the example below.

## Quick start

If you have a kubernetes cluster, you can deploy khealth.

```sh
cd khealth/
kubectl create -f ./contrib/k8s/khealth-ns.yaml
kubectl create -f ./contrib/k8s/khealth-{rc,service}.yaml
```

This will create a nodeport service which exposes the following status endpoints.

| Command  | NodePort |
| ------------- |:-------------:|
| rcscheduler  | 31337 |

## Architecture

A khealth *Module* is a single command that invokes a set of Routines and a single Collector. The Collector gathers events from the Routines and exposes metrics on its status endpoint.

* **cmd/**:

  This is where the `Module` entrypoint programs live. Each `Module` should have exactly one `main` package package in an eponymous directory beneath `cmd/`.

* **pkg/routines**:

  Routines are defined in structures that implement the `RoutineHandler` interface.
  ```go
  type RoutineHandler interface {
  	Init() error
	Poll() error
	Cleanup() error
  }
  ```
  `Init` is called, and then `Poll` in a loop. Once the ttl has expired, `Cleanup` is called. Rinse-repeat. Every step generates an event which is received by the collector.

  The `NewRoutine` function can be used to create a new Routine.

* **pkg/collectors**:

  ```go
  type Collector interface {
  	Start() error
	Status(w http.ResponseWriter, r *http.Request)
	Terminate() error
  }
  ```
  Collectors should implement the Collector interface and make use of routines. The pattern for wiring up routines inside a collector implementation is as follows:
  * `Start` : Call `Start` on all routines this collector uses. Then begin reading events from each routines' `Events` channel and collating current state.
  * `Status`: Serialize current state to http response.
  * `Terminate`: Call `SignalTerminate` on each routine.
      `SignalTerminate` non-blocking, so afterwards you'll want to block until each Routines' `Events` channel has emitted a `nil` value.

## Modules
- **cmd/rscheduler**
  This module uses a single routine which schedules/unschedules pause pods via a replication controller.

  The program exposes a single health endpoint which reports the state of the latest event.

## Roadmap
- More routines: We want routines that do everything! Test network latency. Write to disk. Compute fibonacci sequences.

* Prometheus integration: Collectors expose prometheus-compatbile status endpoints and metrics. Pre-built infrastructure for aggregating metrics from a dynamic set of canary pods purpose-built to exercise Kubernetes clusters in a specific way.

* Alerting: Use experimental [alertmanager](https://github.com/prometheus/alertmanager) to alert on metrics

## Who should use this?

* **Cluster administrators**: Gain insight into your kubernetes cluster's performance.  Monitor health endpoints which report on various testing routines.

* **Kubernetes developers**: A convenient way to "smoke test" a cluster. Feel free to write *Modules* that exist solely to torture test a cluster and have no business running on the same cluster as production assets. And turn the replica count way up!

