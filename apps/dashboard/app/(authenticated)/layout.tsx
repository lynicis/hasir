import type React from "react";

import { HeaderClient } from "@/components/header-client";

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <>
      <HeaderClient />
      {children}
    </>
  );
}
