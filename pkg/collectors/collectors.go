package collectors

import (
    "net/http"
    "github.com/coreos/khealth/pkg/routines"
)

type Collector interface{
    Start()error
    Status(w http.ResponseWriter, r *http.Request)
    Terminate()error
}

func mergeEvents(last,incoming *routines.Event){
    if incoming.Status != 0 {
        last.Status = incoming.Status
    }
    last.Message = incoming.Message
}


