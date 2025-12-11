"use client";

import { createContext, useContext, useEffect, useState } from "react";
import { useRouter, usePathname } from "next/navigation";
import { toast } from "sonner";

import { SessionData } from "./session";

interface SessionContextType {
  session: SessionData | null;
  loading: boolean;
  refreshSession: () => Promise<void>;
}

const SessionContext = createContext<SessionContextType>({
  session: null,
  loading: true,
  refreshSession: async () => {},
});

export function useSession() {
  const context = useContext(SessionContext);
  if (!context) {
    throw new Error("useSession must be used within a SessionProvider");
  }
  return context;
}

export function SessionProvider({ children }: { children: React.ReactNode }) {
  const [session, setSession] = useState<SessionData | null>(null);
  const [loading, setLoading] = useState(true);
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    const fetchSession = async () => {
      try {
        const response = await fetch("/api/auth/session");

        if (response.ok) {
          const data = await response.json();
          setSession(data);
        } else {
          setSession(null);

          const publicPaths = ["/login", "/register", '/forgot-password', '/reset-password'];
          const isRootPath = pathname === "/";
          if (
            !isRootPath &&
            !publicPaths.some((path) => pathname.startsWith(path))
          ) {
            router.push("/login");
          }
        }

        setLoading(false);
      } catch {
        toast.error("Failed to fetch session");
        setSession(null);
        setLoading(false);
      }
    };

    fetchSession();
  }, [pathname, router]);

  const refreshSession = async () => {
    try {
      const response = await fetch("/api/auth/session");

      if (response.ok) {
        const data = await response.json();
        setSession(data);
      } else {
        setSession(null);
      }
    } catch {
      toast.error("Failed to refresh session");
      setSession(null);
    }
  };

  return (
    <SessionContext.Provider
      value={{
        session,
        loading,
        refreshSession,
      }}
    >
      {children}
    </SessionContext.Provider>
  );
}
