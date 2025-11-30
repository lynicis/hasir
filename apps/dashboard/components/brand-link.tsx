import Image from "next/image";
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
      <Image
        src="/logo.webp"
        alt="Hasir Logo"
        width={24}
        height={24}
        className="size-6"
      />
      {label}
    </Link>
  );
}
