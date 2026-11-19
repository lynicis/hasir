package registry

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/gliderlabs/ssh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	gossh "golang.org/x/crypto/ssh"

	"hasir/api/internal/user"
	"hasir/api/pkg/config"
)

type stubSshHandler struct {
	called bool
	userId string
	err    error
}

func (s *stubSshHandler) HandleSession(_ ssh.Session, userId string) error {
	s.called = true
	s.userId = userId
	return s.err
}

type fakeSshContext struct {
	context.Context
	mu   sync.Mutex
	vals map[any]any
}

func newFakeSshContext() *fakeSshContext {
	return &fakeSshContext{
		Context: context.Background(),
		vals:    make(map[any]any),
	}
}

func (c *fakeSshContext) Value(key any) any {
	c.mu.Lock()
	defer c.mu.Unlock()
	if v, ok := c.vals[key]; ok {
		return v
	}
	return c.Context.Value(key)
}

func (c *fakeSshContext) Lock()                           {}
func (c *fakeSshContext) Unlock()                         {}
func (c *fakeSshContext) User() string                    { return "git" }
func (c *fakeSshContext) SessionID() string               { return "sess-1" }
func (c *fakeSshContext) ClientVersion() string           { return "SSH-2.0-test" }
func (c *fakeSshContext) ServerVersion() string           { return "SSH-2.0-hasir" }
func (c *fakeSshContext) RemoteAddr() net.Addr            { return &net.TCPAddr{} }
func (c *fakeSshContext) LocalAddr() net.Addr             { return &net.TCPAddr{} }
func (c *fakeSshContext) Permissions() *ssh.Permissions   { return &ssh.Permissions{} }
func (c *fakeSshContext) SetValue(key, value any)         { c.mu.Lock(); defer c.mu.Unlock(); c.vals[key] = value }

type fakeSshSession struct {
	ctx      ssh.Context
	rawCmd   string
	stderr   bytes.Buffer
	exitCode int
	exited   bool
}

func (s *fakeSshSession) Read([]byte) (int, error)         { return 0, io.EOF }
func (s *fakeSshSession) Write(p []byte) (int, error)      { return len(p), nil }
func (s *fakeSshSession) Close() error                     { return nil }
func (s *fakeSshSession) CloseWrite() error                { return nil }
func (s *fakeSshSession) SendRequest(string, bool, []byte) (bool, error) { return false, nil }
func (s *fakeSshSession) Stderr() io.ReadWriter            { return &s.stderr }
func (s *fakeSshSession) User() string                     { return "git" }
func (s *fakeSshSession) RemoteAddr() net.Addr             { return &net.TCPAddr{} }
func (s *fakeSshSession) LocalAddr() net.Addr              { return &net.TCPAddr{} }
func (s *fakeSshSession) Environ() []string                { return nil }
func (s *fakeSshSession) Exit(code int) error              { s.exited = true; s.exitCode = code; return nil }
func (s *fakeSshSession) Command() []string                { return nil }
func (s *fakeSshSession) RawCommand() string               { return s.rawCmd }
func (s *fakeSshSession) Subsystem() string                { return "" }
func (s *fakeSshSession) PublicKey() ssh.PublicKey         { return nil }
func (s *fakeSshSession) Context() ssh.Context             { return s.ctx }
func (s *fakeSshSession) Permissions() ssh.Permissions     { return ssh.Permissions{} }
func (s *fakeSshSession) Pty() (ssh.Pty, <-chan ssh.Window, bool) { return ssh.Pty{}, nil, false }
func (s *fakeSshSession) Signals(chan<- ssh.Signal)        {}
func (s *fakeSshSession) Break(chan<- bool)                {}

func TestLoadOrGenerateHostKey(t *testing.T) {
	t.Run("generates key when missing and reloads existing key", func(t *testing.T) {
		keyPath := filepath.Join(t.TempDir(), "host_key")

		signer1, err := LoadOrGenerateHostKey(keyPath)
		require.NoError(t, err)
		require.NotNil(t, signer1)

		info, err := os.Stat(keyPath)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

		signer2, err := LoadOrGenerateHostKey(keyPath)
		require.NoError(t, err)
		assert.Equal(t, signer1.PublicKey().Marshal(), signer2.PublicKey().Marshal())
	})

	t.Run("returns error when destination directory does not exist", func(t *testing.T) {
		_, err := LoadOrGenerateHostKey(filepath.Join(t.TempDir(), "missing", "host_key"))
		require.Error(t, err)
	})
}

