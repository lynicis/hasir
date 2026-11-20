package sdkgenerator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type DocumentationGenerator struct {
	*protocGenerator
}

func NewDocumentationGenerator(runner CommandRunner) *DocumentationGenerator {
	return &DocumentationGenerator{
		protocGenerator: newProtocGenerator("", "docs", runner, buildDocumentationArgs),
	}
}

func (g *DocumentationGenerator) Generate(ctx context.Context, input GeneratorInput) (*GeneratorOutput, error) {
	ws, err := ResolveBufWorkspace(input.RepoPath)
	if err != nil {
		return nil, err
	}

	var output *GeneratorOutput
	if ws.HasConfig {
		output, err = g.generateWithBuf(ctx, input, ws)
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

func (g *DocumentationGenerator) generateWithBuf(ctx context.Context, input GeneratorInput, ws BufWorkspace) (*GeneratorOutput, error) {
	docTemplatePath := filepath.Join(input.OutputPath, "buf.gen.doc.yaml")
	var sb strings.Builder
	sb.WriteString("version: v2\n")
	sb.WriteString("plugins:\n")
	sb.WriteString("  - local: protoc-gen-doc\n")
	sb.WriteString("    out: ")
	sb.WriteString(filepath.Clean(input.OutputPath))
	sb.WriteString("\n")
	sb.WriteString("    strategy: all\n")
	sb.WriteString("    opt:\n")
	sb.WriteString("      - markdown,index.md\n")

	// #nosec G306 -- template needs to be readable by buf process
	if err := os.WriteFile(docTemplatePath, []byte(sb.String()), 0o644); err != nil {
		return nil, fmt.Errorf("failed to write buf.gen.doc.yaml: %w", err)
	}
	defer func() {
		_ = os.Remove(docTemplatePath)
	}()

	if err := runBufDepUpdate(ctx, g.runner, ws); err != nil {
		return nil, err
	}

	if _, err := g.runner.Run(ctx, "buf", []string{"generate", "--template", docTemplatePath}, ws.Dir); err != nil {
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

		if strings.EqualFold(trimmed, "- [Scalar Value Types](#scalar-value-types)") {
			continue
		}

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
	args := []string{
		"--proto_path=" + filepath.Clean(input.RepoPath),
		"--doc_out=" + filepath.Clean(input.OutputPath),
		"--doc_opt=markdown,index.md",
	}
	return append(args, input.ProtoFiles...)
}
