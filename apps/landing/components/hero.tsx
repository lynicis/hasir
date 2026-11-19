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
      <div className="container hl-hero__container">
        <div className="hl-hero__head">
          <h1 className="hl-hero__title">
            Self-hosted <span className="whitespace-nowrap">Git-native</span>{" "}
            protobuf registry.
          </h1>

          <div className="hl-hero__bar">
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
        </div>

        <div
          className="hl-counter"
          data-slot="hero"
          data-kind="product"
          data-alt="Hasir commit-to-SDK counter showing a before protobuf diff, server-side buf validation gate, and after commit-pinned client SDK"
        >
          <div
            className="hl-counter__rail"
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

          <div className="hl-counter__matrix">
            <div className="hl-counter__pane">
              <div className="hl-counter__pane-head">
                <span className="hl-counter__step">01 / Before</span>
                <span className="hl-counter__file">{active.file}</span>
              </div>
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

            <div className="hl-counter__pane hl-counter__pane--gate">
              <div className="hl-counter__pane-head">
                <span className="hl-counter__step">02 / Buf gate</span>
                <span className="hl-counter__hash">{active.hash}</span>
              </div>
              <div className="hl-counter__push">
                <div className="flex gap-2">
                  <span className="text-accent select-none">$</span>
                  <span className="text-ink">{active.command}</span>
                </div>
                <div className="text-muted">
                  Validating schemas with buf…{" "}
                  <span className="text-accent">ok</span>
                </div>
                <div className="text-muted">
                  Targets: <span className="text-ink-2">{active.targets}</span>
                </div>
                <div className="text-ink">{active.published}</div>
              </div>
            </div>

            <div className="hl-counter__pane">
              <div className="hl-counter__pane-head">
                <span className="hl-counter__step">03 / After</span>
                <span className="hl-counter__file">{active.sdkLabel}</span>
              </div>
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

          <div className="hl-counter__install">
            <span className="hl-counter__install-key">Install</span>
            <code className="hl-counter__install-cmd">
              {active.installParts.map((part, i) => (
                <Fragment key={part}>
                  {i > 0 && <wbr />}
                  {part}
                </Fragment>
              ))}
            </code>
            <span className="hl-counter__install-state">buf check passed</span>
          </div>
        </div>
      </div>
    </section>
  );
}
