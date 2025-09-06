package endpoint

import "github.com/gorilla/mux"

func Routes(r *mux.Router, h endpointList) {

	r.Methods("GET").Path("/health").Handler(h.HealthCheckHandler)
}
