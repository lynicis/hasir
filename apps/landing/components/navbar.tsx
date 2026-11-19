import Logo from "@hasir/ui/logo.svg";
import { Star } from "lucide-react";
import Image from "next/image";
import Link from "next/link";

export function Navbar() {
  return (
    <header className="hl-nav">
      <div className="container hl-nav__inner">
        <Link href="/" className="hl-nav__brand" aria-label="Hasir home">
          <Image src={Logo} alt="" width={24} height={24} />
          <span>hasir</span>
        </Link>

        <nav aria-label="Primary" className="hl-nav__links">
          <Link href="#features" className="hl-nav__link">
            Features
          </Link>
          <Link href="#workflow" className="hl-nav__link">
            Workflow
          </Link>
          <Link href="#languages" className="hl-nav__link">
            Languages
          </Link>
        </nav>

        <Link
          href="https://github.com/lynicis/hasir"
          className="hl-btn hl-btn--quiet hl-btn--nav"
        >
          <Star className="size-4 shrink-0" aria-hidden="true" />
          <span>Star on GitHub</span>
        </Link>
      </div>
    </header>
  );
}
