package lesson

import (
	"APIGateway/internal/app/middleware"
	"APIGateway/internal/clients/lesson_service"
	permissionschecker "APIGateway/pkg/tools/permissions_checker"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mailru/easyjson"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	Service lesson_service.LessonService
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPost, "/admin/lessons", middleware.AuthMiddleware(h.SetLesson))
}

// SetLesson godoc
// @Summary Set new lesson \\ ADMIN ONLY
// @Description Set new lesson
// @Security TokenAuth
// @Accept  json
// @Produce  json
// @Param Request body lesson_service.SetLessonReq true "Request"
// @Success 201 "Created"
// @Failure 400 "Bad request"
// @Failure 401 "Unauthorized"
// @Failure 403 "Permission denied"
// @Failure 409 "Lesson already exists"
// @Failure 500 "Internal"
// @Router /admin/lessons [post]
func (h *Handler) SetLesson(w http.ResponseWriter, r *http.Request) {
	permissionschecker.IsAdmin(w, r)
	w.Header().Set("Content-Type", "application/json")

	defer r.Body.Close()
	req := &lesson_service.SetLessonReq{}

	if err := easyjson.UnmarshalFromReader(r.Body, req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	status, err := h.Service.SetLesson(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), int(status))
		return
	}

	w.WriteHeader(int(status))
}

func (h *Handler) GetAllTeacherLessons(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	teacherID := r.Context().Value("user-id")

	lessons, err := h.Service.GetAllTeacherLessons(r.Context(), teacherID.(int64))
	if err != nil {
		if status.Code(err) == codes.NotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, lesson := range lessons {
		if r.URL.Query().Get("lesson") == lesson.LessonName && r.URL.Query().Get("classname") == lesson.ClassName {
			return
		}
	}
	http.Error(w, "permission denied", http.StatusForbidden)
}
