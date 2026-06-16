package sdkgenerator

import (
	"context"
	"fmt"
	"path/filepath"
)

type bufGenerator struct {
	baseGenerator
	runner        CommandRunner
	plugins       []BufGenPlugin
	customModules map[string]string
}

func NewBufGenerator(sdk SDK, dirName string, runner CommandRunner, plugins []BufGenPlugin, customModules map[string]string) *bufGenerator {
	return &bufGenerator{
		baseGenerator: baseGenerator{
			sdk:     sdk,
			dirName: dirName,
			runner:  runner,
		},
		runner:        runner,
		plugins:       plugins,
		customModules: customModules,
	}
}

func (g *bufGenerator) Generate(ctx context.Context, input GeneratorInput) (*GeneratorOutput, error) {
	if err := g.Validate(input); err != nil {
		return nil, err
	}

	imports, err := ParseProtoImportsFromFiles(input.RepoPath, input.ProtoFiles)
	if err != nil {
		return nil, fmt.Errorf("failed to parse proto imports: %w", err)
	}

	modules := ResolveBsrModulesWithCustom(imports, g.customModules)

	if err := GenerateBufYaml(input.RepoPath, modules); err != nil {
		return nil, fmt.Errorf("failed to generate buf.yaml: %w", err)
	}

	outputPath := filepath.Clean(input.OutputPath)
	plugins := make([]BufGenPlugin, len(g.plugins))
	for i, p := range g.plugins {
		plugins[i] = BufGenPlugin{
			Remote: p.Remote,
			Out:    outputPath,
			Opt:    p.Opt,
			Opts:   p.Opts,
		}
	}

	if err := GenerateBufGenYaml(input.RepoPath, plugins, input.GoPackagePrefix); err != nil {
		return nil, fmt.Errorf("failed to generate buf.gen.yaml: %w", err)
	}

	if _, err := g.runner.Run(ctx, "buf", []string{"mod", "update"}, input.RepoPath); err != nil {
		return nil, fmt.Errorf("buf mod update failed: %w", err)
	}

	if _, err := g.runner.Run(ctx, "buf", []string{"generate"}, input.RepoPath); err != nil {
		return nil, fmt.Errorf("buf generate failed: %w", err)
	}

	return &GeneratorOutput{
		OutputPath: input.OutputPath,
		FilesCount: len(input.ProtoFiles),
	}, nil
}
