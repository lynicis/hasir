# path: Makefile
# Top-level developer workflow. The Makefile is a thin convenience layer over
# the `turbo`/`bun` commands already exposed in package.json so newcomers and
# muscle-memory contributors have a familiar entry point. Commands compose:
#   make setup && make proto && make dev
#
# Everything here has a real subprocess under package.json or scripts/. The
# Makefile contains no orchestration logic of its own (one of ADR-0001's
# rules — single source of truth lives in turbo.json).

.PHONY: setup dev build test lint typecheck proto docker helm-lint clean release

setup:           # install deps + generate proto + update helm deps
	bun install
	bun run proto
	helm dependency update deploy/helm/hasir >/dev/null 2>&1 || true

dev:            # API + dashboard in parallel via turbo
	bun run dev

build:          # build everything affected
	bun run build --affected

test:           # run affected tests
	bun run test --affected

lint:           # eslint across JS workspaces, golangci-lint across Go services
	bun run lint

typecheck:      # tsc across JS workspaces
	bun run typecheck

proto:          # buf generate, writing into proto/gen/ and consumers
	bun run proto

docker:         # buildx + bake all images
	./scripts/docker-build.sh

helm-lint:      # helm lint + template validation
	./scripts/helm-lint.sh

clean:          # nuke across the tree
	bun run clean
	rm -rf tmp/ coverage/ junit.xml

release:        # path-scoped tag + GH release; usage: make release app=api bump=patch
	./scripts/release.sh "$(app)" "$(bump)"
