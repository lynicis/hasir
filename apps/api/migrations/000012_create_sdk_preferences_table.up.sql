-- Create SDK enum type
CREATE TYPE sdk_type AS ENUM (
    'GO_PROTOBUF',
    'GO_CONNECTRPC',
    'GO_GRPC',
    'JS_BUFBUILD_ES',
    'JS_PROTOBUF',
    'JS_CONNECTRPC'
);

CREATE TABLE IF NOT EXISTS sdk_preferences (
    id VARCHAR(36) PRIMARY KEY,
    repository_id VARCHAR(36) NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    sdk sdk_type NOT NULL,
    status BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(repository_id, sdk)
);

CREATE INDEX idx_sdk_preferences_repository_id ON sdk_preferences(repository_id);
CREATE INDEX idx_sdk_preferences_sdk ON sdk_preferences(sdk);
