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

type bufModule struct {
	path string
	name string
}

func GenerateBufYaml(dir string, modules []string) error {
	bufYamlPath := filepath.Join(dir, "buf.yaml")
	var existingModules []bufModule

	if existingContent, err := os.ReadFile(bufYamlPath); err == nil {
		lines := strings.Split(string(existingContent), "\n")
		var current *bufModule

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "- path:") {
				if current != nil {
					existingModules = append(existingModules, *current)
				}
				path := strings.TrimSpace(strings.TrimPrefix(trimmed, "- path:"))
				current = &bufModule{path: path}
			} else if strings.HasPrefix(trimmed, "path:") && current == nil {
				path := strings.TrimSpace(strings.TrimPrefix(trimmed, "path:"))
				current = &bufModule{path: path}
			} else if current != nil {
				if strings.HasPrefix(trimmed, "name:") {
					current.name = strings.TrimSpace(strings.TrimPrefix(trimmed, "name:"))
				} else if strings.HasPrefix(trimmed, "-") || (strings.Contains(trimmed, ":") && !strings.HasPrefix(trimmed, "name:")) {
					existingModules = append(existingModules, *current)
					current = nil
				}
			}
		}
		if current != nil {
			existingModules = append(existingModules, *current)
		}
	}

	if len(existingModules) == 0 {
		existingModules = []bufModule{{path: "."}}
	}

	var sb strings.Builder
	sb.WriteString("version: v2\n")
	sb.WriteString("modules:\n")
	for _, m := range existingModules {
		sb.WriteString("  - path: ")
		sb.WriteString(m.path)
		sb.WriteString("\n")
		if m.name != "" {
			sb.WriteString("    name: ")
			sb.WriteString(m.name)
			sb.WriteString("\n")
		}
	}

	if len(modules) > 0 {
		sb.WriteString("deps:\n")
		for _, module := range modules {
			sb.WriteString("  - ")
			sb.WriteString(module)
			sb.WriteString("\n")
		}
	}

	// #nosec G306 -- buf.yaml needs to be readable by buf CLI
	return os.WriteFile(bufYamlPath, []byte(sb.String()), 0o644)
}

type BufGenPlugin struct {
	Remote string
	Out    string
	Opt    string
	Opts   []string
}

func GenerateBufGenYaml(dir string, plugins []BufGenPlugin, goPackagePrefix string) error {
	var sb strings.Builder
	sb.WriteString("version: v2\n")
	sb.WriteString("managed:\n")
	sb.WriteString("  enabled: true\n")
	sb.WriteString("  disable:\n")
	sb.WriteString("    - file_option: go_package\n")
	sb.WriteString("      module: buf.build/bufbuild/protovalidate\n")
	if goPackagePrefix != "" {
		sb.WriteString("  override:\n")
		sb.WriteString("    - file_option: go_package_prefix\n")
		sb.WriteString("      value: ")
		sb.WriteString(goPackagePrefix)
		sb.WriteString("\n")
	}
	sb.WriteString("plugins:\n")

	for _, plugin := range plugins {
		sb.WriteString("  - remote: ")
		sb.WriteString(plugin.Remote)
		sb.WriteString("\n")
		sb.WriteString("    out: ")
		sb.WriteString(plugin.Out)
		sb.WriteString("\n")

		opts := plugin.Opts
		if plugin.Opt != "" && len(opts) == 0 {
			opts = []string{plugin.Opt}
		}

		if len(opts) > 0 {
			sb.WriteString("    opt:\n")
			for _, opt := range opts {
				sb.WriteString("      - ")
				sb.WriteString(opt)
				sb.WriteString("\n")
			}
		}
	}

	// #nosec G306 -- buf.gen.yaml needs to be readable by buf CLI
	return os.WriteFile(filepath.Join(dir, "buf.gen.yaml"), []byte(sb.String()), 0o644)
}
