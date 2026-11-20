# Tasks

- [x] Add `ResolveBufWorkspace` + `skipDir` (`pkg/sdkgenerator/buf_workspace.go`)
- [x] Buf generator: run in workspace dir, generate `buf.yaml` only when absent, `buf dep update` unless locked, `--template hasir.buf.gen.yaml`
- [x] Documentation generator: use the same workspace resolution and dep-update rule
- [x] `FindProtoFiles`, `FindProtoFilesInBareRepoWithRunner`, and `detectSdksFromRepoCommit` skip dot-dirs, `node_modules`, `vendor`
- [x] Tests: workspace table tests, nested monorepo generation, ambiguous error, untouched `buf.yaml`/`buf.gen.yaml`
- [x] Verified end-to-end against pushed commit `41db29d7c72af1f53bb50fc990cda4df916f196a` (all 5 SDKs + docs generated)
- [ ] Manual: restart `apps/api` (`make dev`) and re-push to confirm queue processing end-to-end
