package service

import (
	"encoding/json"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

func WriteObject(w http.ResponseWriter, o interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if o == nil {
		o = map[string]interface{}{
			"status": http.StatusText(http.StatusOK),
		}
	}
	e := json.NewEncoder(w)
	e.SetIndent("", "")
	err := e.Encode(o)
	if err != nil {
		zap.L().Warn("failed to serialize response", zap.Error(err))
	}
}

func WriteError(w http.ResponseWriter, code int, msg ...string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	e := json.NewEncoder(w)
	e.SetIndent("", "")
	errorText := http.StatusText(code)
	if len(msg) > 0 {
		errorText = strings.Join(msg, " ")
	}
	err := e.Encode(map[string]interface{}{
		"error": errorText,
	})
	if err != nil {
		zap.L().Warn("failed to serialize error response", zap.Error(err))
	}
}
