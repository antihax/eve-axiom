package axiom

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/antihax/goesi/esi"
)

func (s *Axiom) killmailHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var v esi.GetKillmailsKillmailIdKillmailHashOk
	err := decoder.Decode(&v)
	if err != nil {
		log.Println(err)
		http.Error(w, `{"error":"400", "description":"invalid ESI killmail format"}`, http.StatusBadRequest)
		return
	}

	km, err := s.getAttributesFromKillmail(&v)
	if err != nil {
		log.Println(err)
		http.Error(w, `{"error":"500", "description":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	json, err := json.Marshal(km)
	if err != nil {
		log.Println(err)
		http.Error(w, `{"error":"500", "description":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	w.Write(json)
}
