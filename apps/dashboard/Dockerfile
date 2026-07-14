FROM oven/bun:1-alpine AS build
WORKDIR /app

COPY package.json bun.lock ./

RUN echo "@buf:registry=https://buf.build/gen/npm/v1" > .npmrc
RUN bun install --frozen-lockfile

COPY . .

RUN bun run build

FROM node:24-alpine AS release
WORKDIR /app

RUN adduser --system --uid 1001 hasir

COPY --chown=hasir:hasir --from=build /app/.next/standalone ./
COPY --chown=hasir:hasir --from=build /app/.next/static ./.next/static
COPY --chown=hasir:hasir --from=build /app/public ./public

USER hasir

EXPOSE 3000
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

CMD ["node", ".next/standalone/server.js"]
