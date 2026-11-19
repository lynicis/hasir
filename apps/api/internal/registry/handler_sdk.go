package registry

import (
	"errors"
	"fmt"
	"html"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gliderlabs/ssh"
	"go.uber.org/zap"
)

type SdkHttpHandler struct {
	sdkReposPath string
}

func NewSdkHttpHandler(sdkReposPath string) *SdkHttpHandler {
	return &SdkHttpHandler{
		sdkReposPath: sdkReposPath,
	}
}

func isValidPathComponent(component string) bool {
	if strings.Contains(component, "..") {
		return false
	}

	if filepath.IsAbs(component) {
		return false
	}

	if strings.ContainsAny(component, "/\\") {
		return false
	}

	cleaned := filepath.Clean(component)
	if cleaned != component || cleaned == "." {
		return false
	}

	return true
}

func parseSdkPath(path string) (orgId, repoId, commitHash, sdkType, subPath string, ok bool) {
	var segments []string
	for _, p := range strings.Split(path, "/") {
		if p != "" {
			segments = append(segments, p)
		}
	}

	if len(segments) < 3 {
		return "", "", "", "", "", false
	}

	sdkTypes := map[string]bool{
		"go-protobuf":    true,
		"go-connectrpc":  true,
		"go-grpc":        true,
		"js-bufbuild-es": true,
		"js-protobuf":    true,
		"js-connectrpc":  true,
	}

	sdkTypeSegmentIndex := -1
	for i, seg := range segments {
		trimmed := strings.TrimSuffix(seg, ".git")
		if sdkTypes[trimmed] {
			sdkTypeSegmentIndex = i
			break
		}
	}

	if sdkTypeSegmentIndex == -1 {
		// Fallback to old format: orgId/repoId/sdkType
		orgId = segments[0]
		repoId = strings.TrimSuffix(segments[1], ".git")
		sdkType = strings.TrimSuffix(segments[2], ".git")
		subPath = strings.Join(segments[3:], "/")
		return orgId, repoId, "", sdkType, subPath, true
	}

	if sdkTypeSegmentIndex < 2 {
		return "", "", "", "", "", false
	}

	orgId = segments[0]
	repoId = strings.TrimSuffix(segments[1], ".git")

	if sdkTypeSegmentIndex == 3 {
		commitHash = segments[2]
	}

	sdkType = strings.TrimSuffix(segments[sdkTypeSegmentIndex], ".git")
	subPath = strings.Join(segments[sdkTypeSegmentIndex+1:], "/")
	return orgId, repoId, commitHash, sdkType, subPath, true
}

func resolveValidatedSdkRepoPath(sdkReposPath, orgId, repoId, commitHash, sdkType string) (string, int, error) {
	if !isValidPathComponent(orgId) || !isValidPathComponent(repoId) || !isValidPathComponent(sdkType) || (commitHash != "" && !isValidPathComponent(commitHash)) {
		return "", http.StatusBadRequest, errors.New("invalid path component")
	}

	var repoPath string
	if commitHash != "" {
		repoPath = filepath.Join(sdkReposPath, orgId, repoId, commitHash, sdkType)
	} else {
		repoPath = filepath.Join(sdkReposPath, orgId, repoId, sdkType)
	}

	absRepoPath, err := filepath.Abs(repoPath)
	if err != nil {
		return "", http.StatusBadRequest, errors.New("invalid path")
	}
	absSdkReposPath, err := filepath.Abs(sdkReposPath)
	if err != nil {
		return "", http.StatusInternalServerError, errors.New("internal server error")
	}
	if !strings.HasPrefix(absRepoPath, absSdkReposPath+string(filepath.Separator)) {
		return "", http.StatusBadRequest, errors.New("invalid path")
	}

	return absRepoPath, http.StatusOK, nil
}

