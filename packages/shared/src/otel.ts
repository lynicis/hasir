// path: packages/shared/src/otel.ts
// OTEL bootstrap helper for Bun/TS apps. The dashboard already wires
// @vercel/otel directly in apps/dashboard/instrumentation.ts; this helper
// is the platform-level fallback for any future non-Next.js Bun app
// that wants tracing without re-implementing the boilerplate.
//
// noop by default — keeps the shared package zero-dependency at build time.
// Enable via `initOtel({ serviceName })` from the app's entrypoint.

export interface OtelConfig {
  serviceName: string;
  otlpEndpoint?: string;
}

export let otelTracer: unknown = null;
export let otelMeter: unknown = null;

export function initOtel(_config: OtelConfig): void {
  // Stub until @hasir/shared pulls in @opentelemetry/sdk-trace-base.
  // The dashboard uses @vercel/otel directly; other Bun apps will opt in.
  otelTracer = null;
  otelMeter = null;
}
