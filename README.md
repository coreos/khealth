# khealth

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

### Directory Layout

#### `cmd/`

This is where the `Module` entrypoint programs live. Each `Module` should have exactly one `main` package in an eponymous directory beneath `cmd/`.

#### `pkg/routines/`

Routines are defined in structures that implement the `RoutineHandler` interface.

```go
type RoutineHandler interface {
  Init() error
  Poll() error
  Cleanup() error
  }
```

`Init` is called, and then `Poll` in a loop. When the TTL expires, `Poll` terminates, and `Cleanup` is called. Each iteration of this cycle generates events, which are sent on the Routine's `Events` channel, usually to a Collector.

The `NewRoutine` function returns a pointer to a khealth `Routine` struct. It takes the following arguments:

* `client`: the Kubernetes API client
* `pollInterval`: how often (in seconds) `Poll` is called
* `podTTL`: how many seconds to loop on `Poll` before calling `Cleanup`
* `handler`: the `RoutineHandler` for this routine

#### `pkg/collectors/`

```go
type Collector interface {
  Start() error
  Status(w http.ResponseWriter, r *http.Request)
  Terminate() error
}
```

Collectors must implement the Collector interface and make use of Routines. To wire Routines to a Collector implementation, follow this general pattern:

* `Start` : Call `Start` on all routines this collector uses. Then begin reading events from each routines' `Events` channel and collating current state.
* `Status`: Serialize current state to HTTP response.
* `Terminate`: Call `SignalTerminate` on each routine. `SignalTerminate` is non-blocking, so before returning you'll want to block until each Routine's `Events` channel has emitted a `nil` value. That way, when `Terminate` returns you can be assured your Routines have all cleaned up.

## Included Modules

### `cmd/rcscheduler/`

This module uses a single routine which schedules/unschedules pause pods via a replication controller. The program exposes a single health endpoint which reports the state of the latest event.

## Roadmap

* More routines: We want routines that do everything! Test network latency. Write to disk. Compute fibonacci sequences.

* Prometheus integration: Collectors expose Prometheus-compatible status endpoints and metrics, providing readymade infrastructure to aggregate statistics from a set of canary pods, designed specifically to exercise Kubernetes cluster resources.

* Alerting: Use the experimental [alertmanager](https://github.com/prometheus/alertmanager) to alert on metrics

## Who should use this?

**Cluster administrators**: Gain insight into your Kubernetes cluster's performance.  Monitor health endpoints which report on various testing routines.

**Kubernetes developers**: A convenient way to "smoke test" a cluster. Feel free to write *Modules* that exist solely to torture test a cluster and have no business running on the same cluster as production assets. And turn the replica count way up!
