"use client";

import { Check, Copy } from "lucide-react";
import { Fragment, useState } from "react";

interface SdkTarget {
  id: "typescript" | "javascript" | "go";
  label: string;
  install: string;
  installParts: string[];
  protocol: string;
  transport: string;
  snippetFile: string;
  snippet: string[];
}

const TARGETS: SdkTarget[] = [
  {
    id: "typescript",
    label: "TypeScript",
    install:
      "npm install git+ssh://git@registry.example.com:2222/sdk/payments/checkout/a3f9c1e/js-connectrpc/",
    installParts: [
      "npm install ",
      "git+ssh://",
      "git@registry.example.com:2222/",
      "sdk/",
      "payments/",
      "checkout/",
      "a3f9c1e/",
      "js-connectrpc/",
    ],
    protocol: "Connect, gRPC-Web",
    transport: "Git-SSH or HTTPS",
    snippetFile: "checkout-client.ts",
    snippet: [
      "import { createClient } from \"@connectrpc/connect\";",
      "import { createConnectTransport } from \"@connectrpc/connect-web\";",
      "import { CheckoutService } from \"@hasir-sdk/payments\";",
      "",
      "const transport = createConnectTransport({",
      "  baseUrl: \"https://api.example.com\",",
      "});",
      "const client = createClient(CheckoutService, transport);",
      "const res = await client.issueRefund({ orderId: \"ord_1042\" });",
    ],
  },
  {
    id: "javascript",
    label: "JavaScript",
    install:
      "npm install git+https://registry.example.com/sdk/payments/checkout/a3f9c1e/js-connectrpc/",
    installParts: [
      "npm install ",
      "git+https://",
      "registry.example.com/",
      "sdk/",
      "payments/",
      "checkout/",
      "a3f9c1e/",
      "js-connectrpc/",
    ],
    protocol: "ES modules, CommonJS",
    transport: "Commit-pinned package",
    snippetFile: "checkout-client.mjs",
    snippet: [
      "import { createClient } from \"@connectrpc/connect\";",
      "import { createConnectTransport } from \"@connectrpc/connect-node\";",
      "import { CheckoutService } from \"@hasir-sdk/payments\";",
      "",
      "const transport = createConnectTransport({",
      "  baseUrl: \"https://api.example.com\",",
      "  httpVersion: \"2\",",
      "});",
      "const client = createClient(CheckoutService, transport);",
      "const refund = await client.getRefund({ id: \"ref_881\" });",
    ],
  },
  {
    id: "go",
    label: "Go",
    install:
      "go get registry.example.com/sdk/payments/checkout/a3f9c1e/go-connectrpc@a3f9c1e",
    installParts: [
      "go get registry.example.com/",
      "sdk/",
      "payments/",
      "checkout/",
      "a3f9c1e/",
      "go-connectrpc@a3f9c1e",
    ],
    protocol: "Connect, gRPC, HTTP/2",
    transport: "GOPRIVATE module proxy",
    snippetFile: "checkout_client.go",
    snippet: [
      "import (",
      "  \"net/http\"",
      "  checkout \"registry.example.com/sdk/payments/go\"",
      ")",
      "",
      "client := checkout.NewCheckoutServiceClient(",
      "  http.DefaultClient,",
      "  \"https://api.example.com\",",
      ")",
    ],
  },
];

export function LanguagesSection() {
  const [selectedId, setSelectedId] = useState<SdkTarget["id"]>("typescript");
  const [copied, setCopied] = useState(false);

  const active = TARGETS.find((t) => t.id === selectedId) ?? TARGETS[0]!;

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(active.install);
      setCopied(true);
      setTimeout(() => setCopied(false), 1800);
    } catch {
      setCopied(false);
    }
  };

  return (
    <section id="languages" className="hl-section--void">
      <div className="container">
        <div className="hl-head-block--split">
          <h2>One push. Every SDK.</h2>

          <div
            className="hl-seg"
            role="tablist"
            aria-label="Supported SDK languages"
          >
            {TARGETS.map((target) => (
              <button
                key={target.id}
                type="button"
                role="tab"
                aria-selected={active.id === target.id}
                className="hl-seg__btn"
                onClick={() => {
                  setSelectedId(target.id);
                  setCopied(false);
                }}
              >
                {target.label}
              </button>
            ))}
          </div>
        </div>

        <div className="hl-sdk-panel" role="tabpanel" aria-label={active.label}>
          <div className="hl-sdk-panel__left">
            <div className="flex flex-col gap-4">
              <h3>{active.label} client package</h3>
              <p className="text-muted">
                Every pushed commit publishes an immutable, commit-pinned module
                ready for package managers.
              </p>

              <div className="hl-sdk-cmd">
                <div className="hl-sdk-cmd__text">
                  <code>
                    {active.installParts.map((part, i) => (
                      <Fragment key={part}>
                        {i > 0 && <wbr />}
                        {part}
                      </Fragment>
                    ))}
                  </code>
                </div>
                <button
                  type="button"
                  className="hl-copy-btn"
                  onClick={handleCopy}
                  aria-label={`Copy ${active.label} install command`}
                >
                  {copied ? (
                    <Check className="size-4 text-accent" aria-hidden="true" />
                  ) : (
                    <Copy className="size-4" aria-hidden="true" />
                  )}
                  <span className="hidden sm:inline">
                    {copied ? "Copied" : "Copy"}
                  </span>
                </button>
              </div>
            </div>

            <dl className="hl-sdk-facts">
              <div>
                <dt>Protocol</dt>
                <dd>{active.protocol}</dd>
              </div>
              <div>
                <dt>Distribution</dt>
                <dd>{active.transport}</dd>
              </div>
            </dl>
          </div>

          <div className="hl-sdk-panel__right">
            <div className="hl-inspector__bar">
              <span className="text-ink font-medium">{active.snippetFile}</span>
              <span className="text-muted">ConnectRPC</span>
            </div>
            <div className="hl-inspector__body">
              <pre>
                {active.snippet.map((line, i) => (
                  <div
                    key={i}
                    className={
                      line.startsWith("import") || line.startsWith("const")
                        ? "text-ink"
                        : "text-ink-2"
                    }
                  >
                    {line || "\u00A0"}
                  </div>
                ))}
              </pre>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
