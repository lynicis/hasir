package registry

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"connectrpc.com/connect"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/timestamppb"

	"hasir/api/internal/registry"
	registryv1 "hasir/proto/gen/go/registry/v1"
)

func openRepoAndHead(span trace.Span, repoPath string) (*git.Repository, *plumbing.Reference, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		span.RecordError(err)
		return nil, nil, connect.NewError(connect.CodeNotFound, errors.New("failed to open git repository"))
	}

	ref, err := getRepoHead(repo)
	if err != nil {
		span.RecordError(err)
		if errors.Is(err, plumbing.ErrReferenceNotFound) {
			return nil, nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("failed to get repository HEAD: %w", registry.ErrEmptyRepository))
		}
		return nil, nil, connect.NewError(connect.CodeNotFound, errors.New("failed to get repository HEAD"))
	}

	return repo, ref, nil
}

func toProtoCommit(c *object.Commit) *registryv1.Commit {
	return &registryv1.Commit{
		Id:      c.Hash.String(),
		Message: c.Message,
		User: &registryv1.Commit_User{
			Id:       c.Author.Email,
			Username: c.Author.Name,
		},
		CommitedAt: timestamppb.New(c.Author.When),
	}
}

func (r *PgRepository) GetCommits(ctx context.Context, repoPath string, page, pageSize int) ([]*registryv1.Commit, int, error) {
	var span trace.Span
	_, span = r.tracer.Start(ctx, "GetCommits", trace.WithAttributes(
		attribute.String("repoPath", repoPath),
		attribute.Int("page", page),
		attribute.Int("pageSize", pageSize),
	))
	defer span.End()

	repo, ref, err := openRepoAndHead(span, repoPath)
	if err != nil {
		return nil, 0, err
	}

	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		span.RecordError(err)
		return nil, 0, connect.NewError(connect.CodeInternal, errors.New("failed to get commit log"))
	}
	defer commitIter.Close()

	totalCount := 0
	offset := (page - 1) * pageSize
	var commits []*registryv1.Commit

	err = commitIter.ForEach(func(c *object.Commit) error {
		if totalCount >= offset && len(commits) < pageSize {
			commits = append(commits, toProtoCommit(c))
		}
		totalCount++
		return nil
	})
	if err != nil {
		span.RecordError(err)
		return nil, 0, connect.NewError(connect.CodeInternal, errors.New("failed to iterate commits"))
	}

	return commits, totalCount, nil
}

func (r *PgRepository) GetRecentCommit(ctx context.Context, repoPath string) (*registryv1.Commit, error) {
	var span trace.Span
	_, span = r.tracer.Start(ctx, "GetRecentCommit", trace.WithAttributes(
		attribute.String("repoPath", repoPath),
	))
	defer span.End()

	repo, ref, err := openRepoAndHead(span, repoPath)
	if err != nil {
		return nil, err
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to get commit object"))
	}

	return toProtoCommit(commit), nil
}

func (r *PgRepository) GetFileTree(ctx context.Context, repoPath string, subPath *string) (*registryv1.GetFileTreeResponse, error) {
	var span trace.Span
	_, span = r.tracer.Start(ctx, "GetFileTree", trace.WithAttributes(
		attribute.String("repoPath", repoPath),
	))
	defer span.End()

	repo, ref, err := openRepoAndHead(span, repoPath)
	if err != nil {
		return nil, err
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to get commit object"))
	}

	tree, err := commit.Tree()
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to get tree object"))
	}

	targetTree := tree
	targetPath := ""
	if subPath != nil && *subPath != "" {
		targetPath = *subPath
		targetTree, err = tree.Tree(targetPath)
		if err != nil {
			span.RecordError(err)
			return nil, connect.NewError(connect.CodeNotFound, errors.New("path not found in repository"))
		}
	}

	return &registryv1.GetFileTreeResponse{
		Nodes: buildFileTreeNodes(tree, targetTree, targetPath),
	}, nil
}

