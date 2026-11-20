package sdkgenerator

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

var ErrAmbiguousBufWorkspace = errors.New("multiple buf workspaces found")

// skipDir reports directories that are never scanned for protos or buf configs.
func skipDir(name string) bool {
	return strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor"
}

// IsSkippedPath reports whether any parent directory segment in a slash- or OS-separated
// relative path should be skipped when scanning repositories.
func IsSkippedPath(relPath string) bool {
	dir := path.Dir(filepath.ToSlash(relPath))
	if dir == "." || dir == "/" {
		return false
	}
	for _, seg := range strings.Split(dir, "/") {
		if seg != "" && skipDir(seg) {
			return true
		}
	}
	return false
}

type BufWorkspace struct {
	// Dir is the absolute directory buf commands run in.
	Dir string
	// HasConfig is true when the repository ships its own buf.yaml or buf.work.yaml, which must not be rewritten.
	HasConfig bool
	HasLock   bool
	// DepUpdateDirs lists relative module directories inside a buf.work.yaml workspace that lack a buf.lock.
	DepUpdateDirs []string
}

type bufWorkConfig struct {
	Directories []string `yaml:"directories"`
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

func hasBufConfig(dir string) bool {
	for _, name := range []string{"buf.yaml", "buf.work.yaml"} {
		if fileExists(filepath.Join(dir, name)) {
			return true
		}
	}
	return false
}

// ResolveBufWorkspace picks the buf workspace directory inside repoPath: the root when it has a buf config,
// otherwise the single outermost directory containing one. Without any config the root is returned.
func ResolveBufWorkspace(repoPath string) (BufWorkspace, error) {
	root, err := filepath.Abs(repoPath)
	if err != nil {
		return BufWorkspace{}, fmt.Errorf("failed to resolve repository path: %w", err)
	}

	dir, hasConfig := root, hasBufConfig(root)
	if !hasConfig {
		var candidates []string
		walkErr := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() || p == root {
				return nil
			}
			if skipDir(d.Name()) {
				return filepath.SkipDir
			}
			if hasBufConfig(p) {
				candidates = append(candidates, p)
				// Nested modules belong to this outer workspace.
				return filepath.SkipDir
			}
			return nil
		})
		if walkErr != nil {
			return BufWorkspace{}, fmt.Errorf("failed to scan for buf workspaces: %w", walkErr)
		}

		switch len(candidates) {
		case 0:
		case 1:
			dir, hasConfig = candidates[0], true
		default:
			rel := make([]string, 0, len(candidates))
			for _, c := range candidates {
				r, _ := filepath.Rel(root, c)
				rel = append(rel, r)
			}
			return BufWorkspace{}, fmt.Errorf(
				"%w: %s; add a root buf.yaml (v2 modules) or buf.work.yaml",
				ErrAmbiguousBufWorkspace, strings.Join(rel, ", "),
			)
		}
	}

	if hasConfig && !fileExists(filepath.Join(dir, "buf.yaml")) && fileExists(filepath.Join(dir, "buf.work.yaml")) {
		unlockedDirs, hasWorkDirs, err := resolveBufWorkUnlockedDirs(dir)
		if err != nil {
			return BufWorkspace{}, err
		}
		if hasWorkDirs {
			return BufWorkspace{
				Dir:           dir,
				HasConfig:     true,
				HasLock:       len(unlockedDirs) == 0,
				DepUpdateDirs: unlockedDirs,
			}, nil
		}
	}

	return BufWorkspace{Dir: dir, HasConfig: hasConfig, HasLock: fileExists(filepath.Join(dir, "buf.lock"))}, nil
}

func resolveBufWorkUnlockedDirs(dir string) ([]string, bool, error) {
	workPath := filepath.Join(dir, "buf.work.yaml")
	// #nosec G304 -- workPath is resolved within the validated checkout directory
	content, err := os.ReadFile(workPath)
	if err != nil {
		return nil, false, fmt.Errorf("failed to read buf.work.yaml: %w", err)
	}

	var cfg bufWorkConfig
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return nil, false, fmt.Errorf("failed to parse buf.work.yaml: %w", err)
	}
	if len(cfg.Directories) == 0 {
		return nil, false, nil
	}

	var unlocked []string
	for _, entry := range cfg.Directories {
		clean := filepath.Clean(entry)
		if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || strings.HasPrefix(clean, "-") {
			return nil, false, fmt.Errorf("invalid directory %q in buf.work.yaml", entry)
		}
		if !fileExists(filepath.Join(dir, clean, "buf.lock")) {
			unlocked = append(unlocked, filepath.ToSlash(clean))
		}
	}

	return unlocked, true, nil
}
