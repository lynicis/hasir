# Graph Report - hasir  (2026-09-23)

## Corpus Check
- Large corpus: 781 files · ~458,085 words. Semantic extraction will be expensive (many Claude tokens). Consider running on a subfolder.

## Summary
- 4151 nodes · 10419 edges · 1 communities
- Extraction: 96% EXTRACTED · 4% INFERRED · 0% AMBIGUOUS · INFERRED: 411 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- Entire Codebase

## God Nodes (most connected - your core abstractions)
1. `cn()` - 366 edges
2. `@testing-library/react` - 49 edges
3. `NewMockRepository()` - 43 edges
4. `Button()` - 43 edges
5. `NewMockCommandRunner()` - 40 edges
6. `@base-ui/react` - 39 edges
7. `useClient()` - 36 edges
8. `Service` - 35 edges
9. `CommandRunner` - 35 edges
10. `@connectrpc/connect` - 32 edges

## Surprising Connections (you probably didn't know these)
- `TestHandler_CreateOrganization()` --calls--> `NewOrganizationServiceClient()`  [EXTRACTED]
  apps/api/internal/organization/handler_test.go → packages/proto/gen/go/organization/v1/organizationv1connect/organization.connect.go
- `TestHandler_GetOrganizations()` --calls--> `NewOrganizationServiceClient()`  [EXTRACTED]
  apps/api/internal/organization/handler_test.go → packages/proto/gen/go/organization/v1/organizationv1connect/organization.connect.go
- `TestHandler_DeleteOrganization()` --calls--> `NewOrganizationServiceClient()`  [EXTRACTED]
  apps/api/internal/organization/handler_test.go → packages/proto/gen/go/organization/v1/organizationv1connect/organization.connect.go
- `TestHandler_UpdateOrganization()` --calls--> `NewOrganizationServiceClient()`  [EXTRACTED]
  apps/api/internal/organization/handler_test.go → packages/proto/gen/go/organization/v1/organizationv1connect/organization.connect.go
- `TestHandler_InviteMember()` --calls--> `NewOrganizationServiceClient()`  [EXTRACTED]
  apps/api/internal/organization/handler_test.go → packages/proto/gen/go/organization/v1/organizationv1connect/organization.connect.go

## Import Cycles
- None detected.

## Communities (1 total, 0 thin omitted)

### Community 0 - "Entire Codebase"
Cohesion: 0.00
Nodes (2554): go-lsp script, PATH, ts-lsp script, validate-chart.sh script, success(), warning(), error(), $schema (+2546 more)

## Knowledge Gaps
- **1108 isolated node(s):** `$schema`, `changelog`, `commit`, `fixed`, `linked` (+1103 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 1602 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **What connects `$schema`, `changelog`, `commit` to the rest of the system?**
  _1108 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Entire Codebase` be split into smaller, more focused modules?**
  _Cohesion score 0.0012089407981238371 - nodes in this community are weakly interconnected._