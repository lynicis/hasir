"use client";

import { Terminal, Shield, Settings, Users, Database, Play, CheckCircle2 } from "lucide-react";
import { motion } from "framer-motion";
import { useState } from "react";
import Image from "next/image";
import Link from "next/link";

export default function LandingPage() {
  return (
    <div className="relative overflow-hidden selection:bg-accent selection:text-background">
      {/* Ambient background blur */}
      <div className="pointer-events-none absolute left-[-20%] top-[-10%] h-[500px] w-[500px] rounded-full bg-accent/5 blur-[120px]" />
      <div className="pointer-events-none absolute right-[-20%] bottom-[-10%] h-[500px] w-[500px] rounded-full bg-primary/5 blur-[140px]" />

      <Navbar />

      <main className="mx-auto max-w-7xl px-6 sm:px-8">
        <HeroSection />
        <FeaturesSection />
        <WorkflowSection />
        <LanguagesSection />
        <BottomCTA />
      </main>

      <Footer />
    </div>
  );
}

function Navbar() {
  return (
    <nav className="sticky top-0 z-50 w-full border-b border-border/40 premium-glass">
      <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-6 sm:px-8">
        <div className="flex items-center gap-2">
          <Image src="/logo.svg" alt="Hasir Logo" width={24} height={24} />
          <span className="font-semibold tracking-tight text-foreground">hasir</span>
        </div>

        <div className="hidden md:flex items-center gap-8 text-sm font-medium text-muted-foreground">
          <Link href="#features" className="hover:text-foreground transition-colors">Features</Link>
          <Link href="#workflow" className="hover:text-foreground transition-colors">Workflow</Link>
          <Link href="#languages" className="hover:text-foreground transition-colors">Languages</Link>
        </div>

        <div className="flex items-center gap-4">
          <Link href="https://github.com/lynicis/hasir" className="rounded-none bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors">
            Star on GitHub ⭐️
          </Link>
        </div>
      </div>
    </nav>
  );
}

function HeroSection() {
  return (
    <section className="relative pt-24 pb-32 sm:pt-32 sm:pb-40">
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-16 lg:gap-8 items-center">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
          className="flex flex-col gap-8"
        >
          <div className="inline-flex items-center rounded-full border border-border/50 bg-card/50 px-3 py-1 text-sm text-muted-foreground w-fit">
            <span className="flex h-2 w-2 rounded-full bg-accent mr-2"></span>
            v1 now available
          </div>

          <h1 className="text-5xl sm:text-6xl md:text-7xl font-bold tracking-tighter text-foreground leading-[1.1]">
            Git-native <br className="hidden sm:block" />
            protobuf registry <br className="hidden sm:block" />
            <span className="text-muted-foreground">for teams who ship.</span>
          </h1>

          <p className="text-lg text-muted-foreground max-w-xl leading-relaxed">
            Push .proto files over standard Git-SSH. Hasir validates with buf, generates client SDKs, and keeps every schema version-controlled.
          </p>

          <div className="flex flex-wrap items-center gap-4">
            <Link href="#features" className="inline-flex h-12 items-center justify-center rounded-none bg-primary px-8 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors">
              View docs
            </Link>
          </div>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0.2 }}
          className="relative lg:ml-auto w-full max-w-lg"
        >
          <div className="premium-glass rounded-xl border border-border overflow-hidden shadow-2xl">
            <div className="flex items-center px-4 py-3 border-b border-border bg-card/80">
              <div className="flex gap-2">
                <div className="w-3 h-3 rounded-full bg-border/80"></div>
                <div className="w-3 h-3 rounded-full bg-border/80"></div>
                <div className="w-3 h-3 rounded-full bg-border/80"></div>
              </div>
              <div className="mx-auto text-xs font-mono text-muted-foreground">hasir push</div>
            </div>
            <div className="p-6 font-mono text-sm leading-relaxed bg-[#0a0b0d]">
              <div className="flex gap-2 text-muted-foreground">
                <span className="text-accent">$</span>
                <span className="text-foreground">git push hasir main</span>
              </div>
              <div className="mt-4 text-muted-foreground">Validating schemas with buf… <span className="text-accent">ok</span></div>
              <div className="mt-2 text-muted-foreground">Generating SDKs: typescript, go</div>
              <div className="mt-4 text-primary">+ 2 SDKs published in 1.4s</div>
            </div>
          </div>

          <div className="absolute -bottom-12 -right-4 sm:-right-12 premium-glass rounded-xl border border-border p-4 shadow-xl w-[320px]">
            <div className="text-xs font-medium text-muted-foreground uppercase tracking-wider mb-4 flex items-center gap-2">
              <div className="w-2 h-2 rounded-full bg-accent animate-pulse"></div>
              live registry activity
            </div>
            <div className="flex flex-col gap-3">
              <ActivityRow hash="a3f9c1e" text="payments/checkout.proto — add refund status enum" time="2m ago" />
              <ActivityRow hash="9e21bd4" text="identity/session.proto — 4 SDKs republished" time="14m ago" />
            </div>
          </div>
        </motion.div>
      </div>
    </section>
  );
}

