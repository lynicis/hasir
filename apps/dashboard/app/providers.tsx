"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createConnectTransport } from "@connectrpc/connect-web";
import { TransportProvider } from "@connectrpc/connect-query";

import { UserStoreProvider } from "@/stores/user-store-provider";
import { Toaster } from "@/components/ui/sonner";

const finalTransport = createConnectTransport({
  baseUrl: process.env.NEXT_PUBLIC_API_URL ?? "",
});

const queryClient = new QueryClient();

export default function Providers({ children }: { children: React.ReactNode }) {
  return (
    <TransportProvider transport={finalTransport}>
      <QueryClientProvider client={queryClient}>
        <UserStoreProvider>
          <main>{children}</main>
          <Toaster />
        </UserStoreProvider>
      </QueryClientProvider>
    </TransportProvider>
  );
}
