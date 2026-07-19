import "./globals.css";

import type { Metadata } from "next";

import { Geist, Geist_Mono } from "next/font/google";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  metadataBase: new URL("https://hasir.lynicis.dev"),
  title: "Hasir | Self-Hosted Git-Native Protobuf Registry",
  description: "Hasir is a self-hosted, Git-native protobuf registry. Push .proto files via SSH, validate with buf, and automatically generate client SDKs.",
  alternates: {
    canonical: "/",
  },
  openGraph: {
    title: "Hasir | Self-Hosted Git-Native Protobuf Registry",
    description: "Hasir is a self-hosted, Git-native protobuf registry. Push .proto files via SSH, validate with buf, and automatically generate client SDKs.",
    url: "https://hasir.lynicis.dev",
    siteName: "Hasir",
    locale: "en_US",
    type: "website",
  },
  twitter: {
    card: "summary_large_image",
    title: "Hasir | Self-Hosted Git-Native Protobuf Registry",
    description: "Hasir is a self-hosted, Git-native protobuf registry. Push .proto files via SSH, validate with buf, and automatically generate client SDKs.",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="dark">
      <body className={`${geistSans.variable} ${geistMono.variable} antialiased noise-bg min-h-screen selection:bg-accent selection:text-accent-foreground`}>
        {children}
      </body>
    </html>
  );
}
