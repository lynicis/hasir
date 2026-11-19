package registry

import (
	"context"
	"errors"
	"path/filepath"

	"go.uber.org/zap"

	"hasir/api/pkg/authentication"
	"hasir/api/pkg/authorization"
	"hasir/api/pkg/pagination"
	"hasir/api/pkg/sdkgenerator"
	registryv1 "hasir/proto/gen/go/registry/v1"
)

func (s *service) authorizeRepoMember(ctx context.Context, repoId string) (string, *RepositoryDTO, error) {
	userId, err := authentication.MustGetUserID(ctx)
	if err != nil {
		return "", nil, err
	}

	repo, err := s.repository.GetRepositoryById(ctx, repoId)
	if err != nil {
		return "", nil, err
	}

	if err := authorization.IsUserMember(ctx, s.orgRepo, repo.OrganizationId, userId); err != nil {
		return "", nil, err
	}

	return userId, repo, nil
}

func (s *service) GetCommits(
	ctx context.Context,
	req *registryv1.GetCommitsRequest,
) (*registryv1.GetCommitsResponse, error) {
	_, repo, err := s.authorizeRepoMember(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	page, pageSize := pagination.Normalize(req.Pagination)

	commits, totalCount, err := s.repository.GetCommits(ctx, repo.Path, page, pageSize)
	if err != nil {
		return nil, err
	}

	nextPage, totalPages, err := pagination.ComputePageMeta(totalCount, page, pageSize)
	if err != nil {
		return nil, err
	}

	return &registryv1.GetCommitsResponse{
		Commits:   commits,
		NextPage:  nextPage,
		TotalPage: totalPages,
	}, nil
}

func (s *service) GetRecentCommit(
	ctx context.Context,
	req *registryv1.GetRecentCommitRequest,
) (*registryv1.Commit, error) {
	_, repo, err := s.authorizeRepoMember(ctx, req.GetRepositoryId())
	if err != nil {
		return nil, err
	}

	return s.repository.GetRecentCommit(ctx, repo.Path)
}

func (s *service) GetFileTree(
	ctx context.Context,
	req *registryv1.GetFileTreeRequest,
) (*registryv1.GetFileTreeResponse, error) {
	_, repo, err := s.authorizeRepoMember(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	var subPath *string
	if req.GetPath() != "" {
		path := req.GetPath()
		subPath = &path
	}

	return s.repository.GetFileTree(ctx, repo.Path, subPath)
}

func (s *service) GetFilePreview(
	ctx context.Context,
	req *registryv1.GetFilePreviewRequest,
) (*registryv1.GetFilePreviewResponse, error) {
	_, repo, err := s.authorizeRepoMember(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	return s.repository.GetFilePreview(ctx, repo.Path, req.GetPath())
}

func (s *service) ValidateSshAccess(
	ctx context.Context,
	userId, repoPath string,
	operation SshOperation,
) (bool, error) {
	repoId := filepath.Base(repoPath)
	repo, err := s.repository.GetRepositoryById(ctx, repoId)
	if err != nil {
		return false, err
	}

	role, err := s.orgRepo.GetMemberRole(ctx, repo.OrganizationId, userId)
	if err != nil {
		zap.L().Warn("SSH access denied: user not member of organization",
			zap.String("userId", userId),
			zap.String("repoPath", repoPath),
			zap.String("organizationId", repo.OrganizationId),
		)
		return false, nil
	}

	switch operation {
	case SshOperationWrite:
		if role == authorization.MemberRoleOwner || role == authorization.MemberRoleAuthor {
			zap.L().Info("SSH write access granted",
				zap.String("userId", userId),
				zap.String("repoPath", repoPath),
				zap.String("role", string(role)),
			)
			return true, nil
		}
		zap.L().Warn("SSH write access denied: insufficient role",
			zap.String("userId", userId),
			zap.String("repoPath", repoPath),
			zap.String("role", string(role)),
		)
		return false, nil
	case SshOperationRead:
		zap.L().Info("SSH read access granted",
			zap.String("userId", userId),
			zap.String("repoPath", repoPath),
			zap.String("role", string(role)),
		)
		return true, nil
	default:
		return false, errors.New("unknown SSH operation")
	}
}

func (s *service) HasProtoFiles(ctx context.Context, repoPath, commitHash string) (bool, error) {
	protoFiles, err := sdkgenerator.FindProtoFilesInBareRepoWithRunner(ctx, s.commandRunner(), repoPath, commitHash)
	if err != nil {
		return false, err
	}
	return len(protoFiles) > 0, nil
}
