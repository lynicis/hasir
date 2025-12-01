import { create } from "zustand";

interface RefreshState {
  organizationsRefreshKey: number;
  repositoriesRefreshKey: number;
  refreshOrganizations: () => void;
  refreshRepositories: () => void;
}

export const useRefreshStore = create<RefreshState>((set) => ({
  organizationsRefreshKey: 0,
  repositoriesRefreshKey: 0,
  refreshOrganizations: () =>
    set((state) => ({
      organizationsRefreshKey: state.organizationsRefreshKey + 1,
    })),
  refreshRepositories: () =>
    set((state) => ({
      repositoriesRefreshKey: state.repositoriesRefreshKey + 1,
    })),
}));
