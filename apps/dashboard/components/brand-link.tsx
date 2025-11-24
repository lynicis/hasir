import { Box } from "lucide-react";
import Link from "next/link";

type BrandLinkProps = {
  href?: string;
  label?: string;
  className?: string;
};

export function BrandLink({
  href = "/",
  label = "Hasir Proto Schema Registry",
  className,
}: BrandLinkProps) {
  return (
    <Link
      href={href}
      className={`flex items-center gap-2 self-center font-medium ${
        className ?? ""
      }`.trim()}
    >
      <div className="bg-primary text-primary-foreground flex size-6 items-center justify-center rounded-md">
        <Box className="size-4" />
      </div>
      {label}
    </Link>
  );
}
