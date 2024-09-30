package rest

import (
	"encoding/json"
	"github.com/Benzogang-Tape/Reddit/internal/models"
	"net/http"
)

func newSession(w http.ResponseWriter, r *http.Request, statusCode int) {
	if r.Header.Get("Content-Type") != "application/json" {
		jsonSimpleErr(w, http.StatusBadRequest, models.NewSimpleErr(models.ErrUnknownPayload))
		return
	}
	payload, ok := r.Context().Value(models.Payload).(models.TokenPayload)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sess, err := models.NewSession(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp, err := json.Marshal(sess)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(statusCode)
	if _, err = w.Write(resp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
