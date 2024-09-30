package rest

import (
	"encoding/json"
	"github.com/Benzogang-Tape/Reddit/internal/models"
	"net/http"
)

func jsonSimpleErr(w http.ResponseWriter, statusCode int, errMsg models.SimpleErr) {
	resp, err := json.Marshal(errMsg)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(statusCode)
	if _, err = w.Write(resp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func jsonComplexErr(w http.ResponseWriter, statusCode int, errMsg models.ComplexErrArr) {
	resp, err := json.Marshal(errMsg)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(statusCode)
	if _, err = w.Write(resp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
