package registry

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gliderlabs/ssh"
	"go.uber.org/zap"
	gossh "golang.org/x/crypto/ssh"

	"hasir/api/internal/user"
	"hasir/api/pkg/config"
)

type SshSessionHandler interface {
	HandleSession(session ssh.Session, userId string) error
}

func NewSshServer(
	cfg *config.Config,
	userRepo user.Repository,
	gitSshHandler SshSessionHandler,
	sdkSshHandler SshSessionHandler,
) (*ssh.Server, error) {
	hostKey, err := LoadOrGenerateHostKey(cfg.Ssh.HostKeyPath)
	if err != nil {
		return nil, err
	}

	sshServer := &ssh.Server{
		Addr: ":" + cfg.Ssh.Port,
		PublicKeyHandler: func(ctx ssh.Context, key ssh.PublicKey) bool {
			publicKeyStr := strings.TrimSpace(string(gossh.MarshalAuthorizedKey(key)))
			userDTO, err := userRepo.GetUserBySshPublicKey(context.Background(), publicKeyStr)
			if err != nil {
				zap.L().Debug("SSH auth failed", zap.Error(err))
				return false
			}
			ctx.SetValue("userId", userDTO.Id)
			zap.L().Info("SSH auth success", zap.String("userId", userDTO.Id))
			return true
		},
		Handler: func(session ssh.Session) {
			handleSshSession(session, gitSshHandler, sdkSshHandler)
		},
	}
	sshServer.AddHostKey(hostKey)

	return sshServer, nil
}

func StartSshServer(
	cfg *config.Config,
	userRepo user.Repository,
	gitSshHandler SshSessionHandler,
	sdkSshHandler SshSessionHandler,
) *ssh.Server {
	sshServer, err := NewSshServer(cfg, userRepo, gitSshHandler, sdkSshHandler)
	if err != nil {
		zap.L().Fatal("failed to load SSH host key", zap.Error(err))
	}

	go func() {
		zap.L().Info("SSH server starting", zap.String("port", cfg.Ssh.Port))
		if err := sshServer.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			zap.L().Fatal("SSH server error", zap.Error(err))
		}
	}()

	return sshServer
}

func handleSshSession(session ssh.Session, gitSshHandler, sdkSshHandler SshSessionHandler) {
	userId, ok := session.Context().Value("userId").(string)
	if !ok || userId == "" {
		_, _ = fmt.Fprintln(session.Stderr(), "Authentication required")
		_ = session.Exit(1)
		return
	}

	target := selectSshHandler(session.RawCommand(), gitSshHandler, sdkSshHandler)
	if handlerErr := target.HandleSession(session, userId); handlerErr != nil {
		_, _ = fmt.Fprintln(session.Stderr(), handlerErr.Error())
		_ = session.Exit(1)
		return
	}

	_ = session.Exit(0)
}

func selectSshHandler(cmd string, gitSshHandler, sdkSshHandler SshSessionHandler) SshSessionHandler {
	if cmd == "" {
		return gitSshHandler
	}

	parts := strings.SplitN(cmd, " ", 2)
	if len(parts) != 2 {
		return gitSshHandler
	}

	repoPath := strings.Trim(parts[1], "'\"")
	repoPath = strings.TrimPrefix(repoPath, "/")
	if strings.HasPrefix(repoPath, "sdk/") {
		return sdkSshHandler
	}

	return gitSshHandler
}

func LoadOrGenerateHostKey(path string) (gossh.Signer, error) {
	// #nosec G304 -- path is from application config, not user input
	keyBytes, err := os.ReadFile(path)
	if err == nil {
		signer, parseErr := gossh.ParsePrivateKey(keyBytes)
		if parseErr != nil {
			return nil, fmt.Errorf("failed to parse existing SSH host key %s: %w", path, parseErr)
		}
		zap.L().Info("Loaded SSH host key", zap.String("path", path))
		return signer, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("failed to read SSH host key %s: %w", path, err)
	}

	zap.L().Info("Generating SSH host key", zap.String("path", path))
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// #nosec G304 -- path is from application config, not user input
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			// #nosec G304 -- path is from application config, not user input
			existingBytes, readErr := os.ReadFile(path)
			if readErr != nil {
				return nil, fmt.Errorf("failed to read concurrently created SSH host key %s: %w", path, readErr)
			}
			signer, parseErr := gossh.ParsePrivateKey(existingBytes)
			if parseErr != nil {
				return nil, fmt.Errorf("failed to parse concurrently created SSH host key %s: %w", path, parseErr)
			}
			zap.L().Info("Loaded SSH host key", zap.String("path", path))
			return signer, nil
		}
		return nil, err
	}

	if _, writeErr := file.Write(privateKeyPEM); writeErr != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return nil, writeErr
	}
	if closeErr := file.Close(); closeErr != nil {
		_ = os.Remove(path)
		return nil, closeErr
	}

	return gossh.NewSignerFromKey(privateKey)
}