function ActivityRow({ hash, text, time }: { hash: string, text: string, time: string }) {
  return (
    <div className="flex flex-col gap-1 text-sm border-b border-border/50 pb-3 last:border-0 last:pb-0">
      <div className="flex items-center justify-between">
        <span className="font-mono text-xs text-accent">{hash}</span>
        <span className="text-xs text-muted-foreground">{time}</span>
      </div>
      <span className="text-foreground text-xs truncate">{text}</span>
    </div>
  );
}

function FeaturesSection() {
  const features = [
    {
      icon: <Terminal className="w-5 h-5 text-accent" />,
      title: "Git-over-SSH push",
      desc: "Push .proto files with plain git push over a custom SSH server built for schemas."
    },
    {
      icon: <Play className="w-5 h-5 text-accent" />,
      title: "Automatic SDK generation",
      desc: "Every push validates and packages client SDKs for TypeScript, JavaScript, and Go."
    },
    {
      icon: <Shield className="w-5 h-5 text-accent" />,
      title: "Buf-native validation",
      desc: "Definitions are checked for breaking changes and lint rules before they land."
    },
    {
      icon: <Database className="w-5 h-5 text-accent" />,
      title: "Postgres-backed registry",
      desc: "Every schema version is stored, indexed, and queryable from a single source of truth."
    },
    {
      icon: <Settings className="w-5 h-5 text-accent" />,
      title: "Org & repo management",
      desc: "Manage organizations, repositories, and permissions from one dashboard."
    },
    {
      icon: <Users className="w-5 h-5 text-accent" />,
      title: "SSH key authentication",
      desc: "Developers authenticate with their own SSH keys — no shared tokens, no leaks."
    }
  ];

  return (
    <section id="features" className="py-24 sm:py-32 border-t border-border/40">
      <div className="flex flex-col gap-4 text-center max-w-2xl mx-auto mb-16">
        <h2 className="text-sm font-mono text-accent uppercase tracking-widest">Registry Core</h2>
        <h3 className="text-3xl sm:text-4xl font-bold tracking-tight text-foreground">Everything a schema registry should do.</h3>
        <p className="text-muted-foreground">Built for teams who treat protobuf definitions as source of truth, not an afterthought.</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {features.map((f, i) => (
          <motion.div
            key={i}
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.4, delay: i * 0.1 }}
            className="group p-6 rounded-xl border border-border bg-card hover:bg-card/80 transition-colors"
          >
            <div className="w-10 h-10 rounded-lg bg-background border border-border flex items-center justify-center mb-6">
              {f.icon}
            </div>
            <h4 className="text-lg font-semibold text-foreground mb-2">{f.title}</h4>
            <p className="text-sm text-muted-foreground leading-relaxed">{f.desc}</p>
          </motion.div>
        ))}
      </div>
    </section>
  );
}

