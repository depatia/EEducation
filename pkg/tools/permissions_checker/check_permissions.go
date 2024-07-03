package permissionschecker

import (
	"net/http"
)

func IsAdmin(w http.ResponseWriter, r *http.Request) {
	if lvl := r.Context().Value("permission_lvl").(int64); lvl < 3 {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func IsTeacher(w http.ResponseWriter, r *http.Request) {
	if lvl := r.Context().Value("permission_lvl").(int64); lvl < 2 {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}