func (h *SdkHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/sdk/")

	orgId, repoId, commitHash, sdkType, subPath, ok := parseSdkPath(path)
	if !ok {
		http.Error(w, "Invalid SDK path. Format: /sdk/{orgId}/{repoId}/{sdkType}/ or /sdk/{orgId}/{repoId}/{commitHash}/{sdkType}/", http.StatusNotFound)
		return
	}

	absRepoPath, statusCode, err := resolveValidatedSdkRepoPath(h.sdkReposPath, orgId, repoId, commitHash, sdkType)
	if err != nil {
		http.Error(w, err.Error(), statusCode)
		return
	}

	// #nosec G304 G703 -- absRepoPath validated via isValidPathComponent + path prefix check above
	if _, err := os.Stat(filepath.Join(absRepoPath, ".git")); os.IsNotExist(err) { // lgtm [go/path-injection]
		http.Error(w, "SDK repository not found", http.StatusNotFound)
		return
	}

	// Handle go get request
	if r.URL.Query().Get("go-get") == "1" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		importHost := r.Host
		if h, _, err := net.SplitHostPort(r.Host); err == nil {
			importHost = h
		}
		importPath := importHost + r.URL.Path
		importPath = strings.TrimSuffix(importPath, "/")

		scheme := "http"
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		repoUrl := fmt.Sprintf("%s://%s%s", scheme, r.Host, strings.TrimSuffix(r.URL.Path, "/"))

		_, _ = fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
<meta name="go-import" content="%s git %s">
</head>
<body>
go get %s
</body>
</html>
`, html.EscapeString(importPath), html.EscapeString(repoUrl), html.EscapeString(importPath))
		return
	}

	serviceName := r.URL.Query().Get("service")
	if serviceName == "git-receive-pack" || subPath == "git-receive-pack" {
		http.Error(w, "SDK repositories are read-only", http.StatusForbidden)
		return
	}

	switch {
	case subPath == "info/refs":
		h.handleInfoRefs(w, r, absRepoPath)
	case subPath == "git-upload-pack" && r.Method == http.MethodPost:
		h.handleUploadPack(w, r, absRepoPath)
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

func (h *SdkHttpHandler) handleInfoRefs(w http.ResponseWriter, r *http.Request, repoPath string) {
	serviceName := r.URL.Query().Get("service")
	if serviceName != "git-upload-pack" {
		http.Error(w, "Only git-upload-pack is supported for SDK repos", http.StatusBadRequest)
		return
	}

	writeGitSmartHttpAdvertisement(w, serviceName)

	// #nosec G204 G702 -- repoPath validated via isValidPathComponent + path prefix check
	cmd := exec.Command("git-upload-pack", "--stateless-rpc", "--advertise-refs", repoPath)
	cmd.Stdout = w
	cmd.Stderr = w

	if err := cmd.Run(); err != nil {
		zap.L().Error("Failed to run git command for SDK", zap.String("service", serviceName), zap.Error(err))
	}
}

func (h *SdkHttpHandler) handleUploadPack(w http.ResponseWriter, r *http.Request, repoPath string) {
	w.Header().Set("Content-Type", "application/x-git-upload-pack-result")
	w.Header().Set("Cache-Control", "no-cache")

	// #nosec G204 G702 -- repoPath validated via isValidPathComponent + path prefix check
	cmd := exec.Command("git-upload-pack", "--stateless-rpc", repoPath)
	cmd.Stdin = r.Body
	cmd.Stdout = w
	cmd.Stderr = w

	if err := cmd.Run(); err != nil {
		zap.L().Error("git-upload-pack failed for SDK", zap.Error(err))
	}
}

type SdkSshHandler struct {
	sdkReposPath string
}

func NewSdkSshHandler(sdkReposPath string) *SdkSshHandler {
	return &SdkSshHandler{
		sdkReposPath: sdkReposPath,
	}
}

func (h *SdkSshHandler) HandleSession(session ssh.Session, userId string) error {
	cmd := session.RawCommand()
	if cmd == "" {
		_, _ = fmt.Fprint(session, banner)
		_, _ = fmt.Fprintf(session, "\n✅ SSH connection successful!\n")
		_, _ = fmt.Fprintf(session, "🔐 Authentication completed.\n\n")
		return nil
	}

	parts := strings.SplitN(cmd, " ", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid command format: %s", cmd)
	}

	gitCmd := parts[0]
	repoPath := strings.Trim(parts[1], "'\"")
	repoPath = strings.TrimPrefix(repoPath, "/")

	if gitCmd != "git-upload-pack" {
		if gitCmd == "git-receive-pack" {
			return fmt.Errorf("SDK repositories are read-only")
		}
		return fmt.Errorf("unsupported git command: %s", gitCmd)
	}

	if !strings.HasPrefix(repoPath, "sdk/") {
		return fmt.Errorf("invalid SDK path format, must start with 'sdk/'")
	}

	path := strings.TrimPrefix(repoPath, "sdk/")
	orgId, repoId, commitHash, sdkType, _, ok := parseSdkPath(path)
	if !ok {
		return fmt.Errorf("invalid SDK path format: sdk/{orgId}/{repoId}/{sdkType} or sdk/{orgId}/{repoId}/{commitHash}/{sdkType}")
	}

	absRepoPath, _, err := resolveValidatedSdkRepoPath(h.sdkReposPath, orgId, repoId, commitHash, sdkType)
	if err != nil {
		return err
	}

	zap.L().Info("Executing Git command for SDK",
		zap.String("userId", userId),
		zap.String("command", gitCmd),
		zap.String("repoPath", absRepoPath))

	if _, err := os.Stat(absRepoPath); os.IsNotExist(err) {
		zap.L().Error("SDK repository path does not exist", zap.String("path", absRepoPath))
		return fmt.Errorf("SDK repository not found: %s", absRepoPath)
	}

	if _, err := os.Stat(filepath.Join(absRepoPath, ".git")); os.IsNotExist(err) {
		zap.L().Error("Not a valid git repository", zap.String("path", absRepoPath))
		return fmt.Errorf("not a git repository: %s", absRepoPath)
	}

	// #nosec G204 -- command is hardcoded, path is validated and sanitized
	execCmd := exec.Command("git-upload-pack", absRepoPath)
	execCmd.Dir = filepath.Dir(absRepoPath)
	execCmd.Stdin = session
	execCmd.Stdout = session
	execCmd.Stderr = session.Stderr()

	zap.L().Debug("Starting git command for SDK",
		zap.String("command", gitCmd),
		zap.String("repoPath", absRepoPath))

	if err := execCmd.Start(); err != nil {
		zap.L().Error("Failed to start git command", zap.Error(err))
		return fmt.Errorf("failed to start %s: %w", gitCmd, err)
	}

	return execCmd.Wait()
}
