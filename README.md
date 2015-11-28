khealth
=====

[![Docker Repository on Quay.io](https://quay.io/repository/coreos/khealth/status "Docker Repository on Quay.io")](https://quay.io/repository/coreos/khealth)

**khealth** is a kubernetes cluster monitoring suite. Routines are defined which exercise various kubernetes subsystems. Collectors receive status update events from Routines and compute current state.

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

A module is defined as a single entrypoint command which makes use of a set of routines and a single collector to expose a metric set via the collector's status endpoint.

- **cmd/**:

  This is where the module entrypoint programs live. Each module should have exactly one package in cmd/ which produces an executable which "enacts" that module.

- **pkg/routines**:

  Routines are defined via structs that implement the handler interface.
  ```go
  type RoutineHandler interface {
  	Init() error
	Poll() error
	Cleanup() error
  }
  Init() is called, and then Poll() in a loop. Once the ttl has expired, cleanup is called. Rinse-repeat. Every step generates an event which is received by the collector.
  ```
  The `NewRoutine` function can be used to create a new Routine from the handler.

- **pkg/collectors**:

  ```go
  type Collector interface {
  	Start() error
	Status(w http.ResponseWriter, r *http.Request)
	Terminate() error
  }
  ```
  Collectors should implement this interface and make use of Routines via `struct RoutineHandler` and `func NewRoutine`.

## Modules
- **cmd/rscheduler**
  This module uses a single routine which schedules/unschedules pause pods via a replication controller.

  The program exposes a single health endpoint which reports the state of the latest event.

## Roadmap
- More routines: We want routines that do everything! Test network latency. Write to disk. Computer fibonacci sequences.

- Prometheus integration: collectors expose prometheus-compatbile statyus endpoints and metrics. Pre-built infrastructure for aggregating metrics from a dynamic set of canary pods.

- Alerting: Use experimental [alertmanager](https://github.com/prometheus/alertmanager) to alert on metrics

## Who should use this?

- **Cluster administrators**: Gain insight into your kubernetes cluster's performance.  Monitor health endpoints which report on various testing routines.

- **Kubernetes developers**: A convenient way to "work-out" a cluster. Feel free to write modules that exist solely to torture-test a cluster and have no business running on the same cluster as production assests. And turn the replica count way up!

