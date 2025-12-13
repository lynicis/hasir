package registry

import (
	"time"

	"hasir-api/pkg/proto"

	registryv1 "buf.build/gen/go/hasir/hasir/protocolbuffers/go/registry/v1"
)

type RepositoryDTO struct {
	Id             string           `db:"id"`
	Name           string           `db:"name"`
	CreatedBy      string           `db:"created_by"`
	OrganizationId string           `db:"organization_id"`
	Path           string           `db:"path"`
	Visibility     proto.Visibility `db:"visibility"`
	CreatedAt      time.Time        `db:"created_at"`
	UpdatedAt      *time.Time       `db:"updated_at"`
	DeletedAt      *time.Time       `db:"deleted_at"`
}

type SDK string

const (
	SdkGoProtobuf   SDK = "GO_PROTOBUF"
	SdkGoConnectRpc SDK = "GO_CONNECTRPC"
	SdkGoGrpc       SDK = "GO_GRPC"
	SdkJsBufbuildEs SDK = "JS_BUFBUILD_ES"
	SdkJsProtobuf   SDK = "JS_PROTOBUF"
	SdkJsConnectrpc SDK = "JS_CONNECTRPC"
)

var SdkProtoToDbEnum = map[registryv1.SDK]SDK{
	registryv1.SDK_SDK_GO_PROTOBUF:    SdkGoProtobuf,
	registryv1.SDK_SDK_GO_CONNECTRPC:  SdkGoConnectRpc,
	registryv1.SDK_SDK_GO_GRPC:        SdkGoGrpc,
	registryv1.SDK_SDK_JS_BUFBUILD_ES: SdkJsBufbuildEs,
	registryv1.SDK_SDK_JS_PROTOBUF:    SdkJsProtobuf,
	registryv1.SDK_SDK_JS_CONNECTRPC:  SdkJsConnectrpc,
}

var SdkDbToProtoEnum = map[SDK]registryv1.SDK{
	SdkGoProtobuf:   registryv1.SDK_SDK_GO_PROTOBUF,
	SdkGoConnectRpc: registryv1.SDK_SDK_GO_CONNECTRPC,
	SdkGoGrpc:       registryv1.SDK_SDK_GO_GRPC,
	SdkJsBufbuildEs: registryv1.SDK_SDK_JS_BUFBUILD_ES,
	SdkJsProtobuf:   registryv1.SDK_SDK_JS_PROTOBUF,
	SdkJsConnectrpc: registryv1.SDK_SDK_JS_CONNECTRPC,
}

type SdkPreferencesDTO struct {
	Id           string     `db:"id"`
	RepositoryId string     `db:"repository_id"`
	Sdk          SDK        `db:"sdk"`
	Status       bool       `db:"status"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    *time.Time `db:"updated_at"`
}

type SshOperation string

const (
	SshOperationRead  SshOperation = "read"
	SshOperationWrite SshOperation = "write"
)

type SdkGenerationJobStatus string

const (
	SdkGenerationJobStatusPending    SdkGenerationJobStatus = "pending"
	SdkGenerationJobStatusProcessing SdkGenerationJobStatus = "processing"
	SdkGenerationJobStatusCompleted  SdkGenerationJobStatus = "completed"
	SdkGenerationJobStatusFailed     SdkGenerationJobStatus = "failed"
)

type SdkGenerationJobDTO struct {
	Id           string                 `db:"id"`
	RepositoryId string                 `db:"repository_id"`
	CommitHash   string                 `db:"commit_hash"`
	Sdk          SDK                    `db:"sdk"`
	Status       SdkGenerationJobStatus `db:"status"`
	Attempts     int                    `db:"attempts"`
	MaxAttempts  int                    `db:"max_attempts"`
	CreatedAt    time.Time              `db:"created_at"`
	ProcessedAt  *time.Time             `db:"processed_at"`
	CompletedAt  *time.Time             `db:"completed_at"`
	ErrorMessage *string                `db:"error_message"`
}

type SdkTriggerJobDTO struct {
	Id           string                 `db:"id"`
	RepositoryId string                 `db:"repository_id"`
	RepoPath     string                 `db:"repo_path"`
	Status       SdkGenerationJobStatus `db:"status"`
	Attempts     int                    `db:"attempts"`
	MaxAttempts  int                    `db:"max_attempts"`
	CreatedAt    time.Time              `db:"created_at"`
	ProcessedAt  *time.Time             `db:"processed_at"`
	CompletedAt  *time.Time             `db:"completed_at"`
	ErrorMessage *string                `db:"error_message"`
}
