import Logo from "@hasir/ui/logo.svg";
import Image from "next/image";
import Link from "next/link";

export function Footer() {
  return (
    <footer className="hl-footer">
      <div className="container">
        <div className="hl-footer__grid">
          <div className="hl-footer__mast">
            <div className="flex flex-col gap-2">
              <Link
                href="/"
                className="hl-footer__brand w-fit"
                aria-label="Hasir home"
              >
                <Image src={Logo} alt="" width={24} height={24} />
                <span>hasir</span>
              </Link>
              <p className="hl-footer__desc">
                Self-hosted protobuf schema registry with Git-native workflows.
              </p>
            </div>

            <p className="hl-footer__legal">
              © {new Date().getFullYear()} Emre Sırmalı &amp; Contributors.
              Hasir v1, open-source under MIT.
            </p>
          </div>

          <div className="hl-footer__col">
            <h3 className="hl-footer__head">Product</h3>
            <Link href="#workflow" className="hl-footer__link">
              Workflow
            </Link>
            <Link href="#languages" className="hl-footer__link">
              Supported Languages
            </Link>
            <Link href="#features" className="hl-footer__link">
              Features
            </Link>
          </div>

          <div className="hl-footer__col">
            <h3 className="hl-footer__head">Resources</h3>
            <Link
              href="https://github.com/lynicis/hasir"
              className="hl-footer__link"
            >
              GitHub Repository
            </Link>
            <Link
              href="https://github.com/lynicis/hasir/tree/main/docs"
              className="hl-footer__link"
            >
              Documentation
            </Link>
            <Link href="mailto:me@lynicis.dev" className="hl-footer__link">
              Contact Support
            </Link>
          </div>

          <div className="hl-footer__col">
            <h3 className="hl-footer__head">Community</h3>
            <Link
              href="https://github.com/lynicis/hasir/issues"
              className="hl-footer__link"
            >
              Report an Issue
            </Link>
            <Link
              href="https://github.com/lynicis/hasir/pulls"
              className="hl-footer__link"
            >
              Contribute
            </Link>
            <Link
              href="https://github.com/lynicis/hasir/blob/main/LICENSE"
              className="hl-footer__link"
            >
              License (MIT)
            </Link>
          </div>
        </div>
      </div>
    </footer>
  );
}
