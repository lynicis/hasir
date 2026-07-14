package sdkgenerator

import (
	"context"
	"embed"
	"fmt"
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
	bufYamlPath := filepath.Join(input.RepoPath, "buf.yaml")
	var output *GeneratorOutput
	var err error

	if _, statErr := os.Stat(bufYamlPath); statErr == nil {
		output, err = g.generateWithBuf(ctx, input)
	} else {
		output, err = g.protocGenerator.Generate(ctx, input)
	}

	if err != nil {
		return nil, err
	}

	indexMdPath := filepath.Join(input.OutputPath, "index.md")
	if _, err := os.Stat(indexMdPath); err == nil {
		if err := removeScalarValueTypesSection(indexMdPath); err != nil {
			return output, err
		}
	}

	return output, nil
}

func (g *DocumentationGenerator) generateWithBuf(ctx context.Context, input GeneratorInput) (*GeneratorOutput, error) {
	bufGenYamlPath := filepath.Join(input.RepoPath, "buf.gen.yaml")
	var generatedBufGenYaml bool

	if _, err := os.Stat(bufGenYamlPath); os.IsNotExist(err) {
		// 1. Write the template file
		templateContent, err := templateFS.ReadFile("template/proto-doc-tempate.mustache")
		var templatePath string
		if err == nil {
			templatePath = filepath.Join(input.OutputPath, "proto-doc-template.mustache")
			// #nosec G306 -- template file needs to be readable by protoc/buf process
			if err := os.WriteFile(templatePath, templateContent, 0o644); err != nil {
				templatePath = "" // Fallback to no template if write fails
			}
		}

		// 2. Generate buf.gen.yaml
		var sb strings.Builder
		sb.WriteString("version: v2\n")
		sb.WriteString("plugins:\n")
		sb.WriteString("  - local: protoc-gen-doc\n")
		sb.WriteString("    out: ")
		sb.WriteString(filepath.Clean(input.OutputPath))
		sb.WriteString("\n")
		if templatePath != "" {
			sb.WriteString("    opt:\n")
			sb.WriteString("      - ")
			sb.WriteString(templatePath)
			sb.WriteString(",index.md\n")
		} else {
			sb.WriteString("    opt:\n")
			sb.WriteString("      - markdown,index.md\n")
		}

		// #nosec G306 -- buf.gen.yaml needs to be readable by buf CLI
		if err := os.WriteFile(bufGenYamlPath, []byte(sb.String()), 0o644); err != nil {
			return nil, fmt.Errorf("failed to write buf.gen.yaml: %w", err)
		}
		generatedBufGenYaml = true
	}

	if generatedBufGenYaml {
		defer func() {
			_ = os.Remove(bufGenYamlPath)
		}()
	}

	// 3. Run buf mod update to download BSR dependencies
	if _, err := g.runner.Run(ctx, "buf", []string{"mod", "update"}, input.RepoPath); err != nil {
		return nil, fmt.Errorf("buf mod update failed: %w", err)
	}

	// 4. Run buf generate
	if _, err := g.runner.Run(ctx, "buf", []string{"generate"}, input.RepoPath); err != nil {
		return nil, fmt.Errorf("buf generate failed: %w", err)
	}

	return &GeneratorOutput{
		OutputPath: input.OutputPath,
		FilesCount: len(input.ProtoFiles),
	}, nil
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

	// #nosec G306 G703 -- documentation files need to be readable by the server process serving them
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
		"--doc_opt=" + templatePath + ",index.md",
	}
	args = append(args, input.ProtoFiles...)

	return args
}
