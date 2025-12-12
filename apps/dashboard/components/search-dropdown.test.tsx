import type { Organization } from "@buf/hasir_hasir.bufbuild_es/organization/v1/organization_pb";
import type { Repository } from "@buf/hasir_hasir.bufbuild_es/registry/v1/registry_pb";

import { Visibility } from "@buf/hasir_hasir.bufbuild_es/shared/visibility_pb";
import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { SearchDropdown } from "./search-dropdown";

describe("SearchDropdown", () => {
  const mockOnResultClick = vi.fn();

  const mockOrganizations: Organization[] = [
    {
      id: "org-1",
      name: "Test Organization 1",
      visibility: Visibility.UNSPECIFIED,
    } as Organization,
    {
      id: "org-2",
      name: "Test Organization 2",
      visibility: Visibility.PUBLIC,
    } as Organization,
  ];

  const mockRepositories: Repository[] = [
    {
      $typeName: "registry.v1.Repository",
      id: "repo-1",
      name: "test-repo-1",
      visibility: Visibility.PUBLIC,
      sdkPreferences: [],
    },
    {
      $typeName: "registry.v1.Repository",
      id: "repo-2",
      name: "test-repo-2",
      visibility: Visibility.PRIVATE,
      sdkPreferences: [],
    },
  ];

  it("should not render when query is empty", () => {
    const { container } = render(
      <SearchDropdown
        query=""
        organizations={[]}
        repositories={[]}
        isLoading={false}
        error={null}
        onResultClick={mockOnResultClick}
      />
    );

    expect(container.firstChild).toBeNull();
  });

  it("should show loading skeletons when isLoading is true", () => {
    const { container } = render(
      <SearchDropdown
        query="test"
        organizations={[]}
        repositories={[]}
        isLoading={true}
        error={null}
        onResultClick={mockOnResultClick}
      />
    );

    const skeletons = container.querySelectorAll('[data-slot="skeleton"]');
    expect(skeletons.length).toBeGreaterThan(0);
  });

  it("should display error message when error occurs", () => {
    const error = new Error("Search failed");

    render(
      <SearchDropdown
        query="test"
        organizations={[]}
        repositories={[]}
        isLoading={false}
        error={error}
        onResultClick={mockOnResultClick}
      />
    );

    expect(screen.getByText(/Error loading search results/i)).toBeInTheDocument();
    expect(screen.getByText(/Search failed/i)).toBeInTheDocument();
  });

  it("should show no results message when no organizations or repositories found", () => {
    render(
      <SearchDropdown
        query="nonexistent"
        organizations={[]}
        repositories={[]}
        isLoading={false}
        error={null}
        onResultClick={mockOnResultClick}
      />
    );

    expect(
      screen.getByText(/No results found for "nonexistent"/i)
    ).toBeInTheDocument();
  });

  it("should display organizations with correct visibility", () => {
    const { container } = render(
      <SearchDropdown
        query="test"
        organizations={mockOrganizations}
        repositories={[]}
        isLoading={false}
        error={null}
        onResultClick={mockOnResultClick}
      />
    );

    expect(screen.getByText("Test Organization 1")).toBeInTheDocument();
    expect(screen.getByText("Test Organization 2")).toBeInTheDocument();

    const publicOrg = container.textContent?.includes("Public organization");
    const privateOrg = container.textContent?.includes("Private organization");

    expect(publicOrg).toBe(true);
    expect(privateOrg).toBe(true);
  });

  it("should display repositories with correct visibility", () => {
    render(
      <SearchDropdown
        query="test"
        organizations={[]}
        repositories={mockRepositories}
        isLoading={false}
        error={null}
        onResultClick={mockOnResultClick}
      />
    );

    expect(screen.getByText("test-repo-1")).toBeInTheDocument();
    expect(screen.getByText("test-repo-2")).toBeInTheDocument();
  });

  it("should display both organizations and repositories", () => {
    render(
      <SearchDropdown
        query="test"
        organizations={mockOrganizations}
        repositories={mockRepositories}
        isLoading={false}
        error={null}
        onResultClick={mockOnResultClick}
      />
    );

    expect(screen.getByText(/Organizations \(2\)/i)).toBeInTheDocument();
    expect(screen.getByText(/Repositories \(2\)/i)).toBeInTheDocument();
  });

  it("should render correct links for organizations", () => {
    render(
      <SearchDropdown
        query="test"
        organizations={mockOrganizations}
        repositories={[]}
        isLoading={false}
        error={null}
        onResultClick={mockOnResultClick}
      />
    );

    const orgLink1 = screen
      .getByText("Test Organization 1")
      .closest("a");
    const orgLink2 = screen
      .getByText("Test Organization 2")
      .closest("a");

    expect(orgLink1).toHaveAttribute("href", "/organization/org-1/users");
    expect(orgLink2).toHaveAttribute("href", "/organization/org-2/users");
  });

  it("should render correct links for repositories", () => {
    render(
      <SearchDropdown
        query="test"
        organizations={[]}
        repositories={mockRepositories}
        isLoading={false}
        error={null}
        onResultClick={mockOnResultClick}
      />
    );

    const repoLink1 = screen.getByText("test-repo-1").closest("a");
    const repoLink2 = screen.getByText("test-repo-2").closest("a");

    expect(repoLink1).toHaveAttribute("href", "/repository/repo-1");
    expect(repoLink2).toHaveAttribute("href", "/repository/repo-2");
  });

  it("should show 'View all results' link when results exist", () => {
    render(
      <SearchDropdown
        query="test query"
        organizations={mockOrganizations}
        repositories={mockRepositories}
        isLoading={false}
        error={null}
        onResultClick={mockOnResultClick}
      />
    );

    const viewAllLink = screen.getByText("View all results");
    expect(viewAllLink).toBeInTheDocument();
    expect(viewAllLink.closest("a")).toHaveAttribute(
      "href",
      "/search?q=test%20query"
    );
  });

  it("should show 'View all results' link with only organizations", () => {
    render(
      <SearchDropdown
        query="test"
        organizations={mockOrganizations}
        repositories={[]}
        isLoading={false}
        error={null}
        onResultClick={mockOnResultClick}
      />
    );

    expect(screen.getByText("View all results")).toBeInTheDocument();
  });

  it("should show 'View all results' link with only repositories", () => {
    render(
      <SearchDropdown
        query="test"
        organizations={[]}
        repositories={mockRepositories}
        isLoading={false}
        error={null}
        onResultClick={mockOnResultClick}
      />
    );

    expect(screen.getByText("View all results")).toBeInTheDocument();
  });
});
