package main

import (
	"encoding/json"
	"net/http"
	"sync"

	api "server/api/v1"

	"github.com/gorilla/mux"
)

type RecordWrapper struct {
	Record *api.Record `json:"record"`
}

type OffsetWrapper struct {
	Offset uint64 `json:"offset"`
}

type Log struct {
	mu      sync.Mutex
	records []api.Record
}

var log Log

func addLog(w http.ResponseWriter, r *http.Request) {
	log.mu.Lock()
	defer log.mu.Unlock()
	var recordWrapper RecordWrapper
	err := json.NewDecoder(r.Body).Decode(&recordWrapper)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	record := recordWrapper.Record

	if record.Value == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Missing 'value' field in record"}`))
		return
	}

	record.Offset = uint64(len(log.records))
	log.records = append(log.records, *record)
	var offsetWrapper OffsetWrapper
	offsetWrapper.Offset = record.Offset

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(offsetWrapper)
}

func getLog(w http.ResponseWriter, r *http.Request) {
	log.mu.Lock()
	defer log.mu.Unlock()

	if len(log.records) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Empty log"}`))
		return
	}

	var offsetWrapper OffsetWrapper
	err := json.NewDecoder(r.Body).Decode(&offsetWrapper)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if uint64(len(log.records)) <= offsetWrapper.Offset {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Record not found"}`))
		return
	}

	var recordWrapper RecordWrapper
	recordWrapper.Record = &log.records[offsetWrapper.Offset]

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(recordWrapper)
}

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/", addLog).Methods("POST")
	router.HandleFunc("/", getLog).Methods("GET")
	http.ListenAndServe(":8080", router)
}
