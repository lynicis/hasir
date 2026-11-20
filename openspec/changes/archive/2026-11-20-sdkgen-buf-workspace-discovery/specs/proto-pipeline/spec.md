## ADDED Requirements

### Requirement: SDK Generation Honors Repository Buf Workspace

SDK and documentation generation SHALL run buf in the repository's own buf workspace and SHALL NOT rewrite an existing `buf.yaml` or `buf.gen.yaml`.

#### Scenario: Monorepo with nested buf workspace

- **WHEN** a pushed commit has no root buf config and exactly one `buf.yaml` outside `.`-prefixed, `node_modules`, and `vendor` dirs
- **THEN** buf commands run in that directory
- **AND** the existing `buf.yaml` is byte-for-byte unchanged

#### Scenario: Locked dependencies

- **WHEN** the workspace ships its own buf config and `buf.lock`
- **THEN** `buf dep update` is not run

#### Scenario: Ambiguous workspaces

- **WHEN** two or more sibling buf workspaces exist and no root config
- **THEN** generation fails with an error listing each workspace path
