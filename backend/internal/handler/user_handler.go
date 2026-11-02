package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/forgeflow/usermanagement/internal/apperror"
	"github.com/forgeflow/usermanagement/internal/model"
	"github.com/forgeflow/usermanagement/internal/service"
)

type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler { return &UserHandler{svc: svc} }

// Register mounts the user routes on mux.
func (h *UserHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/users", h.create)
	mux.HandleFunc("GET /api/v1/users", h.list)
	mux.HandleFunc("GET /api/v1/users/stats", h.stats)
	mux.HandleFunc("GET /api/v1/users/{id}", h.get)
	mux.HandleFunc("PUT /api/v1/users/{id}", h.update)
	mux.HandleFunc("PATCH /api/v1/users/{id}/status", h.updateStatus)
	mux.HandleFunc("DELETE /api/v1/users/{id}", h.delete)
}

func decode(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		var mbe *http.MaxBytesError
		if errors.As(err, &mbe) {
			return err
		}
		return apperror.Validation("Request body is not valid JSON.", nil)
	}
	return nil
}

func (h *UserHandler) create(w http.ResponseWriter, r *http.Request) {
	var in model.UserInput
	if err := decode(r, &in); err != nil {
		writeError(w, err)
		return
	}
	u, err := h.svc.Create(r.Context(), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, http.StatusCreated, u, nil)
}

func (h *UserHandler) list(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	page, _ := strconv.Atoi(qs.Get("page"))
	size, _ := strconv.Atoi(qs.Get("pageSize"))
	users, meta, err := h.svc.List(r.Context(), model.ListQuery{
		Page: page, PageSize: size, Search: qs.Get("search"), Status: qs.Get("status"),
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, http.StatusOK, users, &meta)
}

func (h *UserHandler) stats(w http.ResponseWriter, r *http.Request) {
	s, err := h.svc.Stats(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, http.StatusOK, s, nil)
}

func (h *UserHandler) get(w http.ResponseWriter, r *http.Request) {
	u, err := h.svc.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, http.StatusOK, u, nil)
}

func (h *UserHandler) update(w http.ResponseWriter, r *http.Request) {
	var in model.UserInput
	if err := decode(r, &in); err != nil {
		writeError(w, err)
		return
	}
	u, err := h.svc.Update(r.Context(), r.PathValue("id"), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, http.StatusOK, u, nil)
}

func (h *UserHandler) updateStatus(w http.ResponseWriter, r *http.Request) {
	var in model.StatusInput
	if err := decode(r, &in); err != nil {
		writeError(w, err)
		return
	}
	u, err := h.svc.UpdateStatus(r.Context(), r.PathValue("id"), in.Status)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, http.StatusOK, u, nil)
}

func (h *UserHandler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Delete(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, http.StatusOK, map[string]bool{"deleted": true}, nil)
}
