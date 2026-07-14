// path: packages/shared/src/index.ts
// Re-exports the baseline production hooks documented in ADR-0001.
// Every Bun/TS workspace imports them via `import { logger, otelTracer } from "@hasir/shared"`
// so structured logging and OTEL plumbing are consistent across the platform.
export { logger, type Logger } from "./log.ts";
export { otelTracer, otelMeter, initOtel } from "./otel.ts";
