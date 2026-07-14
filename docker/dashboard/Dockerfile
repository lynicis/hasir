# path: docker/dashboard/Dockerfile
# Build context is the repository root so Turborepo and workspace deps resolve.
# See docker/base/Dockerfile.node for the base builder image.

FROM hasir-node-base:latest AS builder
WORKDIR /app

# Run turbo build for the dashboard. This also triggers the `proto` task 
# because dashboard^build dependsOn proto.
RUN bunx turbo run build --filter=hasir-dashboard

# ---
FROM node:22-alpine AS release
WORKDIR /app

RUN adduser --system --uid 1001 hasir

# Next.js standalone output tracing captures the monorepo path structure
COPY --chown=hasir:hasir --from=builder /app/apps/dashboard/.next/standalone ./
COPY --chown=hasir:hasir --from=builder /app/apps/dashboard/.next/static ./apps/dashboard/.next/static
COPY --chown=hasir:hasir --from=builder /app/apps/dashboard/public ./apps/dashboard/public

USER hasir
EXPOSE 3000
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

CMD ["node", "apps/dashboard/server.js"]
