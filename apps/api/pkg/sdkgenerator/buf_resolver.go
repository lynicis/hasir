package sdkgenerator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/goccy/go-yaml"
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

type bsrModuleMapping struct {
	prefix string
	module string
}

var defaultBsrModuleMap = []bsrModuleMapping{
	{"google/protobuf/", "buf.build/protocolbuffers/wellknowntypes"},
	{"google/api/", "buf.build/googleapis/googleapis"},
	{"google/rpc/", "buf.build/googleapis/googleapis"},
	{"google/type/", "buf.build/googleapis/googleapis"},
	{"google/longrunning/", "buf.build/googleapis/googleapis"},
	{"buf/validate/", "buf.build/bufbuild/protovalidate"},
	{"grpc/", "buf.build/grpc/grpc"},
}

func ResolveBsrModules(imports []string) []string {
	return ResolveBsrModulesWithCustom(imports, nil)
}

func ResolveBsrModulesWithCustom(imports []string, customModules map[string]string) []string {
	seen := make(map[string]struct{})
	var modules []string

	resolve := func(imp string) {
		for _, mapping := range defaultBsrModuleMap {
			if strings.HasPrefix(imp, mapping.prefix) {
				if _, ok := seen[mapping.module]; !ok {
					seen[mapping.module] = struct{}{}
					modules = append(modules, mapping.module)
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

type BufYaml struct {
	Version string      `yaml:"version"`
	Modules []bufModule `yaml:"modules"`
	Deps    []string    `yaml:"deps,omitempty"`
}

type bufModule struct {
	Path string `yaml:"path"`
	Name string `yaml:"name,omitempty"`
}

func GenerateBufYaml(dir string, modules []string) error {
	bufYamlPath := filepath.Join(dir, "buf.yaml")
	var conf BufYaml

	if existingContent, err := os.ReadFile(bufYamlPath); err == nil { // #nosec G304
		_ = yaml.Unmarshal(existingContent, &conf)
	}

	if conf.Version == "" {
		conf.Version = "v2"
	}
	if len(conf.Modules) == 0 {
		conf.Modules = []bufModule{{Path: "."}}
	}

	if len(modules) > 0 {
		conf.Deps = modules
	}

	out, err := yaml.MarshalWithOptions(&conf, yaml.Indent(2), yaml.IndentSequence(true))
	if err != nil {
		return fmt.Errorf("failed to marshal buf.yaml: %w", err)
	}

	// #nosec G306 -- buf.yaml needs to be readable by buf CLI
	return os.WriteFile(bufYamlPath, out, 0o644)
}

type BufGenPlugin struct {
	Remote string   `yaml:"remote,omitempty"`
	Out    string   `yaml:"out"`
	Opt    string   `yaml:"-"` // Not serialized directly, merged into Opts
	Opts   []string `yaml:"opt,omitempty"`
}

type BufGenYaml struct {
	Version string         `yaml:"version"`
	Managed *BufGenManaged `yaml:"managed,omitempty"`
	Plugins []BufGenPlugin `yaml:"plugins"`
}

type BufGenManaged struct {
	Enabled  bool             `yaml:"enabled"`
	Disable  []BufGenDisable  `yaml:"disable,omitempty"`
	Override []BufGenOverride `yaml:"override,omitempty"`
}

type BufGenDisable struct {
	FileOption string `yaml:"file_option"`
	Module     string `yaml:"module,omitempty"`
}

type BufGenOverride struct {
	FileOption string `yaml:"file_option"`
	Value      string `yaml:"value"`
}

func GenerateBufGenYaml(dir string, plugins []BufGenPlugin, goPackagePrefix string) error {
	conf := BufGenYaml{
		Version: "v2",
		Managed: &BufGenManaged{
			Enabled: true,
			Disable: []BufGenDisable{
				{
					FileOption: "go_package",
					Module:     "buf.build/bufbuild/protovalidate",
				},
			},
		},
	}

	if goPackagePrefix != "" {
		conf.Managed.Override = []BufGenOverride{
			{
				FileOption: "go_package_prefix",
				Value:      goPackagePrefix,
			},
		}
	}

	for i, plugin := range plugins {
		opts := plugin.Opts
		if plugin.Opt != "" && len(opts) == 0 {
			opts = []string{plugin.Opt}
		}
		plugins[i].Opts = opts
	}
	conf.Plugins = plugins

	out, err := yaml.MarshalWithOptions(&conf, yaml.Indent(2), yaml.IndentSequence(true))
	if err != nil {
		return fmt.Errorf("failed to marshal buf.gen.yaml: %w", err)
	}

	// #nosec G306 -- buf.gen.yaml needs to be readable by buf CLI
	return os.WriteFile(filepath.Join(dir, "buf.gen.yaml"), out, 0o644)
}
