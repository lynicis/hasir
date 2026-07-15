FROM oven/bun:1-alpine AS build
WORKDIR /app

# Copy root package files and all workspace member package.jsons
# so bun install can resolve workspace references
COPY package.json bun.lock ./
COPY apps/dashboard/package.json apps/dashboard/package.json
COPY apps/api/package.json apps/api/package.json
COPY packages/config/package.json packages/config/package.json
COPY packages/shared/package.json packages/shared/package.json
COPY packages/tooling/package.json packages/tooling/package.json

RUN echo "@buf:registry=https://buf.build/gen/npm/v1/" > .npmrc
RUN bun install --frozen-lockfile

COPY . .

RUN bunx turbo run build --filter=hasir-dashboard

FROM node:24-alpine AS release
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
