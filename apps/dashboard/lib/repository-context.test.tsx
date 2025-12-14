import type { Repository } from "@buf/hasir_hasir.bufbuild_es/registry/v1/registry_pb";

import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { useContext } from "react";

import { RepositoryContext } from "./repository-context";

describe("RepositoryContext", () => {
  it("has a default value of null", () => {
    const TestComponent = () => {
      const context = useContext(RepositoryContext);
      return <div>{context === null ? "null" : "not null"}</div>;
    };

    render(<TestComponent />);
    expect(screen.getByText("null")).toBeInTheDocument();
  });

  it("provides context values when a provider is used", () => {
    const mockRepository: Repository = {
      $typeName: "registry.v1.Repository",
      id: "test-repo-id",
      name: "test-repo",
      organizationId: "test-org-id",
      visibility: 0,
      sdkPreferences: [],
    };

    const contextValue = {
      repository: mockRepository,
      isLoading: false,
      error: null,
    };

    const TestComponent = () => {
      const context = useContext(RepositoryContext);
      return (
        <div>
          <div>Repository ID: {context?.repository?.id}</div>
          <div>Loading: {context?.isLoading ? "yes" : "no"}</div>
          <div>Error: {context?.error ? "yes" : "no"}</div>
        </div>
      );
    };
    render(
      <RepositoryContext.Provider value={contextValue}>
        <TestComponent />
      </RepositoryContext.Provider>
    );

    expect(screen.getByText("Repository ID: test-repo-id")).toBeInTheDocument();
    expect(screen.getByText("Loading: no")).toBeInTheDocument();
    expect(screen.getByText("Error: no")).toBeInTheDocument();
  });

  it("provides loading state correctly", () => {
    const contextValue = {
      repository: undefined,
      isLoading: true,
      error: null,
    };

    const TestComponent = () => {
      const context = useContext(RepositoryContext);
      return <div>Loading: {context?.isLoading ? "yes" : "no"}</div>;
    };

    render(
      <RepositoryContext.Provider value={contextValue}>
        <TestComponent />
      </RepositoryContext.Provider>
    );

    expect(screen.getByText("Loading: yes")).toBeInTheDocument();
  });

  it("provides error state correctly", () => {
    const contextValue = {
      repository: undefined,
      isLoading: false,
      error: new Error("Test error"),
    };

    const TestComponent = () => {
      const context = useContext(RepositoryContext);
      return <div>Has error: {context?.error ? "yes" : "no"}</div>;
    };

    render(
      <RepositoryContext.Provider value={contextValue}>
        <TestComponent />
      </RepositoryContext.Provider>
    );

    expect(screen.getByText("Has error: yes")).toBeInTheDocument();
  });
});
