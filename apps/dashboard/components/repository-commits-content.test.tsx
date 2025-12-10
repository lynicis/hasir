import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createConnectTransport } from "@connectrpc/connect-web";
import { TransportProvider } from "@connectrpc/connect-query";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { Code, ConnectError } from "@connectrpc/connect";
import { render, screen } from "@testing-library/react";

import RepositoryCommitsContent from "./repository-commits-content";

const { mockUseQuery, defaultMockCommits } = vi.hoisted(() => {
  const createMockTimestamp = (daysAgo: number) => {
    const date = new Date();
    date.setDate(date.getDate() - daysAgo);
    const seconds = BigInt(Math.floor(date.getTime() / 1000));
    return {
      seconds,
      nanos: 0,
    };
  };

  const defaultMockCommits = {
    data: {
      commits: [
        {
          id: "a1b2c3d1234567890",
          message: "feat: Add user authentication schema",
          user: { username: "John Doe" },
          commitedAt: createMockTimestamp(1),
        },
        {
          id: "e4f5g6h1234567890",
          message: "fix: Update product service definitions",
          user: { username: "Jane Smith" },
          commitedAt: createMockTimestamp(2),
        },
        {
          id: "i7j8k9l1234567890",
          message: "refactor: Reorganize proto directory structure",
          user: { username: "Bob Johnson" },
          commitedAt: createMockTimestamp(3),
        },
        {
          id: "m0n1o2p1234567890",
          message: "docs: Update API documentation",
          user: { username: "John Doe" },
          commitedAt: createMockTimestamp(4),
        },
        {
          id: "q3r4s5t1234567890",
          message: "test: Add integration tests",
          user: { username: "Jane Smith" },
          commitedAt: createMockTimestamp(5),
        },
        {
          id: "u6v7w8x1234567890",
          message: "chore: Update dependencies",
          user: { username: "Bob Johnson" },
          commitedAt: createMockTimestamp(6),
        },
        {
          id: "y9z0a1b1234567890",
          message: "feat: Add new endpoint",
          user: { username: "John Doe" },
          commitedAt: createMockTimestamp(7),
        },
        {
          id: "c2d3e4f1234567890",
          message: "fix: Resolve merge conflict",
          user: { username: "Jane Smith" },
          commitedAt: createMockTimestamp(8),
        },
      ],
    },
    isLoading: false,
    error: null,
  };

  const mockUseQuery = vi.fn(() => defaultMockCommits);

  return { mockUseQuery, defaultMockCommits };
});

vi.mock("@connectrpc/connect-query", async () => {
  const actual = await vi.importActual("@connectrpc/connect-query");
  return {
    ...actual,
    useQuery: mockUseQuery,
  };
});

