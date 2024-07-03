package user

import (
	"APIGateway/internal/app/middleware"
	user "APIGateway/internal/clients/user_service"
	permissionschecker "APIGateway/pkg/tools/permissions_checker"
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mailru/easyjson"
)

type Handler struct {
	UserService user.UserService
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPut, "/profile/edit", middleware.AuthMiddleware(h.FillUserProfile))
	router.HandlerFunc(http.MethodPatch, "/admin/profiles", middleware.AuthMiddleware(h.ChangeUserStatus))
	router.HandlerFunc(http.MethodGet, "/teacher/lessons/set-grades", middleware.AuthMiddleware(h.GetStudentsByClassname))
	router.HandlerFunc(http.MethodDelete, "/admin/profiles", middleware.AuthMiddleware(h.DeleteUser))
}

// FillUserProfile godoc
// @Summary Fill user profile
// @Description Fill user profile
// @Security TokenAuth
// @Accept  json
// @Produce  json
// @Param Request body user.FillUserProfileReq true "Request"
// @Success 200 "Success"
// @Failure 400 "Bad request"
// @Failure 401 "Unauthorized"
// @Failure 404 "User not found"
// @Failure 500 "Internal"
// @Router /profile/edit [put]
func (h *Handler) FillUserProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()
	req := &user.FillUserProfileReq{}
	if err := easyjson.UnmarshalFromReader(r.Body, req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	status, err := h.UserService.FillUserProfile(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	w.WriteHeader(status)
}

// ChangeUserStatus godoc
// @Summary Change user activity \\ ADMIN ONLY
// @Description Change user activity
// @Security TokenAuth
// @Accept  json
// @Produce  json
// @Param Request body user.ChangeUserStatusReq true "Request"
// @Success 200 "Success"
// @Failure 400 "Bad request"
// @Failure 401 "Unauthorized"
// @Failure 403 "Permission denied"
// @Failure 404 "User not found"
// @Failure 500 "Internal"
// @Router /admin/profiles [patch]
func (h *Handler) ChangeUserStatus(w http.ResponseWriter, r *http.Request) {
	permissionschecker.IsAdmin(w, r)
	w.Header().Set("Content-Type", "application/json")

	defer r.Body.Close()
	req := &user.ChangeUserStatusReq{}
	if err := easyjson.UnmarshalFromReader(r.Body, req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	status, err := h.UserService.ChangeUserStatus(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	w.WriteHeader(status)
}

// GetStudentsByClassname godoc
// @Summary Get students by classname
// @Description Get students by classname
// @Security TokenAuth
// @Accept  json
// @Produce  json
// @Param classname query string true "Request"
// @Success 200 "Success"
// @Failure 400 "Bad request"
// @Failure 401 "Unauthorized"
// @Failure 403 "Permission denied"
// @Failure 404 "Class not found"
// @Failure 500 "Internal"
// @Router /teacher/lessons/set-grades [get]
func (h *Handler) GetStudentsByClassname(w http.ResponseWriter, r *http.Request) {
	permissionschecker.IsTeacher(w, r)
	w.Header().Set("Content-Type", "application/json")

	students, status, err := h.UserService.GetStudentsByClassname(r.Context(), r.URL.Query().Get("classname"))
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	bytes, err := json.Marshal(students)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(bytes)

	w.WriteHeader(status)
}

// DeleteUser godoc
// @Summary Delete user \\ ADMIN ONLY
// @Description Delete user
// @Security TokenAuth
// @Accept  json
// @Produce  json
// @Param id query uint64 true "Request"
// @Success 200 "Success"
// @Failure 400 "Bad request"
// @Failure 401 "Unauthorized"
// @Failure 403 "Permission denied"
// @Failure 404 "User not found"
// @Failure 500 "Internal"
// @Router /admin/profiles [delete]
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	permissionschecker.IsAdmin(w, r)
	w.Header().Set("Content-Type", "application/json")
	status, err := h.UserService.DeleteUser(r.Context(), r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	w.WriteHeader(status)
}
