import { motion } from "framer-motion";
import Link from "next/link";

import { ActivityRow } from "@/components/activity-row";

export function HeroSection() {
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
            A self-hosted schema registry for teams who ship. Push .proto files over standard Git-SSH, validate with buf, and automatically generate client SDKs.
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
