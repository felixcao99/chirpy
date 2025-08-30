package main

import (
	"net/http"
)

func metricsReset(w http.ResponseWriter, r *http.Request) {
	if apiCfg.platform != "dev" {
		w.WriteHeader(403)
		w.Write([]byte("403 Forbidden"))
		return
	}
	err := apiCfg.dbQueries.ResetUsers(r.Context())
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("500 Internal Server Error"))
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	apiCfg.fileserverHits.Store(0)
}
