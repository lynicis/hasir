import { motion } from "framer-motion";

import { WorkflowStep } from "./workflow-section";

export function WorkflowSection() {
  return (
    <section id="workflow" className="py-24 sm:py-32 border-t border-border/40">
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-16 items-center">
        <div>
          <span className="block text-sm font-mono text-accent uppercase tracking-widest mb-4">How it works</span>
          <h2 className="text-3xl sm:text-4xl font-bold tracking-tight text-foreground mb-12">From commit to client SDK, automatically.</h2>

          <div className="flex flex-col gap-10">
            <WorkflowStep
              number="1"
              title="Push your schema"
              description="Commit .proto files and push over Git-SSH, just like any other repository."
            />
            <WorkflowStep
              number="02"
              title="Buf validates on receipt"
              description="Breaking-change detection and lint rules run before anything is accepted."
            />
            <WorkflowStep
              number="03"
              title="SDKs ship themselves"
              description="Worker pools generate and publish client SDKs for TypeScript, JavaScript, and Go."
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
