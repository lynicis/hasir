import { CheckCircle2 } from "lucide-react";

export function LanguagesSection() {
  const langs = ["TypeScript", "JavaScript", "Go"];

  return (
    <section id="languages" className="py-24 sm:py-32 border-t border-border/40 text-center">
      <span className="block text-sm font-mono text-accent uppercase tracking-widest mb-4">SDK Generation</span>
      <h2 className="text-3xl sm:text-4xl font-bold tracking-tight text-foreground mb-16">One push. Every SDK.</h2>

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
