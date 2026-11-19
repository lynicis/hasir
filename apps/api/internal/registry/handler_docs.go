package registry

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"

	"hasir/api/pkg/authentication"
)

type DocumentationHttpHandler struct {
	service    Service
	repository Repository
	jwtSecret  []byte
	sdkPath    string
}

func NewDocumentationHttpHandler(
	service Service,
	repository Repository,
	jwtSecret []byte,
	sdkPath string,
) *DocumentationHttpHandler {
	return &DocumentationHttpHandler{
		service:    service,
		repository: repository,
		jwtSecret:  jwtSecret,
		sdkPath:    sdkPath,
	}
}

func (h *DocumentationHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userId, err := h.authenticate(r)
	if err != nil {
		h.requireAuth(w)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/docs/")
	parts := strings.SplitN(path, "/", 3)

	if len(parts) < 3 {
		http.Error(w, "Invalid documentation path. Format: /docs/{orgId}/{repoId}/{commitHash}", http.StatusBadRequest)
		return
	}

	orgId := parts[0]
	repoId := parts[1]
	commitHash := parts[2]

	if !isValidPathComponent(orgId) || !isValidPathComponent(repoId) || !isValidPathComponent(commitHash) {
		http.Error(w, "Invalid path component", http.StatusBadRequest)
		return
	}

	docPath := filepath.Join(h.sdkPath, orgId, repoId, commitHash, "docs", "index.md")
	absDocPath, err := filepath.Abs(docPath)
	if err != nil {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	absSdkPath, err := filepath.Abs(h.sdkPath)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !strings.HasPrefix(absDocPath, absSdkPath+string(filepath.Separator)) {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	repo, err := h.repository.GetRepositoryById(r.Context(), repoId)
	if err != nil {
		zap.L().Error("Failed to get repository", zap.Error(err))
		http.Error(w, "Repository not found", http.StatusNotFound)
		return
	}

	if repo.OrganizationId != orgId {
		http.Error(w, "Repository not found", http.StatusNotFound)
		return
	}

	repoPath := repo.Path
	if repoPath == "" {
		repoPath = filepath.Join(DefaultReposPath, repoId)
	}
	hasAccess, err := h.service.ValidateSshAccess(r.Context(), userId, repoPath, SshOperationRead)
	if err != nil {
		zap.L().Error("Access validation failed", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !hasAccess {
		http.Error(w, "Permission denied", http.StatusForbidden)
		return
	}

	root, err := os.OpenRoot(absSdkPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "Documentation not found", http.StatusNotFound)
			return
		}
		zap.L().Error("Failed to open SDK root directory", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer func() {
		_ = root.Close()
	}()

	relDocPath := filepath.Join(orgId, repoId, commitHash, "docs", "index.md")
	docFile, err := root.Open(relDocPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "Documentation not found", http.StatusNotFound)
			return
		}
		zap.L().Error("Failed to open documentation file", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer func() {
		_ = docFile.Close()
	}()

	markdownContent, err := io.ReadAll(docFile)
	if err != nil {
		zap.L().Error("Failed to read documentation file", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	// #nosec G705 -- markdown content served as text/markdown, path is validated above
	_, _ = w.Write(markdownContent)
}

func (h *DocumentationHttpHandler) requireAuth(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Bearer realm="Documentation"`)
	http.Error(w, "Authentication required", http.StatusUnauthorized)
}

func (h *DocumentationHttpHandler) authenticate(r *http.Request) (string, error) {
	_, userID, err := authentication.ParseBearerHeader(r.Header.Get("Authorization"), h.jwtSecret)
	if err != nil {
		return "", err
	}
	return userID, nil
}
