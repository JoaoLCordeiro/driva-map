	package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/empresa/mercado/internal/models"
	"github.com/empresa/mercado/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	repo *repository.Repository
}

func New(repo *repository.Repository) *Handler {
	return &Handler{repo: repo}
}

// helpers

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, models.ErrorResponse{Error: msg})
}

func decode(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func ufFromPath(path, prefix string) string {
	return strings.ToUpper(strings.TrimPrefix(path, prefix))
}

func idFromPath(path, prefix string) (int, bool) {
	s := strings.TrimPrefix(path, prefix)
	id, err := strconv.Atoi(s)
	return id, err == nil
}

// ---- Auth ----

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	// checks method and json
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req models.LoginRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	// tests if all fields are filled and returns info if isn't
	fields := map[string]string{}
	if req.Email == "" {
		fields["email"] = "required"
	}
	if req.Password == "" {
		fields["password"] = "required"
	}
	if len(fields) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, models.ErrorResponse{Error: "validation error", Fields: fields})
		return
	}

	// gets user by email and, if it fails, returns the corresponding status code
	user, err := h.repo.GetUserByEmail(req.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if user == nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// gen the JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})
	signed, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not sign token")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": signed})
}

// ---- States ----

func (h *Handler) ListStates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	states, err := h.repo.ListStates()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, states)
}

func (h *Handler) GetState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	uf := ufFromPath(r.URL.Path, "/states/")
	if len(uf) != 2 {
		writeError(w, http.StatusBadRequest, "invalid uf")
		return
	}

	state, err := h.repo.GetStateByUF(uf)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if state == nil {
		writeError(w, http.StatusNotFound, "state not found")
		return
	}
	writeJSON(w, http.StatusOK, state)
}

// ---- Branches ----

func (h *Handler) ListBranches(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	q := r.URL.Query()
	uf := strings.ToUpper(q.Get("uf"))
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	result, err := h.repo.ListBranches(uf, page, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) CreateBranch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req models.BranchRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	// if the branch validation returns something wrong, stop the creation
	fields := validateBranch(req)
	if len(fields) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, models.ErrorResponse{Error: "validation error", Fields: fields})
		return
	}

	// uses the decoded request to crate the branch
	branch, err := h.repo.CreateBranch(&req)
	if err != nil {
		if err.Error() == "uf not found" {
			writeJSON(w, http.StatusUnprocessableEntity, models.ErrorResponse{
				Error:  "validation error",
				Fields: map[string]string{"uf": "not found"},
			})
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusCreated, branch)
}

func (h *Handler) UpdateBranch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	id, ok := idFromPath(r.URL.Path, "/branches/")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req models.BranchRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	// if the branch validation returns something wrong, stop the update
	fields := validateBranch(req)
	if len(fields) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, models.ErrorResponse{Error: "validation error", Fields: fields})
		return
	}

	// uses the decoded request to update the branch
	branch, err := h.repo.UpdateBranch(id, &req)
	if err != nil {
		if err.Error() == "uf not found" {
			writeJSON(w, http.StatusUnprocessableEntity, models.ErrorResponse{
				Error:  "validation error",
				Fields: map[string]string{"uf": "not found"},
			})
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if branch == nil {
		writeError(w, http.StatusNotFound, "branch not found")
		return
	}
	writeJSON(w, http.StatusOK, branch)
}

func (h *Handler) DeleteBranch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	id, ok := idFromPath(r.URL.Path, "/branches/")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	// only soft deletes, writting current timezoned ts on deleted_at
	found, err := h.repo.SoftDeleteBranch(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "branch not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func validateBranch(req models.BranchRequest) map[string]string {
	// validade all fields from a branch and return it
	fields := map[string]string{}
	if strings.TrimSpace(req.Nome) == "" {
		fields["nome"] = "required"
	}
	if strings.TrimSpace(req.Cidade) == "" {
		fields["cidade"] = "required"
	}
	if len(strings.TrimSpace(req.UF)) != 2 {
		fields["uf"] = "must be a 2-letter state code"
	}
	if req.OpenedAt == "" {
		fields["opened_at"] = "required"
	} else if _, err := time.Parse("2006-01-02", req.OpenedAt); err != nil {
		fields["opened_at"] = "must be YYYY-MM-DD"
	}
	return fields
}