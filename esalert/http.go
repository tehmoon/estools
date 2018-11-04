package main

import (
	"io"
	"log"
	"github.com/tehmoon/errors"
	"net/http"
	"github.com/gorilla/mux"
	"net"
	"bytes"
	"time"
	"compress/gzip"
)

func startHTTP(listen string, alertIndex map[string]*Alert) (error) {
	r := mux.NewRouter()

	r.HandleFunc("/alert/{id}", HTTPGetAlertId(alertIndex)).
		Methods("GET")

	addr, err := net.ResolveTCPAddr("tcp", listen)
	if err != nil {
		return errors.Wrapf(err, "Unable to resolve listen address")
	}

	l, err := net.ListenTCP(addr.Network(), addr)
	if err != nil {
		return errors.Wrapf(err, "Error listening on address: %s", addr)
	}

	server := &http.Server{
		WriteTimeout: 5 * time.Second,
		ReadTimeout: 5 * time.Second,
		Handler: r,
	}

	server.Addr = l.Addr().String()

	go func(server *http.Server, l net.Listener) {
		server.Serve(l)
	}(server, l)

	log.Println("HTTP Server started")

	return nil
}

func HTTPGetAlertId(alertIndex map[string]*Alert) (func (w http.ResponseWriter, r *http.Request)) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		if alert, ok := alertIndex[vars["id"]]; ok {
			if len(alert.Log) == 0 {
				w.Write([]byte("no data"))
				return
			}

			reader := bytes.NewReader(alert.Log)

			zreader, err := gzip.NewReader(reader)
			if err != nil {
				log.Println(errors.Wrapf(err, "Error writing response for alert id: %s", alert.Id).Error())
				w.WriteHeader(500)
				return
			}
			defer zreader.Close()

			w.Header().Add("Content-Type", "text/plain")
			io.Copy(w, zreader)
			return
		}

		w.WriteHeader(404)
	}
}
