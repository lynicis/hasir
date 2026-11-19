"use client";

import { Check, Copy } from "lucide-react";
import { Fragment, useState } from "react";
import Link from "next/link";

const DOCKER_COMMANDS =
  "curl -O https://raw.githubusercontent.com/lynicis/hasir/main/deploy/docker/docker-compose.yml\ndocker compose up -d";

const HELM_COMMANDS =
  "helm repo add hasir https://lynicis.github.io/hasir-charts\nhelm install my-hasir hasir/hasir --create-namespace -n hasir";

const DOCKER_URL_PARTS = [
  "curl -O https://raw.githubusercontent.com/",
  "lynicis/",
  "hasir/",
  "main/",
  "deploy/",
  "docker/",
  "docker-compose.yml",
];

const HELM_URL_PARTS = [
  "helm repo add hasir https://lynicis.github.io/",
  "hasir-charts",
];

export function BottomCTA() {
  const [activeTab, setActiveTab] = useState<"docker" | "helm">("docker");
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(
        activeTab === "docker" ? DOCKER_COMMANDS : HELM_COMMANDS
      );
      setCopied(true);
      setTimeout(() => setCopied(false), 1800);
    } catch {
      setCopied(false);
    }
  };

  return (
    <section id="deploy" className="hl-section--deploy">
      <div className="container">
        <div className="hl-deploy-band">
          <div className="hl-deploy-band__top">
            <div>
              <h2>Run it on Docker Compose or Helm.</h2>
              <p className="lead hl-deploy-band__lead">
                Self-host Hasir on your own server with pre-built ghcr.io images
                or deploy directly to Kubernetes.
              </p>
            </div>

            <Link
              href="https://github.com/lynicis/hasir/tree/main/docs"
              className="hl-btn hl-btn--quiet"
            >
              Read deployment guide
            </Link>
          </div>

          <div className="hl-deploy-strip">
            <div className="hl-deploy-strip__cmd">
              <div className="hl-deploy-strip__bar">
                <div
                  className="hl-deploy-strip__tabs"
                  role="tablist"
                  aria-label="Deployment method"
                >
                  <button
                    type="button"
                    role="tab"
                    aria-selected={activeTab === "docker"}
                    onClick={() => setActiveTab("docker")}
                    className="hl-deploy-tab"
                  >
                    Docker Compose
                  </button>
                  <button
                    type="button"
                    role="tab"
                    aria-selected={activeTab === "helm"}
                    onClick={() => setActiveTab("helm")}
                    className="hl-deploy-tab"
                  >
                    Helm (K8s)
                  </button>
                </div>

                <button
                  type="button"
                  className="hl-copy-btn"
                  onClick={handleCopy}
                  aria-label="Copy deployment commands"
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

              <div className="hl-deploy-strip__body">
                {activeTab === "docker" ? (
                  <div>
                    <div className="text-muted mb-2">
                      # Fetch compose config and run pre-built ghcr.io images
                    </div>
                    <div className="flex gap-3">
                      <span className="text-accent select-none">$</span>
                      <span className="text-ink min-w-0">
                        {DOCKER_URL_PARTS.map((part, i) => (
                          <Fragment key={part}>
                            {i > 0 && <wbr />}
                            {part}
                          </Fragment>
                        ))}
                      </span>
                    </div>
                    <div className="flex gap-3 mt-3">
                      <span className="text-accent select-none">$</span>
                      <span className="text-ink">docker compose up -d</span>
                    </div>
                    <div className="mt-5 text-accent">
                      Hasir is running on port 3000.
                    </div>
                  </div>
                ) : (
                  <div>
                    <div className="text-muted mb-2">
                      # Deploy via Helm chart to your cluster
                    </div>
                    <div className="flex gap-3">
                      <span className="text-accent select-none">$</span>
                      <span className="text-ink min-w-0">
                        {HELM_URL_PARTS.map((part, i) => (
                          <Fragment key={part}>
                            {i > 0 && <wbr />}
                            {part}
                          </Fragment>
                        ))}
                      </span>
                    </div>
                    <div className="flex gap-3 mt-3">
                      <span className="text-accent select-none">$</span>
                      <span className="text-ink">
                        helm install my-hasir hasir/hasir --create-namespace -n
                        hasir
                      </span>
                    </div>
                    <div className="mt-5 text-accent">
                      Hasir deployed to Kubernetes.
                    </div>
                  </div>
                )}
              </div>
            </div>

            <aside className="hl-deploy-strip__aside">
              <dl className="hl-deploy-specs">
                <div>
                  <dt>Runtime ports</dt>
                  <dd className="font-mono tabular-nums">
                    80 / 443 HTTP, 2222 SSH
                  </dd>
                </div>
                <div>
                  <dt>Storage</dt>
                  <dd>PostgreSQL database</dd>
                </div>
                <div>
                  <dt>Container images</dt>
                  <dd className="font-mono">ghcr.io/lynicis/hasir-*</dd>
                </div>
              </dl>
            </aside>
          </div>
        </div>
      </div>
    </section>
  );
}
