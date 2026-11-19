import { createClient, type Client } from "@connectrpc/connect";
import { type DescService } from "@bufbuild/protobuf";
import { useMemo } from "react";

import { connectTransport } from "./transport";

export function useClient<T extends DescService>(service: T): Client<T> {
  return useMemo(() => createClient(service, connectTransport), [service]);
}
