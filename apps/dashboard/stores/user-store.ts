import { createStore } from 'zustand/vanilla'
import { decodeJwt } from 'jose'

export type TokenEnvelope = {
    accessToken: string;
    refreshToken: string;
};

export type UserState = {
  id: string;
  email: string;
  tokens: TokenEnvelope;
}

export type UserActions = {
  setTokens(tokens: TokenEnvelope): void;
}

export type UserStore = UserState & UserActions

function decodeJWT(token: string): { id: string | null; email: string | null } {
  try {
    const payload = decodeJwt(token);
    
    return {
      id: payload.sub as string,
      email: payload.email as string,
    };
  } catch {
    return { id: null, email: null };
  }
}

export const defaultInitState: UserState = {
  id: "",
  email: "",
  tokens: {
      accessToken: "",
      refreshToken: ""
  }
}

export const createUserStore = (
  initState: UserState = defaultInitState,
) => {
  return createStore<UserStore>()((set) => ({
    ...initState,
    setTokens: (tokens) => {
      const decoded = decodeJWT(tokens.accessToken);
      set({
        id: decoded.id || "",
        email: decoded.email || "",
        tokens,
      });
    },
  }))
}