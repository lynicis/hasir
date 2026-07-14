import { create } from "zustand";

interface RegistryState {
  organizationsVersion: number;
  repositoriesVersion: number;
  invalidateOrganizations: () => void;
  invalidateRepositories: () => void;
}

export const useRegistryStore = create<RegistryState>((set) => ({
  organizationsVersion: 0,
  repositoriesVersion: 0,
  invalidateOrganizations: () =>
    set((state) => ({
      organizationsVersion: state.organizationsVersion + 1,
    })),
  invalidateRepositories: () =>
    set((state) => ({
      repositoriesVersion: state.repositoriesVersion + 1,
    })),
}));
