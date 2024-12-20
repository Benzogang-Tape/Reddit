package rest

import (
	"encoding/json"
	"net/http"

	"github.com/Benzogang-Tape/Reddit/internal/models/errs"
	"github.com/Benzogang-Tape/Reddit/internal/models/httpresp"
)

const DefaultContentType = "application/json; charset=utf-8"

func sendResponse(data any, w http.ResponseWriter, opts ...httpresp.OptionFunc) {
	response := &httpresp.Response{
		Data:       data,
		StatusCode: http.StatusOK,
	}

	for _, opt := range opts {
		opt(response)
	}

	send(response, w)
}

func sendErrorResponse(w http.ResponseWriter, statusCode int, errMsg errs.RespError) {
	resp, err := errMsg.Marshal()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", DefaultContentType)
	w.WriteHeader(statusCode)
	if _, err = w.Write(resp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func send(r *httpresp.Response, w http.ResponseWriter) {
	resp, err := json.Marshal(r.Data)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, errs.NewSimpleErr(errs.ErrResponseError.Error()))
		return
	}

	if r.ContentType != "" {
		w.Header().Set("Content-Type", r.ContentType)
	} else {
		w.Header().Set("Content-Type", DefaultContentType)
	}

	w.WriteHeader(r.StatusCode)
	if _, err := w.Write(resp); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, errs.NewSimpleErr(errs.ErrResponseError.Error()))
	}
}
