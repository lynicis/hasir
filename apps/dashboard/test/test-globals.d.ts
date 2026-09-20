export {};

declare global {
  var __sessionMock:
    | (() => {
        session: import("@/lib/session").SessionData | null;
        loading: boolean;
        refreshSession: () => Promise<void>;
      })
    | null
    | undefined;
}
