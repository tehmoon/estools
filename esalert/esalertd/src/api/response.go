package api

import (
	"net/http"
	"github.com/gorilla/mux"
	"../response"
)

func HTTPGetResponseTag(rsm *response.Manager) (http.HandlerFunc) {
	return HTTPGetResponseTagV1(rsm)
}

func HTTPGetResponseTagV1(rsm *response.Manager) (http.HandlerFunc) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		tag := vars["tag"]

		payload, found := rsm.GetCache(tag)
		if ! found {
			w.WriteHeader(404)
			return
		}

		w.Header().Add("Content-Type", "aplication/json")
		w.Write(payload)
	}
}
