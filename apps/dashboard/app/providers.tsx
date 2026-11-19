"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { TransportProvider } from "@connectrpc/connect-query";
import { Toaster } from "@hasir/ui/components/sonner";

import { ThemeProvider } from "@/components/theme-provider";
import { SessionProvider } from "@/lib/session-provider";
import { connectTransport } from "@/lib/transport";

const queryClient = new QueryClient();

export default function Providers({ children }: { children: React.ReactNode }) {
  return (
    <TransportProvider transport={connectTransport}>
      <QueryClientProvider client={queryClient}>
        <SessionProvider>
          <ThemeProvider
            attribute="class"
            defaultTheme="system"
            enableSystem
            disableTransitionOnChange
          >
            <main>{children}</main>
            <Toaster />
          </ThemeProvider>
        </SessionProvider>
      </QueryClientProvider>
    </TransportProvider>
  );
}
