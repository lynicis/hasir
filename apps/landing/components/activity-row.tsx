interface ActivityRowProps {
  hash: string;
  text: string;
  time: string;
  selected?: boolean;
  onSelect?: () => void;
}

export function ActivityRow({
  hash,
  text,
  time,
  selected = false,
  onSelect,
}: ActivityRowProps) {
  return (
    <button
      type="button"
      className="hl-commit-row focus-visible:outline-2 focus-visible:outline-focus active:translate-y-px disabled:opacity-55"
      aria-pressed={selected}
      onClick={onSelect}
    >
      <span className="hl-commit-row__hash">{hash}</span>
      <span className="hl-commit-row__text">{text}</span>
      <span className="hl-commit-row__time">{time}</span>
    </button>
  );
}
