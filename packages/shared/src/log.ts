// path: packages/shared/src/log.ts
// Platform-wide structured-log convention (see ADR-0001 section: Baseline
// Production Hooks).
//
// Contract: every Bun/TS app (currently only the dashboard) outermost
// initialization calls `logger.info(...)` with a JSON-shaped payload. The
// dashboard shapes logs as JSON so they can be ingested by any structured
// logger (Loki, Datadog, CloudWatch) without re-parsing.
//
// The Go api uses its own zap-based equivalent (see apps/api/internal/health
// scaffold) — both languages emit JSON with the same field names:
//   { "ts": <iso8601>, "level": <level>, "msg": <string>, ...fields }

export type LogLevel = "debug" | "info" | "warn" | "error" | "fatal";

export interface Logger {
  debug(msg: string, fields?: Record<string, unknown>): void;
  info(msg: string, fields?: Record<string, unknown>): void;
  warn(msg: string, fields?: Record<string, unknown>): void;
  error(msg: string, fields?: Record<string, unknown>): void;
  fatal(msg: string, fields?: Record<string, unknown>): void;
}

function emit(level: LogLevel, msg: string, fields?: Record<string, unknown>): void {
  // eslint-disable-next-line no-console
  console.log(
    JSON.stringify({
      ts: new Date().toISOString(),
      level,
      msg,
      ...fields,
    })
  );
}

export const logger: Logger = {
  debug: (m, f) => emit("debug", m, f),
  info: (m, f) => emit("info", m, f),
  warn: (m, f) => emit("warn", m, f),
  error: (m, f) => emit("error", m, f),
  // fatal: emit then exit so a fail-fast behavior matches zap.Fatal semantics
  fatal: (m, f) => {
    emit("fatal", m, f);
    process.exit(1);
  },
};
