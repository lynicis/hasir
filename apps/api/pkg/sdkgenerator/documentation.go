package sdkgenerator

import (
	"embed"
	"os"
	"path/filepath"
)

//go:embed template/*.mustache
var templateFS embed.FS

type DocumentationGenerator struct {
	*protocGenerator
}

func NewDocumentationGenerator(runner CommandRunner) *DocumentationGenerator {
	return &DocumentationGenerator{
		protocGenerator: newProtocGenerator("", "docs", runner, buildDocumentationArgs),
	}
}

func buildDocumentationArgs(input GeneratorInput) []string {
	templateContent, err := templateFS.ReadFile("template/proto-doc-tempate.mustache")
	if err != nil {
		args := []string{
			"--proto_path=" + filepath.Clean(input.RepoPath),
			"--doc_out=" + filepath.Clean(input.OutputPath),
			"--doc_opt=markdown,index.md",
		}
		args = append(args, input.ProtoFiles...)
		return args
	}

	templatePath := filepath.Join(input.OutputPath, "proto-doc-template.mustache")
	if err := os.WriteFile(templatePath, templateContent, 0o644); err != nil {
		args := []string{
			"--proto_path=" + filepath.Clean(input.RepoPath),
			"--doc_out=" + filepath.Clean(input.OutputPath),
			"--doc_opt=markdown,index.md",
		}
		args = append(args, input.ProtoFiles...)
		return args
	}

	args := []string{
		"--proto_path=" + filepath.Clean(input.RepoPath),
		"--doc_out=" + filepath.Clean(input.OutputPath),
		"--doc_opt=markdown,index.md",
	}
	args = append(args, input.ProtoFiles...)

	return args
}
