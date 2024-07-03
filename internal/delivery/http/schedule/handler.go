package schedule

import (
	"APIGateway/internal/app/middleware"
	schedule "APIGateway/internal/clients/schedule_service"
	permissionschecker "APIGateway/pkg/tools/permissions_checker"
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mailru/easyjson"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	Service schedule.ScheduleService
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPatch, "/teacher/schedule/set-schedule/lessons", middleware.AuthMiddleware(h.SetHomework))
	router.HandlerFunc(http.MethodGet, "/schedule", middleware.AuthMiddleware(h.GetWeekScheduleByClass))
	router.HandlerFunc(http.MethodPost, "/teacher/schedule/set-schedule/import", middleware.AuthMiddleware(h.SetWeeklySchedule))

}

// SetHomework godoc
// @Summary Set homework
// @Description Set homework
// @Security TokenAuth
// @Accept  json
// @Produce  json
// @Param Request body schedule.SetHomeworkReq true "Request"
// @Success 200 "Homework updated"
// @Failure 400 "Bad request"
// @Failure 401 "Unauthorized"
// @Failure 403 "Permission denied"
// @Failure 404 "Lesson not found"
// @Failure 500 "Internal"
// @Router /teacher/schedule/set-schedule/lessons [patch]
func (h *Handler) SetHomework(w http.ResponseWriter, r *http.Request) {
	permissionschecker.IsTeacher(w, r)
	w.Header().Set("Content-Type", "application/json")

	defer r.Body.Close()
	req := &schedule.SetHomeworkReq{}
	if err := easyjson.UnmarshalFromReader(r.Body, req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	status, err := h.Service.SetHomework(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), int(status))
		return
	}

	w.WriteHeader(int(status))
}

// GetWeekScheduleByClass godoc
// @Summary Get weekly schedule by class
// @Description Get weekly schedule by class
// @Security TokenAuth
// @Accept  json
// @Produce  json
// @Param classname query string true "Request"
// @Success 200 "Success"
// @Failure 400 "Bad request"
// @Failure 401 "Unauthorized"
// @Failure 404 "Class not found"
// @Failure 500 "Internal"
// @Router /schedule [get]
func (h *Handler) GetWeekScheduleByClass(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	resp, err := h.Service.GetWeekScheduleByClass(r.Context(), r.URL.Query().Get("classname"))
	if err != nil {
		if code := status.Code(err); code == codes.NotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	}

	bytes, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(bytes)
	w.WriteHeader(http.StatusOK)
}

// GetWeekScheduleByClass godoc
// @Summary Get weekly schedule by class
// @Description Get weekly schedule by class
// @Security TokenAuth
// @Accept  mpfd
// @Produce  json
// @Param uploaded-schedule formData file true "Scheudle excel file"
// @Param sheet query string true "Sheet"
// @Success 200 "Success"
// @Failure 400 "Bad request"
// @Failure 401 "Unauthorized"
// @Failure 403 "Permission denied"
// @Failure 500 "Internal"
// @Router /teacher/schedule/set-schedule/import [post]
func (h *Handler) SetWeeklySchedule(w http.ResponseWriter, r *http.Request) {
	permissionschecker.IsTeacher(w, r)
	w.Header().Set("Content-Type", "application/json")

	r.ParseMultipartForm(32 << 20)

	defer r.Body.Close()
	f, head, err := r.FormFile("uploaded-schedule")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, f); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req := &schedule.SetWeeklyScheduleReq{
		Filename: head.Filename,
		FileData: buf.Bytes(),
		Sheet:    r.URL.Query().Get("sheet"),
	}

	status, err := h.Service.SetWeeklySchedule(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), int(status))
		return
	}

	w.WriteHeader(int(status))
}
