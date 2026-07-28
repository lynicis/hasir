export function ActivityRow({ hash, text, time }: { hash: string, text: string, time: string }) {
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
