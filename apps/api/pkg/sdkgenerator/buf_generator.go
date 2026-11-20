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

	ws, err := ResolveBufWorkspace(input.RepoPath)
	if err != nil {
		return nil, err
	}

	if !ws.HasConfig {
		imports, err := ParseProtoImportsFromFiles(input.RepoPath, input.ProtoFiles)
		if err != nil {
			return nil, fmt.Errorf("failed to parse proto imports: %w", err)
		}

		if err := GenerateBufYaml(ws.Dir, ResolveBsrModulesWithCustom(imports, g.customModules)); err != nil {
			return nil, fmt.Errorf("failed to generate buf.yaml: %w", err)
		}
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

	if err := GenerateBufGenYaml(ws.Dir, plugins, input.GoPackagePrefix); err != nil {
		return nil, fmt.Errorf("failed to generate %s: %w", BufGenTemplateName, err)
	}

	if err := runBufDepUpdate(ctx, g.runner, ws); err != nil {
		return nil, err
	}

	templatePath := filepath.Join(ws.Dir, BufGenTemplateName)
	if _, err := g.runner.Run(ctx, "buf", []string{"generate", "--template", templatePath}, ws.Dir); err != nil {
		return nil, fmt.Errorf("buf generate failed: %w", err)
	}

	return &GeneratorOutput{
		OutputPath: input.OutputPath,
		FilesCount: len(input.ProtoFiles),
	}, nil
}

// runBufDepUpdate resolves BSR deps unless the repository already pins them in its own buf.lock.
func runBufDepUpdate(ctx context.Context, runner CommandRunner, ws BufWorkspace) error {
	if ws.HasConfig && ws.HasLock {
		return nil
	}

	if len(ws.DepUpdateDirs) > 0 {
		for _, dir := range ws.DepUpdateDirs {
			if _, err := runner.Run(ctx, "buf", []string{"dep", "update", dir}, ws.Dir); err != nil {
				return fmt.Errorf("buf dep update %s failed: %w", dir, err)
			}
		}
		return nil
	}

	if _, err := runner.Run(ctx, "buf", []string{"dep", "update"}, ws.Dir); err != nil {
		return fmt.Errorf("buf dep update failed: %w", err)
	}

	return nil
}
