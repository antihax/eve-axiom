package axiom

import (
	"log"
	"net/http"
)

// Axiom provides API for parsing ship fits through dogma into attribute data
type Axiom struct {
}

// NewAxiom creates a new Axiom microservice
func NewAxiom() *Axiom {
	return &Axiom{}
}

// RunServer starts listening on port 3005 for API requests
func (s *Axiom) RunServer() {
	http.HandleFunc("/killmail", s.killmailHandler)
	log.Fatal(http.ListenAndServe(":3005", nil))
}