func buildFileTreeNodes(rootTree, tree *object.Tree, basePath string) []*registryv1.FileTreeNode {
	var nodes []*registryv1.FileTreeNode

	for _, entry := range tree.Entries {
		nodePath := entry.Name
		if basePath != "" {
			nodePath = basePath + "/" + entry.Name
		}

		nodeType := registryv1.NodeType_NODE_TYPE_FILE
		var children []*registryv1.FileTreeNode

		if !entry.Mode.IsFile() {
			nodeType = registryv1.NodeType_NODE_TYPE_DIRECTORY
			if subTree, err := rootTree.Tree(nodePath); err == nil {
				children = buildFileTreeNodes(rootTree, subTree, nodePath)
			}
		}

		nodes = append(nodes, &registryv1.FileTreeNode{
			Name:     entry.Name,
			Path:     nodePath,
			Type:     nodeType,
			Children: children,
		})
	}

	return nodes
}

func (r *PgRepository) GetFilePreview(ctx context.Context, repoPath, filePath string) (*registryv1.GetFilePreviewResponse, error) {
	var span trace.Span
	_, span = r.tracer.Start(ctx, "GetFilePreview", trace.WithAttributes(
		attribute.String("repoPath", repoPath),
		attribute.String("filePath", filePath),
	))
	defer span.End()

	repo, ref, err := openRepoAndHead(span, repoPath)
	if err != nil {
		return nil, err
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to get commit object"))
	}

	tree, err := commit.Tree()
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to get tree object"))
	}

	file, err := tree.File(filePath)
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("file not found in repository"))
	}

	content, err := file.Contents()
	if err != nil {
		span.RecordError(err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to read file content"))
	}

	return &registryv1.GetFilePreviewResponse{
		Content:  content,
		MimeType: detectMimeType(filePath, []byte(content)),
		Size:     file.Size,
	}, nil
}

var mimeTypeMap = map[string]string{
	".txt":        "text/plain",
	".md":         "text/markdown",
	".json":       "application/json",
	".yaml":       "application/yaml",
	".yml":        "application/yaml",
	".xml":        "application/xml",
	".html":       "text/html",
	".htm":        "text/html",
	".css":        "text/css",
	".js":         "application/javascript",
	".ts":         "application/typescript",
	".go":         "text/x-go",
	".py":         "text/x-python",
	".java":       "text/x-java",
	".c":          "text/x-c",
	".cpp":        "text/x-c++",
	".cc":         "text/x-c++",
	".cxx":        "text/x-c++",
	".h":          "text/x-c-header",
	".hpp":        "text/x-c-header",
	".rs":         "text/x-rust",
	".sh":         "application/x-sh",
	".sql":        "application/sql",
	".proto":      "text/x-protobuf",
	".dockerfile": "text/x-dockerfile",
	".Dockerfile": "text/x-dockerfile",
}

func detectMimeType(filePath string, content []byte) string {
	ext := filepath.Ext(filePath)
	if mimeType, ok := mimeTypeMap[ext]; ok {
		return mimeType
	}

	if len(content) > 0 && content[0] == 0 {
		return "application/octet-stream"
	}
	return "text/plain"
}

func getRepoHead(repo *git.Repository) (*plumbing.Reference, error) {
	ref, err := repo.Head()
	if err == nil {
		return ref, nil
	}

	branches, bErr := repo.Branches()
	if bErr != nil {
		return nil, err
	}
	defer branches.Close()

	var candidate *plumbing.Reference
	_ = branches.ForEach(func(b *plumbing.Reference) error {
		if b.Name() == "refs/heads/main" {
			candidate = b
			return storer.ErrStop
		}
		if candidate == nil || b.Name() == "refs/heads/master" {
			candidate = b
		}
		return nil
	})

	if candidate != nil {
		_ = repo.Storer.SetReference(plumbing.NewSymbolicReference(plumbing.HEAD, candidate.Name()))
		return candidate, nil
	}

	return nil, err
}
