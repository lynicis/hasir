import Logo from "@hasir/ui/logo.svg";
import Image from "next/image";
import Link from "next/link";

export function Footer() {
  return (
    <footer className="border-t border-border/40 bg-card py-16">
      <div className="mx-auto max-w-7xl px-6 sm:px-8">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-12 md:gap-8 mb-16">
          <div className="flex flex-col gap-4 md:col-span-1">
            <div className="flex items-center gap-2">
              <Image src={Logo} alt="Hasir Logo" width={24} height={24} />
              <span className="font-semibold text-foreground">hasir</span>
            </div>
            <p className="text-sm text-muted-foreground mt-2 max-w-xs">
              A self-hosted protobuf schema registry with Git-native workflows. Built for teams who ship.
            </p>
          </div>

          <div className="flex flex-col gap-3">
            <h4 className="font-semibold text-foreground text-sm">Product</h4>
            <Link href="#features" className="text-sm text-muted-foreground hover:text-foreground transition-colors">Features</Link>
            <Link href="#workflow" className="text-sm text-muted-foreground hover:text-foreground transition-colors">Workflow</Link>
            <Link href="#languages" className="text-sm text-muted-foreground hover:text-foreground transition-colors">Supported Languages</Link>
          </div>

          <div className="flex flex-col gap-3">
            <h4 className="font-semibold text-foreground text-sm">Resources</h4>
            <Link href="https://github.com/lynicis/hasir" className="text-sm text-muted-foreground hover:text-foreground transition-colors">GitHub Repository</Link>
            <Link href="https://github.com/lynicis/hasir/tree/main/docs" className="text-sm text-muted-foreground hover:text-foreground transition-colors">Documentation</Link>
            <Link href="mailto:me@lynicis.dev" className="text-sm text-muted-foreground hover:text-foreground transition-colors">Contact Support</Link>
          </div>

          <div className="flex flex-col gap-3">
            <h4 className="font-semibold text-foreground text-sm">Community</h4>
            <Link href="https://github.com/lynicis/hasir/issues" className="text-sm text-muted-foreground hover:text-foreground transition-colors">Report an Issue</Link>
            <Link href="https://github.com/lynicis/hasir/pulls" className="text-sm text-muted-foreground hover:text-foreground transition-colors">Contribute</Link>
            <Link href="https://github.com/lynicis/hasir/blob/main/LICENSE" className="text-sm text-muted-foreground hover:text-foreground transition-colors">License (MIT)</Link>
          </div>
        </div>

        <div className="pt-8 border-t border-border/40 flex flex-col md:flex-row justify-between items-center gap-4">
          <div className="text-sm text-muted-foreground">
            © {new Date().getFullYear()} Emre Sırmalı & Contributors. Open-source under MIT.
          </div>
          <div className="text-sm text-muted-foreground">
            Built with 🖤 for engineers.
          </div>
        </div>
      </div>
    </footer>
  );
}
