package webhook

import (
	"encoding/json"
	"log"
	"net/http"
)

func WebHookHandler(w http.ResponseWriter, req *http.Request) {
	var hook = new(RegistyHook)
	defer req.Body.Close()
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if err := json.NewDecoder(req.Body).Decode(&hook); err != nil {
		log.Printf("Post decode error: %+v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	repr, err := json.MarshalIndent(hook, "", "  ")
	if err != nil {
		log.Printf("Post marshal error: %+v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("Recieved webhook from %s:\n%s", req.RemoteAddr, string(repr))
	w.WriteHeader(http.StatusOK)
}
