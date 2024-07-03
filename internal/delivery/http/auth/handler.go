package auth

import (
	"APIGateway/internal/app/middleware"
	user "APIGateway/internal/clients/user_service"
	permissionschecker "APIGateway/pkg/tools/permissions_checker"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mailru/easyjson"
)

type Handler struct {
	AuthService user.AuthService
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPost, "/reg", h.RegisterUser)
	router.HandlerFunc(http.MethodPost, "/login", h.Login)
	router.HandlerFunc(http.MethodPatch, "/profile/change-password", middleware.AuthMiddleware(h.UpdatePassword))
	router.HandlerFunc(http.MethodPatch, "/admin/profiles/permissions", middleware.AuthMiddleware(h.SetPermissionLevel))
}

// RegisterUser godoc
// @Summary Register new user
// @Description Registgration
// @Accept  json
// @Produce  json
// @Param Request body user.RegisterReq true "Request"
// @Success 201 "Created"
// @Failure 400 "Bad request"
// @Failure 409 "User already exists"
// @Failure 500 "Internal"
// @Router /reg [post]
func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()
	req := &user.RegisterReq{}

	if err := easyjson.UnmarshalFromReader(r.Body, req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	status, err := h.AuthService.Register(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), status)
	}

	w.WriteHeader(status)
}

// Login godoc
// @Summary Login
// @Description Login
// @Accept  json
// @Produce  json
// @Param Request body user.LoginReq true "Request"
// @Success 200 "OK"
// @Failure 400 "Bad request"
// @Failure 401 "Invalid credentinals"
// @Failure 404 "User with this email not found"
// @Failure 500 "Internal"
// @Router /login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()
	req := &user.LoginReq{}

	if err := easyjson.UnmarshalFromReader(r.Body, req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	token, status, err := h.AuthService.Login(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), status)
	}

	bearer := "Bearer " + token

	w.Header().Add("Authorization", bearer)
	w.WriteHeader(status)
}

// UpdatePassword godoc
// @Summary Update user password
// @Description Update user password
// @Security TokenAuth
// @Accept  json
// @Produce  json
// @Param Request body user.UpdatePasswordReq true "Request"
// @Success 200 "OK"
// @Failure 400 "Bad request"
// @Failure 401 "Incorrect old password"
// @Failure 404 "User not found"
// @Failure 500 "Internal"
// @Router /profile/change-password [patch]
func (h *Handler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()
	req := &user.UpdatePasswordReq{}

	if err := easyjson.UnmarshalFromReader(r.Body, req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	status, err := h.AuthService.UpdatePassword(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), status)
	}

	w.WriteHeader(status)
}

// SetPermissionLevel godoc
// @Summary Set user's permission lvl \\ ADMIN ONLY
// @Description Set user's permission lvl
// @Security TokenAuth
// @Accept  json
// @Produce  json
// @Param Request body user.SetPermissionLevelReq true "Request"
// @Success 200 "Updated"
// @Failure 400 "Bad request"
// @Failure 403 "Permission denied"
// @Failure 404 "User not found"
// @Failure 500 "Internal"
// @Router /admin/profiles/permissions [patch]
func (h *Handler) SetPermissionLevel(w http.ResponseWriter, r *http.Request) {
	permissionschecker.IsAdmin(w, r)
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()
	req := &user.SetPermissionLevelReq{}

	if err := easyjson.UnmarshalFromReader(r.Body, req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	status, err := h.AuthService.SetPermissionLevel(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), status)
	}
	w.WriteHeader(status)
}

func (h *Handler) GetAdminPermissions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()

	lvl, err := h.AuthService.GetPermissionLevel(r.Context(), r.Header.Get("user-id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if lvl < 3 {
		http.Error(w, "permission denied", http.StatusForbidden)
	}

	w.WriteHeader(http.StatusOK)
}
