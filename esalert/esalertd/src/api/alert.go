package api

import (
	"net/http"
	"github.com/tehmoon/errors"
	"github.com/gorilla/mux"
	"../storage"
)

func HTTPGetAlertIdLog(storage *storage.Manager) (http.HandlerFunc) {
	return HTTPGetAlertIdLogV1(storage)
}

func HTTPGetAlertIdLogV1(storage *storage.Manager) (http.HandlerFunc) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		id := vars["id"]
		w.Header().Add("Content-Type", "application/text")
		err := storage.Get(id, w)
		if err != nil {
			err = errors.Wrapf(err, "Error finding logs for alert id %q", id)
			// TODO: don't return full error stack to client
			http.Error(w, err.Error(), 404)
			return
		}
	}
}
