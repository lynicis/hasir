"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createConnectTransport } from "@connectrpc/connect-web";
import { TransportProvider } from "@connectrpc/connect-query";

import { ThemeProvider } from "@/components/theme-provider";
import { authInterceptor } from "@/lib/auth-interceptor";
import { SessionProvider } from "@/lib/session-provider";
import { Toaster } from "@/components/ui/sonner";

const finalTransport = createConnectTransport({
  baseUrl: process.env.NEXT_PUBLIC_API_URL ?? "",
  interceptors: [authInterceptor],
});

const queryClient = new QueryClient();

export default function Providers({ children }: { children: React.ReactNode }) {
  return (
    <TransportProvider transport={finalTransport}>
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
