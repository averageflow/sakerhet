package dummyservices

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type SomeTestAPI interface {
	Start()
	Stop()
}

type dummyAPI struct {
	server      *http.Server
	stopChannel chan os.Signal
}

func timeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Unsupported HTTP method!", http.StatusUnprocessableEntity)
		return
	}

	tm := time.Now().Format(time.RFC1123)
	_, _ = w.Write([]byte("The time is: " + tm))
}

func customStatusResponseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Unsupported HTTP method!", http.StatusUnprocessableEntity)
		return
	}

	type expectedBodyData struct {
		WantedResponseCode int `json:"wantedResponseCode"`
	}

	var e expectedBodyData

	err := json.NewDecoder(r.Body).Decode(&e)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// return the status code that the client wanted
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(e.WantedResponseCode)
	_, _ = w.Write([]byte(fmt.Sprintf(`{"returnedResponseCode": %d}`, e.WantedResponseCode)))
}

func dummyHandlers() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/time", http.HandlerFunc(timeHandler))
	mux.Handle("/custom-status-response", http.HandlerFunc(customStatusResponseHandler))

	return mux
}

func (d *dummyAPI) Start() {
	d.server = &http.Server{Addr: ":8765", Handler: dummyHandlers()}
	go func() {
		_ = d.server.ListenAndServe()
	}()

	d.stopChannel = make(chan os.Signal, 1)
	signal.Notify(d.stopChannel, os.Interrupt)

	<-d.stopChannel
}

func (d *dummyAPI) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = d.server.Shutdown(ctx)
}

func NewSomeTestAPI() SomeTestAPI {
	return &dummyAPI{}
}
