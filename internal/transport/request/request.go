package request

import (
	"encoding/json"
	"net/http"

	"github.com/kunduk1/manage-task-service/internal/transport/response"
)

// DecodeJSON читает JSON-тело запроса в v. При ошибке пишет 400 и возвращает false —
// тогда хендлеру достаточно сделать return.
func DecodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return false
	}
	return true
}
