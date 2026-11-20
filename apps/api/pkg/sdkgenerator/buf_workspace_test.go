package sdkgenerator

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for path, content := range files {
		full := filepath.Join(root, path)
		require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o750))
		require.NoError(t, os.WriteFile(full, []byte(content), 0o644))
	}
}

func TestResolveBufWorkspace(t *testing.T) {
	tests := []struct {
		name          string
		files         map[string]string
		wantDir       string
		wantHasConfig bool
		wantHasLock   bool
		wantErr       error
	}{
		{
			name:          "root buf.yaml",
			files:         map[string]string{"buf.yaml": "version: v2\n", "nested/buf.yaml": "version: v2\n"},
			wantDir:       ".",
			wantHasConfig: true,
		},
		{
			name:          "root buf.work.yaml with lock",
			files:         map[string]string{"buf.work.yaml": "version: v1\n", "buf.lock": "version: v2\n"},
			wantDir:       ".",
			wantHasConfig: true,
			wantHasLock:   true,
		},
		{
			name: "single nested workspace in monorepo",
			files: map[string]string{
				"packages/proto/buf.yaml":                 "version: v2\n",
				"packages/proto/buf.lock":                 "version: v2\n",
				"packages/proto/proto/shared/role.proto":  "syntax = \"proto3\";",
				".agents/skills/protobuf/assets/buf.yaml": "version: v2\n",
				"node_modules/pkg/buf.yaml":               "version: v2\n",
				"vendor/pkg/buf.yaml":                     "version: v2\n",
			},
			wantDir:       "packages/proto",
			wantHasConfig: true,
			wantHasLock:   true,
		},
		{
			name: "outer workspace wins over nested modules",
			files: map[string]string{
				"schemas/buf.work.yaml":   "version: v1\n",
				"schemas/a/buf.yaml":      "version: v1\n",
				"schemas/b/buf.yaml":      "version: v1\n",
				"schemas/a/a.proto":       "syntax = \"proto3\";",
				"schemas/b/nested/b.yaml": "",
			},
			wantDir:       "schemas",
			wantHasConfig: true,
		},
		{
			name:    "no config falls back to root",
			files:   map[string]string{"user.proto": "syntax = \"proto3\";"},
			wantDir: ".",
		},
		{
			name: "multiple workspaces are ambiguous",
			files: map[string]string{
				"a/buf.yaml": "version: v2\n",
				"b/buf.yaml": "version: v2\n",
			},
			wantErr: ErrAmbiguousBufWorkspace,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			writeFiles(t, root, tt.files)

			ws, err := ResolveBufWorkspace(root)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Contains(t, err.Error(), "a, b")
				return
			}

			require.NoError(t, err)
			assert.Equal(t, filepath.Join(root, tt.wantDir), ws.Dir)
			assert.Equal(t, tt.wantHasConfig, ws.HasConfig)
			assert.Equal(t, tt.wantHasLock, ws.HasLock)
		})
	}
}