function WorkflowSection() {
  return (
    <section id="workflow" className="py-24 sm:py-32 border-t border-border/40">
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-16 items-center">
        <div>
          <h2 className="text-sm font-mono text-accent uppercase tracking-widest mb-4">How it works</h2>
          <h3 className="text-3xl sm:text-4xl font-bold tracking-tight text-foreground mb-12">From commit to client SDK, automatically.</h3>

          <div className="flex flex-col gap-10">
            <WorkflowStep
              num="01"
              title="Push your schema"
              desc="Commit .proto files and push over Git-SSH, just like any other repository."
            />
            <WorkflowStep
              num="02"
              title="Buf validates on receipt"
              desc="Breaking-change detection and lint rules run before anything is accepted."
            />
            <WorkflowStep
              num="03"
              title="SDKs ship themselves"
              desc="Worker pools generate and publish client SDKs for TypeScript, JavaScript, and Go."
            />
          </div>
        </div>

        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          whileInView={{ opacity: 1, scale: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5 }}
          className="premium-glass rounded-xl border border-border overflow-hidden"
        >
          <div className="flex items-center justify-between px-4 py-3 border-b border-border bg-card/80">
            <span className="text-xs font-mono text-muted-foreground">buf.gen.yaml</span>
          </div>
          <div className="p-6 font-mono text-sm leading-relaxed overflow-x-auto">
            <pre className="text-muted-foreground">
              <span className="text-primary">version:</span> v2{"\n"}
              <span className="text-primary">plugins:</span>{"\n"}
              {"  "}- <span className="text-primary">remote:</span> buf.build/protocolbuffers/go{"\n"}
              {"  "}- <span className="text-primary">remote:</span> buf.build/community/typescript{"\n"}
              <span className="text-accent">{"  "}+ hasir: auto-publish on push</span>
            </pre>
          </div>
        </motion.div>
      </div>
    </section>
  );
}

function WorkflowStep({ num, title, desc }: { num: string, title: string, desc: string }) {
  return (
    <div className="flex gap-6">
      <div className="font-mono text-2xl font-bold text-border select-none">{num}</div>
      <div>
        <h4 className="text-xl font-semibold text-foreground mb-2">{title}</h4>
        <p className="text-muted-foreground">{desc}</p>
      </div>
    </div>
  );
}

function LanguagesSection() {
  const langs = ["TypeScript", "JavaScript", "Go"];

  return (
    <section id="languages" className="py-24 sm:py-32 border-t border-border/40 text-center">
      <h2 className="text-sm font-mono text-accent uppercase tracking-widest mb-4">SDK Generation</h2>
      <h3 className="text-3xl sm:text-4xl font-bold tracking-tight text-foreground mb-16">One push. Every SDK.</h3>

      <div className="flex flex-wrap justify-center gap-4 max-w-4xl mx-auto">
        {langs.map((lang, i) => (
          <div key={i} className="flex items-center gap-3 px-6 py-4 rounded-xl border border-border bg-card text-foreground font-medium">
            <CheckCircle2 className="w-5 h-5 text-accent" />
            {lang}
          </div>
        ))}
      </div>
    </section>
  );
}

function BottomCTA() {
  const [activeTab, setActiveTab] = useState<"docker" | "helm">("docker");

  return (
    <section className="py-24 sm:py-32 border-t border-border/40 text-center">
      <h3 className="text-3xl sm:text-4xl font-bold tracking-tight text-foreground mb-6">Deploy anywhere in minutes.</h3>
      <p className="text-lg text-muted-foreground max-w-2xl mx-auto mb-10">
        Run Hasir on your favorite cloud provider or own server. Use our pre-built Docker images or deploy directly to Kubernetes via Helm.
      </p>

      <div className="mx-auto max-w-2xl text-left premium-glass rounded-xl border border-border overflow-hidden shadow-2xl">
        <div className="flex items-center justify-between px-4 py-3 border-b border-border bg-card/80">
          <div className="flex gap-2">
            <div className="w-3 h-3 rounded-full bg-border/80"></div>
            <div className="w-3 h-3 rounded-full bg-border/80"></div>
            <div className="w-3 h-3 rounded-full bg-border/80"></div>
          </div>
          <div className="flex gap-4 text-xs font-mono">
            <button
              onClick={() => setActiveTab("docker")}
              className={`transition-colors ${activeTab === "docker" ? "text-foreground font-semibold" : "text-muted-foreground hover:text-foreground"}`}
            >
              Docker Compose
            </button>
            <button
              onClick={() => setActiveTab("helm")}
              className={`transition-colors ${activeTab === "helm" ? "text-foreground font-semibold" : "text-muted-foreground hover:text-foreground"}`}
            >
              Helm (K8s)
            </button>
          </div>
        </div>
        <div className="p-6 font-mono text-sm leading-relaxed bg-[#0a0b0d] overflow-x-auto min-h-[220px]">
          {activeTab === "docker" ? (
            <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ duration: 0.3 }}>
              <div className="text-muted-foreground/60 mb-2"># Fetch compose config & run (uses pre-built ghcr.io images)</div>
              <div className="flex gap-3 text-muted-foreground">
                <span className="text-accent select-none">$</span>
                <span className="text-foreground whitespace-nowrap">curl -O https://raw.githubusercontent.com/lynicis/hasir/main/docker/docker-compose.yml</span>
              </div>
              <div className="flex gap-3 text-muted-foreground mt-3">
                <span className="text-accent select-none">$</span>
                <span className="text-foreground whitespace-nowrap">docker compose up -d</span>
              </div>
              <div className="mt-5 text-primary">✔ Hasir is running on port 3000!</div>
            </motion.div>
          ) : (
            <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ duration: 0.3 }}>
              <div className="text-muted-foreground/60 mb-2"># Deploy via Helm chart to your cluster</div>
              <div className="flex gap-3 text-muted-foreground">
                <span className="text-accent select-none">$</span>
                <span className="text-foreground whitespace-nowrap">helm repo add hasir https://lynicis.github.io/hasir-charts</span>
              </div>
              <div className="flex gap-3 text-muted-foreground mt-3">
                <span className="text-accent select-none">$</span>
                <span className="text-foreground whitespace-nowrap">helm install my-hasir hasir/hasir --create-namespace -n hasir</span>
              </div>
              <div className="mt-5 text-primary">✔ Hasir deployed to Kubernetes!</div>
            </motion.div>
          )}
        </div>
      </div>
    </section>
  );
}

