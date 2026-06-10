package sdkgenerator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var importRegex = regexp.MustCompile(`(?m)^\s*import\s+(?:public\s+|weak\s+)?"([^"]+)"\s*;`)

func ParseProtoImports(content string) []string {
	matches := importRegex.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}

	imports := make([]string, 0, len(matches))
	for _, match := range matches {
		imports = append(imports, match[1])
	}

	return imports
}

func ParseProtoImportsFromFiles(repoPath string, protoFiles []string) ([]string, error) {
	seen := make(map[string]struct{})
	var imports []string

	for _, protoFile := range protoFiles {
		fullPath := filepath.Join(repoPath, protoFile)
		// #nosec G304 -- repoPath is from config, protoFile validated by Validate() which checks path traversal
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read proto file %s: %w", protoFile, err)
		}

		for _, imp := range ParseProtoImports(string(content)) {
			if _, ok := seen[imp]; !ok {
				seen[imp] = struct{}{}
				imports = append(imports, imp)
			}
		}
	}

	return imports, nil
}

var defaultBsrModuleMap = map[string]string{
	"google/protobuf/":    "buf.build/protocolbuffers/wellknowntypes",
	"google/api/":         "buf.build/googleapis/googleapis",
	"google/rpc/":         "buf.build/googleapis/googleapis",
	"google/type/":        "buf.build/googleapis/googleapis",
	"google/longrunning/": "buf.build/googleapis/googleapis",
	"buf/validate/":       "buf.build/bufbuild/protovalidate",
	"grpc/":               "buf.build/grpc/grpc",
}

func ResolveBsrModules(imports []string) []string {
	return ResolveBsrModulesWithCustom(imports, nil)
}

func ResolveBsrModulesWithCustom(imports []string, customModules map[string]string) []string {
	seen := make(map[string]struct{})
	var modules []string

	resolve := func(imp string) {
		for prefix, module := range defaultBsrModuleMap {
			if strings.HasPrefix(imp, prefix) {
				if _, ok := seen[module]; !ok {
					seen[module] = struct{}{}
					modules = append(modules, module)
				}
				return
			}
		}

		for prefix, module := range customModules {
			if strings.HasPrefix(imp, prefix) {
				if _, ok := seen[module]; !ok {
					seen[module] = struct{}{}
					modules = append(modules, module)
				}
				return
			}
		}
	}

	for _, imp := range imports {
		resolve(imp)
	}

	return modules
}

func GenerateBufYaml(dir string, modules []string) error {
	var sb strings.Builder
	sb.WriteString("version: v2\n")
	sb.WriteString("modules:\n")
	sb.WriteString("  - path: .\n")

	if len(modules) > 0 {
		sb.WriteString("deps:\n")
		for _, module := range modules {
			sb.WriteString("  - " + module + "\n")
		}
	}

	// #nosec G306 -- buf.yaml needs to be readable by buf CLI
	return os.WriteFile(filepath.Join(dir, "buf.yaml"), []byte(sb.String()), 0o644)
}

type BufGenPlugin struct {
	Remote string
	Out    string
	Opt    string
	Opts   []string
}

func GenerateBufGenYaml(dir string, plugins []BufGenPlugin) error {
	var sb strings.Builder
	sb.WriteString("version: v2\n")
	sb.WriteString("managed:\n")
	sb.WriteString("  enabled: true\n")
	sb.WriteString("  disable:\n")
	sb.WriteString("    - file_option: go_package\n")
	sb.WriteString("      module: buf.build/bufbuild/protovalidate\n")
	sb.WriteString("plugins:\n")

	for _, plugin := range plugins {
		sb.WriteString("  - remote: " + plugin.Remote + "\n")
		sb.WriteString("    out: " + plugin.Out + "\n")

		opts := plugin.Opts
		if plugin.Opt != "" && len(opts) == 0 {
			opts = []string{plugin.Opt}
		}

		if len(opts) > 0 {
			sb.WriteString("    opt:\n")
			for _, opt := range opts {
				sb.WriteString("      - " + opt + "\n")
			}
		}
	}

	// #nosec G306 -- buf.gen.yaml needs to be readable by buf CLI
	return os.WriteFile(filepath.Join(dir, "buf.gen.yaml"), []byte(sb.String()), 0o644)
}