func TestBufGenerator_Generate_NestedWorkspace(t *testing.T) {
	ctx := context.Background()
	mockRunner := NewMockCommandRunner()
	g := NewBufGoProtobufGenerator(mockRunner)

	root := t.TempDir()
	bufYaml := "version: v2\nmodules:\n  - path: proto/\n    excludes:\n      - proto/internal\ndeps:\n  - buf.build/bufbuild/protovalidate\nlint:\n  use:\n    - BASIC\n"
	userBufGen := "version: v2\nplugins:\n  - local: protoc-gen-go\n    out: gen\n"
	writeFiles(t, root, map[string]string{
		"packages/proto/buf.yaml":                                 bufYaml,
		"packages/proto/buf.lock":                                 "version: v2\n",
		"packages/proto/buf.gen.yaml":                             userBufGen,
		"packages/proto/proto/shared/role.proto":                  "syntax = \"proto3\";",
		"packages/proto/proto/user/v1/user.proto":                 "syntax = \"proto3\";\nimport \"shared/role.proto\";",
		".agents/skills/protobuf/assets/proto/example/v1/x.proto": "syntax = \"proto3\";\nimport \"example/v1/missing.proto\";",
	})

	protoFiles, err := FindProtoFiles(root)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{
		filepath.Join("packages", "proto", "proto", "shared", "role.proto"),
		filepath.Join("packages", "proto", "proto", "user", "v1", "user.proto"),
	}, protoFiles)

	_, err = g.Generate(ctx, GeneratorInput{RepoPath: root, OutputPath: t.TempDir(), ProtoFiles: protoFiles})
	require.NoError(t, err)

	wsDir := filepath.Join(root, "packages", "proto")

	// buf.lock exists, so dependencies are not re-resolved; only buf generate runs.
	require.Len(t, mockRunner.Calls, 1)
	assert.Equal(t, "buf", mockRunner.Calls[0].Name)
	assert.Equal(t, []string{"generate", "--template", filepath.Join(wsDir, BufGenTemplateName)}, mockRunner.Calls[0].Args)
	assert.Equal(t, wsDir, mockRunner.Calls[0].WorkDir)

	gotBufYaml, err := os.ReadFile(filepath.Join(wsDir, "buf.yaml"))
	require.NoError(t, err)
	assert.Equal(t, bufYaml, string(gotBufYaml), "existing buf.yaml must not be rewritten")

	gotUserBufGen, err := os.ReadFile(filepath.Join(wsDir, "buf.gen.yaml"))
	require.NoError(t, err)
	assert.Equal(t, userBufGen, string(gotUserBufGen), "existing buf.gen.yaml must not be rewritten")

	_, err = os.Stat(filepath.Join(root, "buf.yaml"))
	assert.True(t, os.IsNotExist(err), "no buf.yaml should be generated at the repo root")
}

func TestBufGenerator_Generate_NestedWorkspaceWithoutLock(t *testing.T) {
	ctx := context.Background()
	mockRunner := NewMockCommandRunner()
	g := NewBufGoProtobufGenerator(mockRunner)

	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		"proto/buf.yaml":      "version: v2\n",
		"proto/user/v1.proto": "syntax = \"proto3\";",
	})

	_, err := g.Generate(ctx, GeneratorInput{RepoPath: root, OutputPath: t.TempDir(), ProtoFiles: []string{"proto/user/v1.proto"}})
	require.NoError(t, err)

	require.Len(t, mockRunner.Calls, 2)
	assert.Equal(t, []string{"dep", "update"}, mockRunner.Calls[0].Args)
	assert.Equal(t, filepath.Join(root, "proto"), mockRunner.Calls[0].WorkDir)
}

func TestBufGenerator_Generate_AmbiguousWorkspace(t *testing.T) {
	mockRunner := NewMockCommandRunner()
	g := NewBufGoProtobufGenerator(mockRunner)

	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		"a/buf.yaml": "version: v2\n",
		"b/buf.yaml": "version: v2\n",
		"a/x.proto":  "syntax = \"proto3\";",
	})

	_, err := g.Generate(context.Background(), GeneratorInput{RepoPath: root, OutputPath: t.TempDir(), ProtoFiles: []string{"a/x.proto"}})
	require.ErrorIs(t, err, ErrAmbiguousBufWorkspace)
	assert.Empty(t, mockRunner.Calls)
}

func TestDocumentationGenerator_Generate_NestedWorkspace(t *testing.T) {
	mockRunner := NewMockCommandRunner()
	g := NewDocumentationGenerator(mockRunner)

	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		"packages/proto/buf.yaml":                "version: v2\n",
		"packages/proto/buf.lock":                "version: v2\n",
		"packages/proto/proto/shared/role.proto": "syntax = \"proto3\";",
	})
	outputPath := filepath.Join(root, "docs")
	require.NoError(t, os.MkdirAll(outputPath, 0o750))

	_, err := g.Generate(context.Background(), GeneratorInput{
		RepoPath:   root,
		OutputPath: outputPath,
		ProtoFiles: []string{"packages/proto/proto/shared/role.proto"},
	})
	require.NoError(t, err)

	require.Len(t, mockRunner.Calls, 1)
	assert.Equal(t, "buf", mockRunner.Calls[0].Name)
	assert.Contains(t, mockRunner.Calls[0].Args, "--template")
	assert.Equal(t, filepath.Join(root, "packages", "proto"), mockRunner.Calls[0].WorkDir)
}

