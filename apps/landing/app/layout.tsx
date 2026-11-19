import "@hasir/ui/landing.css";

import type { Metadata } from "next";

import { Geist, Geist_Mono } from "next/font/google";
import Script from "next/script";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
  weight: ["400", "500", "700"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
  weight: ["400", "500"],
});

export const metadata: Metadata = {
  metadataBase: new URL("https://hasir.lynicis.dev"),
  title: "Hasir | Self-Hosted Git-Native Protobuf Registry",
  description:
    "Hasir is a self-hosted, Git-native protobuf registry. Push .proto files via SSH, validate with buf, and automatically generate client SDKs.",
  alternates: {
    canonical: "/",
  },
  openGraph: {
    title: "Hasir | Self-Hosted Git-Native Protobuf Registry",
    description:
      "Hasir is a self-hosted, Git-native protobuf registry. Push .proto files via SSH, validate with buf, and automatically generate client SDKs.",
    url: "https://hasir.lynicis.dev",
    siteName: "Hasir",
    locale: "en_US",
    type: "website",
  },
  twitter: {
    card: "summary_large_image",
    title: "Hasir | Self-Hosted Git-Native Protobuf Registry",
    description:
      "Hasir is a self-hosted, Git-native protobuf registry. Push .proto files via SSH, validate with buf, and automatically generate client SDKs.",
  },
};

const HALLMARK_SLOTS_SCRIPT = `window.hallmarkSlots = (m) => { for (const [k, v] of Object.entries(m)) { const s = document.querySelector(\`[data-slot="\${k}"]\`); if (!s) continue; const im = new Image(); im.src = typeof v === 'string' ? v : v.src; im.alt = (typeof v === 'object' && v.alt) || s.dataset.alt || ''; s.querySelectorAll(':scope > img').forEach((o) => o.remove()); s.prepend(im); } };
document.querySelectorAll('[data-slot] img').forEach((i) => { if (i.complete && !i.naturalWidth) i.remove(); else i.addEventListener('error', () => i.remove(), { once: true }); });`;

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`dark overflow-x-clip ${geistSans.variable} ${geistMono.variable}`}
    >
      <body className="overflow-x-clip">
        {children}
        <Script id="hallmark-slots" strategy="afterInteractive">
          {HALLMARK_SLOTS_SCRIPT}
        </Script>
      </body>
    </html>
  );
}
