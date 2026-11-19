interface WorkflowStepProps {
  number: string;
  title: string;
  description: string;
}

export function WorkflowStep({
  number,
  title,
  description,
}: WorkflowStepProps) {
  return (
    <li className="hl-step-col">
      <span className="hl-step-col__num" aria-hidden="true">
        {number}
      </span>
      <h3>{title}</h3>
      <p>{description}</p>
    </li>
  );
}