function Footer() {
  return (
    <footer className="border-t border-border/40 bg-card py-16">
      <div className="mx-auto max-w-7xl px-6 sm:px-8">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-12 md:gap-8 mb-16">
          <div className="flex flex-col gap-4 md:col-span-1">
            <div className="flex items-center gap-2">
              <Image src="/logo.svg" alt="Hasir Logo" width={24} height={24} />
              <span className="font-semibold text-foreground">hasir</span>
            </div>
            <p className="text-sm text-muted-foreground mt-2 max-w-xs">
              A self-hosted protobuf schema registry with Git-native workflows. Built for teams who ship.
            </p>
          </div>

          <div className="flex flex-col gap-3">
            <h4 className="font-semibold text-foreground text-sm">Product</h4>
            <Link href="#features" className="text-sm text-muted-foreground hover:text-foreground transition-colors">Features</Link>
            <Link href="#workflow" className="text-sm text-muted-foreground hover:text-foreground transition-colors">Workflow</Link>
            <Link href="#languages" className="text-sm text-muted-foreground hover:text-foreground transition-colors">Supported Languages</Link>
            <Link href="/login" className="text-sm text-muted-foreground hover:text-foreground transition-colors">Get Started</Link>
          </div>

          <div className="flex flex-col gap-3">
            <h4 className="font-semibold text-foreground text-sm">Resources</h4>
            <Link href="https://github.com/lynicis/hasir" className="text-sm text-muted-foreground hover:text-foreground transition-colors">GitHub Repository</Link>
            <Link href="https://github.com/lynicis/hasir/tree/main/docs" className="text-sm text-muted-foreground hover:text-foreground transition-colors">Documentation</Link>
            <Link href="mailto:me@lynicis.dev" className="text-sm text-muted-foreground hover:text-foreground transition-colors">Contact Support</Link>
          </div>

          <div className="flex flex-col gap-3">
            <h4 className="font-semibold text-foreground text-sm">Community</h4>
            <Link href="https://github.com/lynicis/hasir/issues" className="text-sm text-muted-foreground hover:text-foreground transition-colors">Report an Issue</Link>
            <Link href="https://github.com/lynicis/hasir/pulls" className="text-sm text-muted-foreground hover:text-foreground transition-colors">Contribute</Link>
            <Link href="https://github.com/lynicis/hasir/blob/main/LICENSE" className="text-sm text-muted-foreground hover:text-foreground transition-colors">License (MIT)</Link>
          </div>
        </div>

        <div className="pt-8 border-t border-border/40 flex flex-col md:flex-row justify-between items-center gap-4">
          <div className="text-sm text-muted-foreground">
            © {new Date().getFullYear()} Emre Sırmalı & Contributors. Open-source under MIT.
          </div>
          <div className="text-sm text-muted-foreground">
            Built with 🖤 for engineers.
          </div>
        </div>
      </div>
    </footer>
  );
}
