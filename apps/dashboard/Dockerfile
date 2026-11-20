FROM oven/bun:1.4-alpine AS build
WORKDIR /app

# Copy root package files and all workspace member package.jsons
# so bun install can resolve workspace references
COPY package.json bun.lock ./
COPY apps/api/package.json apps/api/package.json
COPY apps/dashboard/package.json apps/dashboard/package.json
COPY apps/landing/package.json apps/landing/package.json
COPY deploy/helm/package.json deploy/helm/package.json
COPY packages/eslint-config/package.json packages/eslint-config/package.json
COPY packages/proto/package.json packages/proto/package.json
COPY packages/shared/package.json packages/shared/package.json
COPY packages/tooling/package.json packages/tooling/package.json
COPY packages/typescript-config/package.json packages/typescript-config/package.json
COPY packages/ui/package.json packages/ui/package.json

RUN echo "@buf:registry=https://buf.build/gen/npm/v1/" > .npmrc
RUN bun install --frozen-lockfile

COPY . .

RUN bunx turbo run build --filter=hasir-dashboard --only

FROM node:26-alpine AS release
WORKDIR /app

RUN adduser --system --uid 1001 hasir

COPY --chown=hasir:hasir --from=build /app/apps/dashboard/.next/standalone ./
COPY --chown=hasir:hasir --from=build /app/apps/dashboard/.next/static ./.next/static
COPY --chown=hasir:hasir --from=build /app/apps/dashboard/public ./public

USER hasir

EXPOSE 3000
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

CMD ["node", "apps/dashboard/server.js"]
