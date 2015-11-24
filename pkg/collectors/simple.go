package collectors

import (
	"fmt"
	"github.com/coreos/khealth/pkg/routines"
	"net/http"
)

type SimpleCollector struct {
	routine   *routines.Routine
	lastEvent routines.Event
	termChan  chan interface{}
}

func NewSimpleCollector(routine *routines.Routine) *SimpleCollector {
	return &SimpleCollector{
		routine: routine,
		lastEvent: routines.Event{
			Status:  http.StatusServiceUnavailable,
			Message: "routine is not polling yet",
		},
		termChan: make(chan interface{}),
	}
}

func (sc *SimpleCollector) Start() error {

	if err := sc.routine.Start(); err != nil {
		return err
	}

	go func() {
		for {
			event := <-sc.routine.Events
			if event == nil {
				sc.termChan <- nil
				break
			} else {
				mergeEvents(&sc.lastEvent, event)
			}
		}
	}()

	return nil
}

func (sc *SimpleCollector) Status(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(sc.lastEvent.Status)
	w.Write([]byte(fmt.Sprintf("%s\n", sc.lastEvent.Message)))
}

func (sc *SimpleCollector) Terminate() error {
	sc.routine.SignalTerminate()
	<-sc.termChan
	if sc.lastEvent.Status == 200 {
		return nil
	} else {
		return fmt.Errorf(
			"error terminating collector: %s (status=%d)",
			sc.lastEvent.Message,
			sc.lastEvent.Status,
		)
	}
}
