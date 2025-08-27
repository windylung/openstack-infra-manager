package http

import (
    "encoding/json"
    "net/http"
)

func WriteJSON(w http.ResponseWriter, code int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    _ = json.NewEncoder(w).Encode(v)
}

func errString(err error) any {
    if err == nil {
        return nil
    }
    return err.Error()
}


