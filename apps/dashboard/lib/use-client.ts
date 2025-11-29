import { createConnectTransport } from "@connectrpc/connect-web";
import { createClient, type Client } from "@connectrpc/connect";
import { type DescService } from "@bufbuild/protobuf";
import { useMemo } from "react";

import { authInterceptor } from "./auth-interceptor";

const transport = createConnectTransport({
  baseUrl: process.env.NEXT_PUBLIC_API_URL ?? "",
  interceptors: [authInterceptor]
});


export function useClient<T extends DescService>(service: T): Client<T> {
  return useMemo(() => createClient(service, transport), [service]);
}