func TestSelectSshHandler(t *testing.T) {
	gitHandler := &stubSshHandler{}
	sdkHandler := &stubSshHandler{}

	assert.Same(t, gitHandler, selectSshHandler("", gitHandler, sdkHandler))
	assert.Same(t, gitHandler, selectSshHandler("whoami", gitHandler, sdkHandler))
	assert.Same(t, gitHandler, selectSshHandler("git-upload-pack '/repo-123.git'", gitHandler, sdkHandler))
	assert.Same(t, sdkHandler, selectSshHandler("git-upload-pack '/sdk/org/repo/go-protobuf'", gitHandler, sdkHandler))
	assert.Same(t, sdkHandler, selectSshHandler("git-upload-pack 'sdk/org/repo/go-protobuf'", gitHandler, sdkHandler))
}

func TestHandleSshSession(t *testing.T) {
	t.Run("missing userId exits with 1", func(t *testing.T) {
		sess := &fakeSshSession{ctx: newFakeSshContext()}
		gitHandler := &stubSshHandler{}
		sdkHandler := &stubSshHandler{}

		handleSshSession(sess, gitHandler, sdkHandler)
		assert.True(t, sess.exited)
		assert.Equal(t, 1, sess.exitCode)
		assert.Contains(t, sess.stderr.String(), "Authentication required")
		assert.False(t, gitHandler.called)
		assert.False(t, sdkHandler.called)
	})

	t.Run("routes sdk command to sdk handler and exits 0", func(t *testing.T) {
		ctx := newFakeSshContext()
		ctx.SetValue("userId", "user-1")
		sess := &fakeSshSession{
			ctx:    ctx,
			rawCmd: "git-upload-pack '/sdk/org/repo/go-protobuf'",
		}
		gitHandler := &stubSshHandler{}
		sdkHandler := &stubSshHandler{}

		handleSshSession(sess, gitHandler, sdkHandler)
		assert.True(t, sess.exited)
		assert.Equal(t, 0, sess.exitCode)
		assert.True(t, sdkHandler.called)
		assert.Equal(t, "user-1", sdkHandler.userId)
		assert.False(t, gitHandler.called)
	})

	t.Run("handler error exits with 1", func(t *testing.T) {
		ctx := newFakeSshContext()
		ctx.SetValue("userId", "user-2")
		sess := &fakeSshSession{
			ctx:    ctx,
			rawCmd: "git-receive-pack '/repo-1'",
		}
		gitHandler := &stubSshHandler{err: errors.New("permission denied")}
		sdkHandler := &stubSshHandler{}

		handleSshSession(sess, gitHandler, sdkHandler)
		assert.True(t, sess.exited)
		assert.Equal(t, 1, sess.exitCode)
		assert.Contains(t, sess.stderr.String(), "permission denied")
	})
}

func TestNewSshServer_PublicKeyHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockUserRepo := user.NewMockRepository(ctrl)

	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	pubKey, err := gossh.NewPublicKey(&privKey.PublicKey)
	require.NoError(t, err)

	cfg := &config.Config{
		Ssh: config.SshConfig{
			Port:        "2222",
			HostKeyPath: filepath.Join(t.TempDir(), "host_key"),
		},
	}

	srv, err := NewSshServer(cfg, mockUserRepo, &stubSshHandler{}, &stubSshHandler{})
	require.NoError(t, err)

	t.Run("accepts known public key and stores userId in context", func(t *testing.T) {
		mockUserRepo.EXPECT().
			GetUserBySshPublicKey(gomock.Any(), gomock.Any()).
			Return(&user.UserDTO{Id: "user-abc"}, nil)

		ctx := newFakeSshContext()
		assert.True(t, srv.PublicKeyHandler(ctx, pubKey))
		assert.Equal(t, "user-abc", ctx.Value("userId"))
	})

	t.Run("rejects unknown public key", func(t *testing.T) {
		mockUserRepo.EXPECT().
			GetUserBySshPublicKey(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("not found"))

		ctx := newFakeSshContext()
		assert.False(t, srv.PublicKeyHandler(ctx, pubKey))
		assert.Nil(t, ctx.Value("userId"))
	})
}
