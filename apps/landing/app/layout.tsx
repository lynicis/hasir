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
  title: "Hasir - Git-native protobuf registry",
  description: "Git-native protobuf registry for teams who ship. Push .proto files over standard Git-SSH. Hasir validates with buf, generates client SDKs, and keeps every schema version-controlled.",
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
