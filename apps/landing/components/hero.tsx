"use client";

import { Fragment, useState } from "react";
import Link from "next/link";

import { ActivityRow } from "@/components/activity-row";

interface CommitRecord {
  hash: string;
  file: string;
  summary: string;
  time: string;
  command: string;
  targets: string;
  published: string;
  sdkLabel: string;
  diff: string[];
  sdkPreview: string[];
  installParts: string[];
}

const COMMITS: CommitRecord[] = [
  {
    hash: "a3f9c1e",
    file: "payments/checkout.proto",
    summary: "payments/checkout.proto: add refund status enum",
    time: "2m ago",
    command: "git push hasir main",
    targets: "typescript, go",
    published: "+ 2 SDKs published in 1.4s",
    sdkLabel: "Generated Go SDK",
    diff: [
      "  message Refund {",
      "    string id = 1;",
      "+   RefundStatus status = 2;",
      "  }",
      "+ enum RefundStatus {",
      "+   REFUND_PENDING = 1;",
      "+   REFUND_SETTLED = 2;",
      "+ }",
    ],
    sdkPreview: [
      "type RefundStatus int32",
      "const (",
      "  RefundPending RefundStatus = 1",
      "  RefundSettled RefundStatus = 2",
      ")",
      "client := checkoutv1.NewClient(",
      "  http.DefaultClient,",
      "  registryURL,",
      ")",
    ],
    installParts: [
      "go get registry.example.com/",
      "sdk/",
      "payments/",
      "checkout@a3f9c1e",
    ],
  },
  {
    hash: "9e21bd4",
    file: "identity/session.proto",
    summary: "identity/session.proto: 4 SDKs republished",
    time: "14m ago",
    command: "git push hasir main",
    targets: "typescript, javascript, go",
    published: "+ 4 SDKs republished in 1.8s",
    sdkLabel: "Generated TypeScript SDK",
    diff: [
      "  service SessionService {",
      "    rpc Verify(Req) returns (Lease);",
      "  }",
      "+ message Lease {",
      "+   string token_id = 1;",
      "+   int64 expires_unix = 2;",
      "+ }",
    ],
    sdkPreview: [
      "export interface Lease {",
      "  tokenId: string;",
      "  expiresUnix: bigint;",
      "}",
      "const client = createClient(",
      "  SessionService,",
      "  transport,",
      ");",
    ],
    installParts: [
      "npm install git+ssh://git@registry.example.com:2222/",
      "sdk/",
      "identity/",
      "9e21bd4",
    ],
  },
];

export function HeroSection() {
  const [selectedHash, setSelectedHash] = useState<string>(COMMITS[0]!.hash);
  const active = COMMITS.find((c) => c.hash === selectedHash) ?? COMMITS[0]!;

  return (
    <section className="hl-hero">
      <div className="container">
        <div className="hl-hero__grid">
          <div className="hl-hero__copy">
            <h1 className="hl-hero__title">
              Self-hosted <span className="whitespace-nowrap">Git-native</span>{" "}
              protobuf registry.
            </h1>

            <p className="lead hl-hero__lead">
              Push .proto files over Git-SSH, validate with buf, ship typed
              SDKs.
            </p>

            <div className="hl-hero__actions">
              <Link
                href="https://github.com/lynicis/hasir/tree/main/docs"
                className="hl-btn"
              >
                View docs
              </Link>
              <Link href="#deploy" className="hl-btn hl-btn--quiet">
                Self-host setup
              </Link>
            </div>
          </div>

          <div
            className="hl-workbench"
            data-slot="hero"
            data-kind="product"
            data-alt="Hasir dashboard showing a Git-SSH protobuf push, buf validation check, and generated client SDKs"
          >
            <div className="hl-workbench__bar">
              <span className="hl-workbench__heading">{active.file}</span>
              <div className="hl-workbench__meta">
                <span>buf check passed</span>
                <span className="text-accent font-mono tabular-nums">
                  {active.hash}
                </span>
              </div>
            </div>

            <div className="hl-workbench__push">
              <div className="flex gap-2">
                <span className="text-accent select-none">$</span>
                <span className="text-ink">{active.command}</span>
              </div>
              <div className="text-muted">
                Validating schemas with buf…{" "}
                <span className="text-accent">ok</span>
              </div>
              <div className="text-muted">
                Generating SDKs: {active.targets}
              </div>
              <div className="text-ink">{active.published}</div>
            </div>

            <div className="hl-workbench__split">
              <div className="hl-workbench__pane">
                <span className="hl-workbench__pane-label">Schema diff</span>
                <pre className="hl-diff" aria-label="Schema diff excerpt">
                  {active.diff.map((line, i) => (
                    <code
                      key={i}
                      className={
                        line.startsWith("+") ? "text-ink" : "text-muted"
                      }
                    >
                      {line}
                    </code>
                  ))}
                </pre>
              </div>

              <div className="hl-workbench__pane">
                <span className="hl-workbench__pane-label">
                  {active.sdkLabel}
                </span>
                <pre
                  className="hl-code-lines"
                  aria-label="Generated SDK output"
                >
                  {active.sdkPreview.map((line, i) => (
                    <code key={i}>{line}</code>
                  ))}
                </pre>
              </div>
            </div>

            <div className="hl-workbench__artifact">
              <span className="text-muted shrink-0">Install</span>
              <code>
                {active.installParts.map((part, i) => (
                  <Fragment key={part}>
                    {i > 0 && <wbr />}
                    {part}
                  </Fragment>
                ))}
              </code>
            </div>

            <div
              className="hl-workbench__ledger"
              role="region"
              aria-label="Recent schema pushes"
            >
              {COMMITS.map((commit) => (
                <ActivityRow
                  key={commit.hash}
                  hash={commit.hash}
                  text={commit.summary}
                  time={commit.time}
                  selected={active.hash === commit.hash}
                  onSelect={() => setSelectedHash(commit.hash)}
                />
              ))}
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
