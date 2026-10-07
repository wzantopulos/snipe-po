package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// Start starts the HTTP server
func Start(port int) error {
	r := mux.NewRouter()

	// API routes
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/pos", listPOs).Methods("GET")
	api.HandleFunc("/pos", createPO).Methods("POST")
	api.HandleFunc("/pos/{id}", getPO).Methods("GET")
	api.HandleFunc("/pos/{id}/send", sendPO).Methods("POST")
	api.HandleFunc("/pos/{id}/approve", approvePO).Methods("POST")
	api.HandleFunc("/pos/{id}/reject", rejectPO).Methods("POST")
	api.HandleFunc("/pos/{id}/send-to-ap", sendToAP).Methods("POST")
	api.HandleFunc("/pos/{id}/mark-paid", markPaid).Methods("POST")
	api.HandleFunc("/pos/{id}/pdf", getPDF).Methods("GET")

	// Web UI routes
	r.HandleFunc("/", dashboardHandler)
	r.HandleFunc("/create", createHandler)
	r.HandleFunc("/view", viewHandler)

	// Apply form actions that redirect to view pages
	r.HandleFunc("/api/pos/{id}/send", func(w http.ResponseWriter, r *http.Request) {
		sendPO(w, r)
		redirectToView(w, r, mux.Vars(r)["id"])
	}).Methods("POST")

	r.HandleFunc("/api/pos/{id}/approve", func(w http.ResponseWriter, r *http.Request) {
		approvePO(w, r)
		redirectToView(w, r, mux.Vars(r)["id"])
	}).Methods("POST")

	r.HandleFunc("/api/pos/{id}/reject", func(w http.ResponseWriter, r *http.Request) {
		rejectPO(w, r)
		redirectToView(w, r, mux.Vars(r)["id"])
	}).Methods("POST")

	r.HandleFunc("/api/pos/{id}/send-to-ap", func(w http.ResponseWriter, r *http.Request) {
		sendToAP(w, r)
		redirectToView(w, r, mux.Vars(r)["id"])
	}).Methods("POST")

	r.HandleFunc("/api/pos/{id}/mark-paid", func(w http.ResponseWriter, r *http.Request) {
		markPaid(w, r)
		redirectToView(w, r, mux.Vars(r)["id"])
	}).Methods("POST")

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting snipe-po server on %s", addr)
	log.Printf("Web UI: http://localhost%s/", addr)
	log.Printf("API: http://localhost%s/api", addr)

	return http.ListenAndServe(addr, r)
}
