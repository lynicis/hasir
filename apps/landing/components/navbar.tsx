import Logo from "@hasir/ui/logo.svg";
import Image from "next/image";
import Link from "next/link";

export function Navbar() {
  return (
    <nav className="sticky top-0 z-50 w-full border-b border-border/40 premium-glass">
      <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-6 sm:px-8">
        <div className="flex items-center gap-2">
          <Image src={Logo} alt="Hasir Logo" width={24} height={24} />
          <span className="font-semibold tracking-tight text-foreground">hasir</span>
        </div>

        <div className="hidden md:flex items-center gap-8 text-sm font-medium text-muted-foreground">
          <Link href="#features" className="hover:text-foreground transition-colors">Features</Link>
          <Link href="#workflow" className="hover:text-foreground transition-colors">Workflow</Link>
          <Link href="#languages" className="hover:text-foreground transition-colors">Languages</Link>
        </div>

        <div className="flex items-center gap-4">
          <Link href="https://github.com/lynicis/hasir" className="rounded-none bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors">
            Star on GitHub ⭐️
          </Link>
        </div>
      </div>
    </nav>
  );
}