func TestIsSkippedPath(t *testing.T) {
	assert.True(t, IsSkippedPath(".agents/skills/protobuf/assets/buf.gen.yaml"))
	assert.True(t, IsSkippedPath("apps/dashboard/node_modules/foo/a.proto"))
	assert.True(t, IsSkippedPath("vendor/example/a.proto"))
	assert.False(t, IsSkippedPath("buf.gen.yaml"))
	assert.False(t, IsSkippedPath("packages/proto/buf.gen.yaml"))
	assert.False(t, IsSkippedPath("packages/proto/proto/user/v1/user.proto"))
}

func TestBufGenerator_Generate_BufWorkYamlUpdatesUnlockedModules(t *testing.T) {
	mockRunner := NewMockCommandRunner()
	g := NewBufGoProtobufGenerator(mockRunner)

	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		"schemas/buf.work.yaml": "version: v1\ndirectories:\n  - a\n  - b\n",
		"schemas/a/buf.yaml":    "version: v1\n",
		"schemas/a/buf.lock":    "version: v1beta1\n",
		"schemas/a/a.proto":     "syntax = \"proto3\";",
		"schemas/b/buf.yaml":    "version: v1\n",
		"schemas/b/b.proto":     "syntax = \"proto3\";",
	})

	_, err := g.Generate(context.Background(), GeneratorInput{
		RepoPath:   root,
		OutputPath: t.TempDir(),
		ProtoFiles: []string{"schemas/a/a.proto", "schemas/b/b.proto"},
	})
	require.NoError(t, err)

	wsDir := filepath.Join(root, "schemas")
	require.Len(t, mockRunner.Calls, 2)
	assert.Equal(t, []string{"dep", "update", "b"}, mockRunner.Calls[0].Args)
	assert.Equal(t, wsDir, mockRunner.Calls[0].WorkDir)
	assert.Equal(t, []string{"generate", "--template", filepath.Join(wsDir, BufGenTemplateName)}, mockRunner.Calls[1].Args)
	assert.Equal(t, wsDir, mockRunner.Calls[1].WorkDir)
}

func TestResolveBufWorkspace_BufWorkYamlAllLockedAndTraversal(t *testing.T) {
	t.Run("all modules locked skips dep update", func(t *testing.T) {
		root := t.TempDir()
		writeFiles(t, root, map[string]string{
			"buf.work.yaml": "version: v1\ndirectories:\n  - a\n",
			"a/buf.yaml":    "version: v1\n",
			"a/buf.lock":    "version: v1beta1\n",
		})

		ws, err := ResolveBufWorkspace(root)
		require.NoError(t, err)
		assert.True(t, ws.HasConfig)
		assert.True(t, ws.HasLock)
		assert.Empty(t, ws.DepUpdateDirs)
	})

	t.Run("rejects traversal directory in buf.work.yaml", func(t *testing.T) {
		root := t.TempDir()
		writeFiles(t, root, map[string]string{
			"buf.work.yaml": "version: v1\ndirectories:\n  - ../outside\n",
		})

		_, err := ResolveBufWorkspace(root)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid directory")
	})

	t.Run("rejects flag-like directory in buf.work.yaml", func(t *testing.T) {
		root := t.TempDir()
		writeFiles(t, root, map[string]string{
			"buf.work.yaml": "version: v1\ndirectories:\n  - -x\n",
		})

		_, err := ResolveBufWorkspace(root)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid directory")
	})
}

func TestGenerateBufFiles_DoesNotFollowSymlinks(t *testing.T) {
	outsideDir := t.TempDir()
	existingTarget := filepath.Join(outsideDir, "secret.txt")
	require.NoError(t, os.WriteFile(existingTarget, []byte("original"), 0o600))
	danglingTarget := filepath.Join(outsideDir, "created.txt")

	repoDir := t.TempDir()
	require.NoError(t, os.Symlink(existingTarget, filepath.Join(repoDir, BufGenTemplateName)))
	require.NoError(t, os.Symlink(danglingTarget, filepath.Join(repoDir, "buf.yaml")))

	require.NoError(t, GenerateBufGenYaml(repoDir, []BufGenPlugin{{Remote: "buf.build/protocolbuffers/go", Out: "."}}, ""))
	require.NoError(t, GenerateBufYaml(repoDir, nil))

	gotExisting, err := os.ReadFile(existingTarget)
	require.NoError(t, err)
	assert.Equal(t, "original", string(gotExisting), "existing symlink target must not be overwritten")

	_, err = os.Stat(danglingTarget)
	assert.True(t, os.IsNotExist(err), "dangling symlink target must not be created")
}
