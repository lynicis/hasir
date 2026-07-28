import { Database, Play, Settings, Shield, Terminal, Users } from "lucide-react";
import { motion } from "framer-motion";

export function FeaturesSection() {
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
        <span className="block text-sm font-mono text-accent uppercase tracking-widest">Registry Core</span>
        <h2 className="text-3xl sm:text-4xl font-bold tracking-tight text-foreground">Everything a schema registry should do.</h2>
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
