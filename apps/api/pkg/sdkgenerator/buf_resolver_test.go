package sdkgenerator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseProtoImports(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name: "single import",
			content: `syntax = "proto3";
import "google/protobuf/timestamp.proto";
`,
			expected: []string{"google/protobuf/timestamp.proto"},
		},
		{
			name: "multiple imports",
			content: `syntax = "proto3";
import "google/protobuf/timestamp.proto";
import "google/protobuf/empty.proto";
import "buf/validate/validate.proto";
`,
			expected: []string{
				"google/protobuf/timestamp.proto",
				"google/protobuf/empty.proto",
				"buf/validate/validate.proto",
			},
		},
		{
			name: "public imports",
			content: `syntax = "proto3";
import public "google/protobuf/timestamp.proto";
import "buf/validate/validate.proto";
`,
			expected: []string{
				"google/protobuf/timestamp.proto",
				"buf/validate/validate.proto",
			},
		},
		{
			name: "weak imports",
			content: `syntax = "proto3";
import weak "google/protobuf/descriptor.proto";
`,
			expected: []string{"google/protobuf/descriptor.proto"},
		},
		{
			name:     "no imports",
			content:  `syntax = "proto3";`,
			expected: nil,
		},
		{
			name: "skips comments",
			content: `syntax = "proto3";
// import "google/protobuf/timestamp.proto";
import "buf/validate/validate.proto";
`,
			expected: []string{"buf/validate/validate.proto"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseProtoImports(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseProtoImportsFromFiles(t *testing.T) {
	tmpDir := t.TempDir()

	protoContent := `syntax = "proto3";
package user.v1;
import "google/protobuf/timestamp.proto";
import "buf/validate/validate.proto";
import "shared/filter.proto";
`
	err := os.WriteFile(filepath.Join(tmpDir, "user.proto"), []byte(protoContent), 0o644)
	require.NoError(t, err)

	imports, err := ParseProtoImportsFromFiles(tmpDir, []string{"user.proto"})
	require.NoError(t, err)
	assert.Contains(t, imports, "google/protobuf/timestamp.proto")
	assert.Contains(t, imports, "buf/validate/validate.proto")
	assert.Contains(t, imports, "shared/filter.proto")
}

func TestParseProtoImportsFromFiles_Deduplication(t *testing.T) {
	tmpDir := t.TempDir()

	proto1 := `syntax = "proto3";
import "google/protobuf/timestamp.proto";
`
	proto2 := `syntax = "proto3";
import "google/protobuf/timestamp.proto";
import "google/protobuf/empty.proto";
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "a.proto"), []byte(proto1), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "b.proto"), []byte(proto2), 0o644))

	imports, err := ParseProtoImportsFromFiles(tmpDir, []string{"a.proto", "b.proto"})
	require.NoError(t, err)
	assert.Len(t, imports, 2)
}

func TestResolveBsrModules(t *testing.T) {
	tests := []struct {
		name     string
		imports  []string
		expected []string
	}{
		{
			name:     "google well-known types",
			imports:  []string{"google/protobuf/timestamp.proto", "google/protobuf/empty.proto"},
			expected: []string{"buf.build/protocolbuffers/wellknowntypes"},
		},
		{
			name:     "buf validate",
			imports:  []string{"buf/validate/validate.proto"},
			expected: []string{"buf.build/bufbuild/protovalidate"},
		},
		{
			name:     "google apis",
			imports:  []string{"google/api/annotations.proto", "google/rpc/status.proto"},
			expected: []string{"buf.build/googleapis/googleapis"},
		},
		{
			name:     "grpc",
			imports:  []string{"grpc/reflection/v1alpha/reflection.proto"},
			expected: []string{"buf.build/grpc/grpc"},
		},
		{
			name:     "mixed imports",
			imports:  []string{"google/protobuf/timestamp.proto", "buf/validate/validate.proto", "shared/filter.proto"},
			expected: []string{"buf.build/protocolbuffers/wellknowntypes", "buf.build/bufbuild/protovalidate"},
		},
		{
			name:     "local imports only",
			imports:  []string{"shared/filter.proto", "user/v1/user.proto"},
			expected: nil,
		},
		{
			name:     "no imports",
			imports:  nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveBsrModules(tt.imports)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResolveBsrModulesWithCustom(t *testing.T) {
	custom := map[string]string{
		"myorg/": "buf.build/myorg/mymodule",
	}

	imports := []string{"myorg/shared/types.proto", "google/protobuf/timestamp.proto"}
	result := ResolveBsrModulesWithCustom(imports, custom)

	assert.Contains(t, result, "buf.build/myorg/mymodule")
	assert.Contains(t, result, "buf.build/protocolbuffers/wellknowntypes")
	assert.Len(t, result, 2)
}

func TestGenerateBufYaml(t *testing.T) {
	tmpDir := t.TempDir()

	modules := []string{
		"buf.build/protocolbuffers/wellknowntypes",
		"buf.build/bufbuild/protovalidate",
	}

	err := GenerateBufYaml(tmpDir, modules)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "buf.yaml"))
	require.NoError(t, err)

	expected := `version: v2
modules:
  - path: .
deps:
  - buf.build/protocolbuffers/wellknowntypes
  - buf.build/bufbuild/protovalidate
`
	assert.Equal(t, expected, string(content))
}

func TestGenerateBufYaml_NoDeps(t *testing.T) {
	tmpDir := t.TempDir()

	err := GenerateBufYaml(tmpDir, nil)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "buf.yaml"))
	require.NoError(t, err)

	expected := `version: v2
modules:
  - path: .
`
	assert.Equal(t, expected, string(content))
}

func TestGenerateBufYaml_ExistingFile(t *testing.T) {
	tmpDir := t.TempDir()

	existing := "version: v2\nmodules:\n  - path: proto/\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "buf.yaml"), []byte(existing), 0o644))

	err := GenerateBufYaml(tmpDir, []string{"buf.build/protocolbuffers/wellknowntypes"})
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "buf.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "buf.build/protocolbuffers/wellknowntypes")
}
