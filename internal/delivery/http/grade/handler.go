package grade

import (
	"APIGateway/internal/app/middleware"
	clienterrors "APIGateway/internal/client_errors"
	grade "APIGateway/internal/clients/grade_service"
	permissionschecker "APIGateway/pkg/tools/permissions_checker"

	"encoding/json"
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mailru/easyjson"
)

type Handler struct {
	Service grade.GradeService
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPost, "/teacher/lessons/grades", middleware.AuthMiddleware(h.SetGrade))
	router.HandlerFunc(http.MethodGet, "/lessons/grades", middleware.AuthMiddleware(h.GetLessonGrades))
	router.HandlerFunc(http.MethodGet, "/lessons/grades/all", middleware.AuthMiddleware(h.GetAllLessonsGradesByStudentID))
	router.HandlerFunc(http.MethodDelete, "/teacher/lessons/grades", middleware.AuthMiddleware(h.DeleteGrade))
	router.HandlerFunc(http.MethodPatch, "/teacher/lessons/grades", middleware.AuthMiddleware(h.ChangeGrade))
	router.HandlerFunc(http.MethodPost, "/teacher/lessons/term-grades", middleware.AuthMiddleware(h.SetTermGrade))
}

// SetGrade godoc
// @Summary Set grade
// @Description Set grade by student id && lesson
// @Security TokenAuth
// @Accept  json
// @Produce  json
// @Param Request body grade.SetGradeReq true "Request"
// @Param Request body grade.SetGradeReq true "Request"
// @Success 201 "Grade created"
// @Failure 400 "Bad request"
// @Failure 401 "Unauthorized"
// @Failure 403 "Permission denied"
// @Failure 409 "Grade already exists"
// @Failure 500 "Internal"
// @Router /teacher/lessons/grades [post]
func (h *Handler) SetGrade(w http.ResponseWriter, r *http.Request) {
	permissionschecker.IsTeacher(w, r)
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()
	req := &grade.SetGradeReq{}

	if err := easyjson.UnmarshalFromReader(r.Body, req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	status, err := h.Service.SetGrade(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), int(status))
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GetAllLessonsGradesByStudentID godoc
// @Summary Get all student's lesson grades
// @Description Get all student lesson grades by student id
// @Security TokenAuth
// @Accept  json
// @Produce  json
// @Param q    query     string  false  "method could call without any parameters"
// @Success 200 {array} grade.Grade
// @Failure 400 "Bad request"
// @Failure 401 "Unauthorized"
// @Failure 500 "Internal"
// @Router /lessons/grades/all [get]
func (h *Handler) GetAllLessonsGradesByStudentID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	resp, err := h.Service.GetAllLessonsGradesByStudentID(r.Context(), r.Context().Value("user_id").(int64))
	if err != nil {
		if errors.Is(err, clienterrors.ErrAllFieldsRequired) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bytes, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(bytes)
	w.WriteHeader(http.StatusOK)
}

// GetLessonGrades godoc
// @Summary Get student's lesson grades
// @Description Get all student lesson grades by student id
// @Security TokenAuth
// @Accept  json
// @Produce  json
// @Param lesson query string true "Request"
// @Success 200 {array} grade.Grade
// @Failure 400 "Bad request"
// @Failure 401 "Unauthorized"
// @Failure 404 "Not found"
// @Failure 500 "Internal"
// @Router /lessons/grades [get]
func (h *Handler) GetLessonGrades(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp, err := h.Service.GetLessonGrades(r.Context(), r.URL.Query().Get("lesson"), r.Context().Value("user_id").(int64))
	if err != nil {
		if errors.Is(err, clienterrors.ErrAllFieldsRequired) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bytes, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(bytes)
	w.WriteHeader(http.StatusOK)
}

// DeleteGrade godoc
// @Summary Delete grade
// @Description Delete grade
// @Security TokenAuth
// @Accept  json
// @Produce  json
// @Param Request body grade.DeleteGradeReq true "Request"
// @Success 200 "Grade deleted"
// @Failure 400 "Bad request"
// @Failure 401 "Unauthorized"
// @Failure 403 "Permission denied"
// @Failure 404 "Grade not found"
// @Failure 500 "Internal"
// @Router /teacher/lessons/grades [delete]
func (h *Handler) DeleteGrade(w http.ResponseWriter, r *http.Request) {
	permissionschecker.IsTeacher(w, r)
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()
	req := &grade.DeleteGradeReq{}

	if err := easyjson.UnmarshalFromReader(r.Body, req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	status, err := h.Service.DeleteGrade(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), int(status))
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ChangeGrade godoc
// @Summary Change grade
// @Description Change grade
// @Security TokenAuth
// @Accept  json
// @Produce  json
// @Param Request body grade.ChangeGradeReq true "Request"
// @Success 200 "Grade changed"
// @Failure 400 "Bad request"
// @Failure 401 "Unauthorized"
// @Failure 403 "Permission denied"
// @Failure 404 "Grade not found"
// @Failure 500 "Internal"
// @Router /teacher/lessons/grades [patch]
func (h *Handler) ChangeGrade(w http.ResponseWriter, r *http.Request) {
	permissionschecker.IsTeacher(w, r)
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()
	req := &grade.ChangeGradeReq{}

	if err := easyjson.UnmarshalFromReader(r.Body, req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	status, err := h.Service.ChangeGrade(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), int(status))
		return
	}

	w.WriteHeader(http.StatusOK)
}

// SetTermGrade godoc
// @Summary Set term grade
// @Description Set term grade
// @Security TokenAuth
// @Accept  json
// @Produce  json
// @Param Request body grade.SetTermGradeReq true "Request"
// @Success 201 "Grade created"
// @Failure 400 "Bad request"
// @Failure 401 "Unauthorized"
// @Failure 403 "Permission denied"
// @Failure 406 "Not enough grades"
// @Failure 409 "Grade already exists"
// @Failure 500 "Internal"
// @Router /teacher/lessons/term-grades [post]
func (h *Handler) SetTermGrade(w http.ResponseWriter, r *http.Request) {
	permissionschecker.IsTeacher(w, r)
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()
	req := &grade.SetTermGradeReq{}

	if err := easyjson.UnmarshalFromReader(r.Body, req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	status, err := h.Service.SetTermGrade(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), int(status))
		return
	}

	w.WriteHeader(http.StatusCreated)
}
