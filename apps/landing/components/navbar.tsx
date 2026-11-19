"use client";

import { useEffect, useState } from "react";
import Logo from "@hasir/ui/logo.svg";
import { Star } from "lucide-react";
import Image from "next/image";
import Link from "next/link";

export function Navbar() {
  const [retracted, setRetracted] = useState(false);

  useEffect(() => {
    let lastY = window.scrollY;
    let ticking = false;

    const onScroll = () => {
      if (ticking) return;
      ticking = true;
      window.requestAnimationFrame(() => {
        const y = window.scrollY;
        if (y > 96 && y > lastY + 4) {
          setRetracted(true);
        } else if (y < lastY - 4 || y <= 48) {
          setRetracted(false);
        }
        lastY = y;
        ticking = false;
      });
    };

    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  return (
    <>
      <div
        className="hl-nav-banner"
        aria-label="Registry protocol specification"
      >
        <div className="container hl-nav-banner__inner">
          <div className="hl-nav-banner__group">
            <span className="hl-nav-banner__token">ssh :2222</span>
            <span className="hl-nav-banner__token">buf lint + breaking</span>
            <span className="hl-nav-banner__token">connectrpc</span>
          </div>
          <span className="hl-nav-banner__token">ghcr.io/lynicis/hasir</span>
        </div>
      </div>

      <header className="hl-nav" data-retracted={retracted ? "true" : "false"}>
        <div className="container hl-nav__inner">
          <Link href="/" className="hl-nav__brand" aria-label="Hasir home">
            <Image src={Logo} alt="" width={24} height={24} />
            <span>hasir</span>
          </Link>

          <nav aria-label="Primary" className="hl-nav__links">
            <Link href="#workflow" className="hl-nav__link">
              Workflow
            </Link>
            <Link href="#languages" className="hl-nav__link">
              Languages
            </Link>
            <Link href="#features" className="hl-nav__link">
              Features
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
    </>
  );
}
