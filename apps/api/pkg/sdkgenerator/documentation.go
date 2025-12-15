package sdkgenerator

import (
	"context"
	"embed"
	"os"
	"path/filepath"
	"strings"
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

func (g *DocumentationGenerator) Generate(ctx context.Context, input GeneratorInput) (*GeneratorOutput, error) {
	output, err := g.protocGenerator.Generate(ctx, input)
	if err != nil {
		return nil, err
	}

	indexMdPath := filepath.Join(input.OutputPath, "index.md")
	if err := removeScalarValueTypesSection(indexMdPath); err != nil {
		return output, err
	}

	return output, nil
}

func removeScalarValueTypesSection(filePath string) error {
	// #nosec G304 -- filePath is constructed from validated input.OutputPath which is already
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	var filteredLines []string
	inScalarSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "#") &&
			(strings.Contains(strings.ToLower(trimmed), "scalar value types") ||
				strings.Contains(strings.ToLower(trimmed), "scalar value type")) {
			inScalarSection = true
			continue
		}

		if inScalarSection {
			if strings.HasPrefix(trimmed, "#") {
				inScalarSection = false
				filteredLines = append(filteredLines, line)
				continue
			}
			continue
		}

		filteredLines = append(filteredLines, line)
	}

	filteredContent := strings.Join(filteredLines, "\n")
	originalContent := string(content)
	if strings.HasSuffix(originalContent, "\n") && !strings.HasSuffix(filteredContent, "\n") {
		filteredContent += "\n"
	}

	// #nosec G306 -- documentation files need to be readable by the server process serving them
	return os.WriteFile(filePath, []byte(filteredContent), 0o644)
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
	// #nosec G306 -- template file needs to be readable by protoc process
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
