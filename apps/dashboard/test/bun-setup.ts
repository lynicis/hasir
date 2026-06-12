/* eslint-disable */
import { GlobalRegistrator } from "@happy-dom/global-registrator";

GlobalRegistrator.register();

if (typeof document !== "undefined" && !document.body) {
  const body = document.createElement("body");
  const html = document.querySelector("html");
  if (html) {
    html.appendChild(body);
  }
}

if (typeof document !== "undefined" && document.body) {
  document.body.innerHTML = "";
}

import * as matchers from "@testing-library/jest-dom/matchers";
import { afterEach, expect, mock } from "bun:test";

const React = { createElement: (x: any) => x };

expect.extend(matchers);

import { beforeEach } from "bun:test";

const { cleanup } = require("@testing-library/react");

beforeEach(() => {
  if (typeof document !== "undefined" && document.body) {
    document.body.innerHTML = "";
  }
});

afterEach(() => {
  cleanup();
  if (typeof document !== "undefined" && document.body) {
    document.body.innerHTML = "";
    const root = document.createElement("div");
    root.setAttribute("id", "root");
    document.body.appendChild(root);
  }
});

class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
}

if (typeof globalThis.ResizeObserver === "undefined") {
  globalThis.ResizeObserver = ResizeObserver as unknown as typeof ResizeObserver;
}

if (typeof Element.prototype.scrollIntoView === "undefined") {
  Element.prototype.scrollIntoView = () => {};
}

if (typeof Element.prototype.hasPointerCapture === "undefined") {
  Element.prototype.hasPointerCapture = () => false;
}

if (typeof Element.prototype.setPointerCapture === "undefined") {
  Element.prototype.setPointerCapture = () => {};
}

if (typeof Element.prototype.releasePointerCapture === "undefined") {
  Element.prototype.releasePointerCapture = () => {};
}

/* --- Global test mock registries --- */
// @connectrpc/connect-query
(globalThis as any).__connectQueryMocks = new Map<string, () => any>();
(globalThis as any).__defaultUseQuery = () => ({ data: null, isLoading: false });

mock.module("@connectrpc/connect-query", () => {
  const map = (globalThis as any).__connectQueryMocks as Map<string, () => any>;
  const def = (globalThis as any).__defaultUseQuery as () => any;
  return {
    TransportProvider: ({ children }: any) => children,
    useQuery: mock((queryFn: any) => {
      const key = queryFn?.name || queryFn?.toString() || "";
      for (const [pattern, factory] of map) {
        if (key.includes(pattern)) return factory();
      }
      return def();
    }),
    useMutation: () => ({
      mutateAsync: mock().mockResolvedValue({}),
      isPending: false,
    }),
  };
});

// next/navigation
(globalThis as any).__nextNav = {
  useParams: () => ({}),
  useSearchParams: () => new URLSearchParams(),
  usePathname: () => "/",
  useRouter: () => ({ push: mock(), replace: mock(), back: mock(), prefetch: mock() }),
};

mock.module("next/navigation", () => {
  const n = (globalThis as any).__nextNav;
  return {
    useRouter: mock(() => n.useRouter()),
    useParams: mock(() => n.useParams()),
    useSearchParams: mock(() => n.useSearchParams()),
    usePathname: mock(() => n.usePathname()),
  };
});

// @/lib/use-client (configurable via __client global)
(globalThis as any).__client = {};

mock.module("@/lib/use-client", () => ({
  useClient: () => (globalThis as any).__client,
}));

// @connectrpc/connect-web (returns a transport with unary/stream for compatibility)
(globalThis as any).__connectWebTransport = {
  unary: mock(),
  stream: mock(),
};

mock.module("@connectrpc/connect-web", () => ({
  createConnectTransport: mock(() => (globalThis as any).__connectWebTransport),
}));

// next/dynamic — must be in preload so it intercepts imports inside client components
(globalThis as any).__nextDynamic = {
  default: (loader?: any) => () => null,
};

mock.module("next/dynamic", () => {
  const reg = (globalThis as any).__nextDynamic as { default: (loader?: any) => any };
  return {
    __esModule: true,
    default: (loader?: any) => {
      if (reg.default) return reg.default(loader);
      return () => null;
    },
  };
});

// sonner
(globalThis as any).__toast = { success: mock(), error: mock(), info: mock() };

mock.module("sonner", () => ({
  toast: (globalThis as any).__toast,
}));



