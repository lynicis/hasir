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
	ManagedByBuf   bool             `db:"managed_by_buf"`
	CreatedAt      time.Time        `db:"created_at"`
	UpdatedAt      *time.Time       `db:"updated_at"`
	DeletedAt      *time.Time       `db:"deleted_at"`
}

type SDK string

const (
	SdkGoProtobuf       SDK = "GO_PROTOBUF"
	SdkGoConnectRpc     SDK = "GO_CONNECTRPC"
	SdkGoGrpc           SDK = "GO_GRPC"
	SdkJsBufbuildEs     SDK = "JS_BUFBUILD_ES"
	SdkJsProtobuf       SDK = "JS_PROTOBUF"
	SdkJsConnectrpc     SDK = "JS_CONNECTRPC"
	SdkRustProtobuf     SDK = "RUST_PROTOBUF"
	SdkRustConnectrpc   SDK = "RUST_CONNECTRPC"
	SdkRustGrpc         SDK = "RUST_GRPC"
	SdkJavaProtobuf     SDK = "JAVA_PROTOBUF"
	SdkJavaConnectrpc   SDK = "JAVA_CONNECTRPC"
	SdkJavaGrpc         SDK = "JAVA_GRPC"
	SdkCsharpProtobuf   SDK = "CSHARP_PROTOBUF"
	SdkCsharpConnectrpc SDK = "CSHARP_CONNECTRPC"
	SdkCsharpGrpc       SDK = "CSHARP_GRPC"
)

var SdkProtoToDbEnum = map[registryv1.SDK]SDK{
	registryv1.SDK_SDK_GO_PROTOBUF:       SdkGoProtobuf,
	registryv1.SDK_SDK_GO_CONNECTRPC:     SdkGoConnectRpc,
	registryv1.SDK_SDK_GO_GRPC:           SdkGoGrpc,
	registryv1.SDK_SDK_JS_BUFBUILD_ES:    SdkJsBufbuildEs,
	registryv1.SDK_SDK_JS_PROTOBUF:       SdkJsProtobuf,
	registryv1.SDK_SDK_JS_CONNECTRPC:     SdkJsConnectrpc,
	registryv1.SDK_SDK_RUST_PROTOBUF:     SdkRustProtobuf,
	registryv1.SDK_SDK_RUST_CONNECTRPC:   SdkRustConnectrpc,
	registryv1.SDK_SDK_RUST_GRPC:         SdkRustGrpc,
	registryv1.SDK_SDK_JAVA_PROTOBUF:     SdkJavaProtobuf,
	registryv1.SDK_SDK_JAVA_CONNECTRPC:   SdkJavaConnectrpc,
	registryv1.SDK_SDK_JAVA_GRPC:         SdkJavaGrpc,
	registryv1.SDK_SDK_CSHARP_PROTOBUF:   SdkCsharpProtobuf,
	registryv1.SDK_SDK_CSHARP_CONNECTRPC: SdkCsharpConnectrpc,
	registryv1.SDK_SDK_CSHARP_GRPC:       SdkCsharpGrpc,
}

var SdkDbToProtoEnum = map[SDK]registryv1.SDK{
	SdkGoProtobuf:       registryv1.SDK_SDK_GO_PROTOBUF,
	SdkGoConnectRpc:     registryv1.SDK_SDK_GO_CONNECTRPC,
	SdkGoGrpc:           registryv1.SDK_SDK_GO_GRPC,
	SdkJsBufbuildEs:     registryv1.SDK_SDK_JS_BUFBUILD_ES,
	SdkJsProtobuf:       registryv1.SDK_SDK_JS_PROTOBUF,
	SdkJsConnectrpc:     registryv1.SDK_SDK_JS_CONNECTRPC,
	SdkRustProtobuf:     registryv1.SDK_SDK_RUST_PROTOBUF,
	SdkRustConnectrpc:   registryv1.SDK_SDK_RUST_CONNECTRPC,
	SdkRustGrpc:         registryv1.SDK_SDK_RUST_GRPC,
	SdkJavaProtobuf:     registryv1.SDK_SDK_JAVA_PROTOBUF,
	SdkJavaConnectrpc:   registryv1.SDK_SDK_JAVA_CONNECTRPC,
	SdkJavaGrpc:         registryv1.SDK_SDK_JAVA_GRPC,
	SdkCsharpProtobuf:   registryv1.SDK_SDK_CSHARP_PROTOBUF,
	SdkCsharpConnectrpc: registryv1.SDK_SDK_CSHARP_CONNECTRPC,
	SdkCsharpGrpc:       registryv1.SDK_SDK_CSHARP_GRPC,
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
