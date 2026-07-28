import { motion } from "framer-motion";
import { useState } from "react";

export function BottomCTA() {
  const [activeTab, setActiveTab] = useState<"docker" | "helm">("docker");

  return (
    <section className="py-24 sm:py-32 border-t border-border/40 text-center">
      <h2 className="text-3xl sm:text-4xl font-bold tracking-tight text-foreground mb-6">Deploy anywhere in minutes.</h2>
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
        <div className="p-6 font-mono text-sm leading-relaxed bg-[#0a0b0d] overflow-x-auto min-h-55">
          {activeTab === "docker" ? (
            <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ duration: 0.3 }}>
              <div className="text-muted-foreground/60 mb-2"># Fetch compose config & run (uses pre-built ghcr.io images)</div>
              <div className="flex gap-3 text-muted-foreground">
                <span className="text-accent select-none">$</span>
                <span className="text-foreground whitespace-nowrap">curl -O https://raw.githubusercontent.com/lynicis/hasir/main/deploy/docker/docker-compose.yml</span>
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