describe("RepositoryCommitsContent", () => {
  let queryClient: QueryClient;
  const transport = createConnectTransport({
    baseUrl: "http://localhost:3000",
  });

  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <TransportProvider transport={transport}>
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    </TransportProvider>
  );

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
      },
    });
    mockUseQuery.mockReturnValue(defaultMockCommits);
  });

  it("renders the commit history card with title and description", () => {
    render(<RepositoryCommitsContent repositoryId="test-repo-id" />, {
      wrapper,
    });

    expect(screen.getByText("Commit History")).toBeInTheDocument();
    expect(
      screen.getByText(/view the complete commit history for this repository/i)
    ).toBeInTheDocument();
  });

  it("renders multiple commits with author information", () => {
    render(<RepositoryCommitsContent repositoryId="test-repo-id" />, {
      wrapper,
    });

    expect(screen.getAllByText("John Doe").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Jane Smith").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Bob Johnson").length).toBeGreaterThan(0);
  });

  it("renders commit messages", () => {
    render(<RepositoryCommitsContent repositoryId="test-repo-id" />, {
      wrapper,
    });

    expect(
      screen.getByText("feat: Add user authentication schema")
    ).toBeInTheDocument();
    expect(
      screen.getByText("fix: Update product service definitions")
    ).toBeInTheDocument();
    expect(
      screen.getByText("refactor: Reorganize proto directory structure")
    ).toBeInTheDocument();
  });

  it("renders commit hashes", () => {
    render(<RepositoryCommitsContent repositoryId="test-repo-id" />, {
      wrapper,
    });

    expect(screen.getByText("a1b2c3d")).toBeInTheDocument();
    expect(screen.getByText("e4f5g6h")).toBeInTheDocument();
    expect(screen.getByText("i7j8k9l")).toBeInTheDocument();
  });

  it("renders avatars with initials for each author", () => {
    render(<RepositoryCommitsContent repositoryId="test-repo-id" />, {
      wrapper,
    });

    // Check for avatar fallbacks with initials (multiple occurrences expected)
    expect(screen.getAllByText("JD").length).toBeGreaterThan(0); // John Doe
    expect(screen.getAllByText("JS").length).toBeGreaterThan(0); // Jane Smith
    expect(screen.getAllByText("BJ").length).toBeGreaterThan(0); // Bob Johnson
  });

  it("displays a note about sample data", () => {
    render(<RepositoryCommitsContent repositoryId="test-repo-id" />, {
      wrapper,
    });

    expect(
      screen.queryByText(/this is a preview with sample commit data/i)
    ).not.toBeInTheDocument();
    expect(
      screen.queryByText(/once the backend commit history api is available/i)
    ).not.toBeInTheDocument();
  });

  it("renders at least 8 commits", () => {
    render(<RepositoryCommitsContent repositoryId="test-repo-id" />, {
      wrapper,
    });

    // Count the number of commit hash codes rendered
    const commitHashes = [
      "a1b2c3d",
      "e4f5g6h",
      "i7j8k9l",
      "m0n1o2p",
      "q3r4s5t",
      "u6v7w8x",
      "y9z0a1b",
      "c2d3e4f",
    ];

    commitHashes.forEach((hash) => {
      expect(screen.getByText(hash)).toBeInTheDocument();
    });
  });

  it("renders commit history icon", () => {
    const { container } = render(
      <RepositoryCommitsContent repositoryId="test-repo-id" />,
      { wrapper }
    );

    // GitCommit icon should be present
    const svgElements = container.querySelectorAll("svg");
    expect(svgElements.length).toBeGreaterThan(0);
  });

  it("shows 'no commits found' message when 404 error is returned", () => {
    mockUseQuery.mockReturnValueOnce({
      data: undefined,
      isLoading: false,
      error: new ConnectError("Repository not found", Code.NotFound),
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } as any);

    render(<RepositoryCommitsContent repositoryId="test-repo-id" />, {
      wrapper,
    });

    expect(
      screen.getByText("No commits found for this repository.")
    ).toBeInTheDocument();
    expect(screen.queryByText("Error loading commits")).not.toBeInTheDocument();
  });

  it("shows 'no commits found' message when empty commits array is returned", () => {
    mockUseQuery.mockReturnValueOnce({
      data: { commits: [] },
      isLoading: false,
      error: null,
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } as any);

    render(<RepositoryCommitsContent repositoryId="test-repo-id" />, {
      wrapper,
    });

    expect(
      screen.getByText("No commits found for this repository.")
    ).toBeInTheDocument();
  });

  it("shows error alert for non-404 errors", () => {
    mockUseQuery.mockReturnValueOnce({
      data: undefined,
      isLoading: false,
      error: new ConnectError("Internal server error", Code.Internal),
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } as any);

    render(<RepositoryCommitsContent repositoryId="test-repo-id" />, {
      wrapper,
    });

    expect(screen.getByText("Error loading commits")).toBeInTheDocument();
    expect(screen.getByText(/internal.*server error/i)).toBeInTheDocument();
    expect(
      screen.queryByText("No commits found for this repository.")
    ).not.toBeInTheDocument();
  });
});
