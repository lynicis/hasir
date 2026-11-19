package registry

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"go.uber.org/zap"

	"hasir/api/internal/user"
)

type GitHttpHandler struct {
	service   Service
	userRepo  user.Repository
	reposPath string
}

func NewGitHttpHandler(service Service, userRepo user.Repository, reposPath string) *GitHttpHandler {
	return &GitHttpHandler{
		service:   service,
		userRepo:  userRepo,
		reposPath: reposPath,
	}
}

func (h *GitHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userId, err := h.authenticate(r)
	if err != nil {
		h.requireAuth(w)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/git/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 1 || parts[0] == "" {
		http.Error(w, "Repository not found", http.StatusNotFound)
		return
	}

	repoUUID := strings.TrimSuffix(parts[0], ".git")
	if !isValidPathComponent(repoUUID) {
		http.Error(w, "Invalid repository", http.StatusBadRequest)
		return
	}
	repoPath := h.reposPath + "/" + repoUUID
	subPath := ""
	if len(parts) > 1 {
		subPath = parts[1]
	}

	var operation SshOperation
	serviceName := r.URL.Query().Get("service")
	if serviceName == "git-receive-pack" || subPath == "git-receive-pack" {
		operation = SshOperationWrite
	} else {
		operation = SshOperationRead
	}

	hasAccess, err := h.service.ValidateSshAccess(r.Context(), userId, repoPath, operation)
	if err != nil {
		zap.L().Error("Access validation failed", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !hasAccess {
		http.Error(w, "Permission denied", http.StatusForbidden)
		return
	}

	switch {
	case subPath == "info/refs":
		h.handleInfoRefs(w, r, repoPath)
	case subPath == "git-upload-pack" && r.Method == http.MethodPost:
		h.handleUploadPack(w, r, repoPath)
	case subPath == "git-receive-pack" && r.Method == http.MethodPost:
		h.handleReceivePack(w, r, repoPath)
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

func (h *GitHttpHandler) requireAuth(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Git Repository"`)
	http.Error(w, "Authentication required", http.StatusUnauthorized)
}

func (h *GitHttpHandler) authenticate(r *http.Request) (string, error) {
	_, apiKey, ok := r.BasicAuth()
	if !ok || apiKey == "" {
		return "", fmt.Errorf("missing credentials")
	}

	userDTO, err := h.userRepo.GetUserByApiKey(r.Context(), apiKey)
	if err != nil {
		return "", err
	}

	return userDTO.Id, nil
}

func writeGitSmartHttpAdvertisement(w http.ResponseWriter, serviceName string) {
	w.Header().Set("Content-Type", fmt.Sprintf("application/x-%s-advertisement", serviceName))
	w.Header().Set("Cache-Control", "no-cache")

	pktLine := fmt.Sprintf("# service=%s\n", serviceName)
	// #nosec G705 -- serviceName validated against whitelist, content-type is not HTML
	_, _ = fmt.Fprintf(w, "%04x%s", len(pktLine)+4, pktLine)
	_, _ = fmt.Fprint(w, "0000")
}

func (h *GitHttpHandler) handleInfoRefs(w http.ResponseWriter, r *http.Request, repoPath string) {
	serviceName := r.URL.Query().Get("service")

	var gitCommand string
	switch serviceName {
	case "git-upload-pack":
		gitCommand = "git-upload-pack"
	case "git-receive-pack":
		gitCommand = "git-receive-pack"
	default:
		http.Error(w, "Invalid service", http.StatusBadRequest)
		return
	}

	writeGitSmartHttpAdvertisement(w, serviceName)

	var cmd *exec.Cmd
	switch gitCommand {
	case "git-upload-pack":
		// #nosec G204 G702 -- repoPath validated via isValidPathComponent + DB auth check
		cmd = exec.Command("git-upload-pack", "--stateless-rpc", "--advertise-refs", repoPath)
	case "git-receive-pack":
		// #nosec G204 G702 -- repoPath validated via isValidPathComponent + DB auth check
		cmd = exec.Command("git-receive-pack", "--stateless-rpc", "--advertise-refs", repoPath)
	default:
		http.Error(w, "Invalid service", http.StatusBadRequest)
		return
	}

	cmd.Stdout = w
	cmd.Stderr = w

	if err := cmd.Run(); err != nil {
		zap.L().Error("Failed to run git command", zap.String("service", serviceName), zap.Error(err))
	}
}

func (h *GitHttpHandler) handleUploadPack(w http.ResponseWriter, r *http.Request, repoPath string) {
	w.Header().Set("Content-Type", "application/x-git-upload-pack-result")
	w.Header().Set("Cache-Control", "no-cache")

	// #nosec G204 G702 -- repoPath validated via isValidPathComponent + DB auth check
	cmd := exec.Command("git-upload-pack", "--stateless-rpc", repoPath)
	cmd.Stdin = r.Body
	cmd.Stdout = w
	cmd.Stderr = w

	if err := cmd.Run(); err != nil {
		zap.L().Error("git-upload-pack failed", zap.Error(err))
	}
}

func (h *GitHttpHandler) handleReceivePack(w http.ResponseWriter, r *http.Request, repoPath string) {
	w.Header().Set("Content-Type", "application/x-git-receive-pack-result")
	w.Header().Set("Cache-Control", "no-cache")

	message := fmt.Sprintf("\x02%s", banner)
	pktLine := fmt.Sprintf("%04x%s", len(message)+4, message)
	_, _ = w.Write([]byte(pktLine))

	unlock := lockRepoRefUpdate(repoPath)
	// #nosec G204 G702 -- repoPath validated via isValidPathComponent + DB auth check
	cmd := exec.Command("git-receive-pack", "--stateless-rpc", repoPath)
	cmd.Stdin = r.Body
	cmd.Stdout = w
	cmd.Stderr = w

	err := cmd.Run()
	unlock()
	if err != nil {
		zap.L().Error("git-receive-pack failed", zap.Error(err))
		return
	}

	triggerSdkGenerationAfterPush(r.Context(), h.service, repoPath)
}
