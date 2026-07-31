interface WorkflowStepProps {
  number: string;
  title: string;
  description: string;
}

export function WorkflowStep({ number, title, description }: WorkflowStepProps) {
  return (
    <div className="flex gap-6">
      <div className="font-mono text-2xl font-bold text-border select-none">{number}</div>
      <div>
        <h4 className="text-xl font-semibold text-foreground mb-2">{title}</h4>
        <p className="text-muted-foreground">{description}</p>
      </div>
    </div>
  );
}
