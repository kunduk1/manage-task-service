package team

import (
	"net/http"
	"time"

	"github.com/kunduk1/manage-task-service/internal/api/team/converter"
	"github.com/kunduk1/manage-task-service/internal/transport/response"
	teamv1 "github.com/kunduk1/manage-task-service/pkg/team/v1"
)

// defaultTopCreatorsWindow — окно по умолчанию, если from/to не заданы (последний месяц).
const defaultTopCreatorsWindow = 30 * 24 * time.Hour

// @Summary      Top task creators per team
// @Description  Returns the top-3 task creators in each team over the given period. Defaults to the last 30 days.
// @Tags         teams
// @Produce      json
// @Security     BearerAuth
// @Param        from  query     string  false  "period start, RFC3339 (default now-30d)"
// @Param        to    query     string  false  "period end, RFC3339 (default now)"
// @Success      200   {object}  teamv1.TopCreatorsResponse
// @Failure      400   {object}  response.ErrorBody
// @Failure      401   {object}  response.ErrorBody
// @Failure      500   {object}  response.ErrorBody
// @Router       /teams/top-creators [get]
func (h *Handler) TopCreators(w http.ResponseWriter, r *http.Request) {
	if _, ok := userID(r); !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	q := r.URL.Query()
	to := time.Now()
	from := to.Add(-defaultTopCreatorsWindow)

	if v := q.Get("from"); v != "" {
		parsed, err := time.Parse(time.RFC3339, v)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "invalid from (expected RFC3339)")
			return
		}
		from = parsed
	}
	if v := q.Get("to"); v != "" {
		parsed, err := time.Parse(time.RFC3339, v)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "invalid to (expected RFC3339)")
			return
		}
		to = parsed
	}
	if !from.Before(to) {
		response.Error(w, http.StatusBadRequest, "from must be before to")
		return
	}

	creators, err := h.svc.TopCreators(r.Context(), from, to)
	if err != nil {
		response.ServiceError(w, err)
		return
	}

	resp := teamv1.TopCreatorsResponse{
		From:     from,
		To:       to,
		Creators: converter.ToTopCreatorResponses(creators),
	}
	response.JSON(w, http.StatusOK, resp)
}
