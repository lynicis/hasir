# Design: buf workspace discovery

## Resolution order (`ResolveBufWorkspace`)
1. Root has `buf.yaml` or `buf.work.yaml`: the workspace is the root.
2. Otherwise, walk the checkout, skipping `.`-prefixed dirs, `node_modules`, and `vendor`. Collect the outermost dirs that contain a buf config; nested modules belong to their outer workspace.
   - Exactly one: use it.
   - Two or more: fail with `ErrAmbiguousBufWorkspace`, listing the paths.
   - None: use the root, with a generated `buf.yaml` (previous behavior).

## Rules
- An existing `buf.yaml` is authoritative and is never rewritten. The previous unmarshal/re-marshal dropped `excludes`, `lint`, `breaking`, and user deps.
- `buf dep update` replaces the deprecated `buf mod update`. It is skipped when the repository ships both a config and a `buf.lock`, which keeps builds reproducible.
- The plugin template is written as `hasir.buf.gen.yaml` and passed via `--template`, so the repository's `buf.gen.yaml` is untouched.
- `FindProtoFiles` skips the same directories as discovery.
