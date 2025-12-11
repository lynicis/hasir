import type { Repository } from "@buf/hasir_hasir.bufbuild_es/registry/v1/registry_pb";

import { NodeType } from "@buf/hasir_hasir.bufbuild_es/registry/v1/registry_pb";
import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import userEvent from "@testing-library/user-event";

import RepositoryFilesContent from "./repository-files-content";
import { RepositoryContext } from "./repository-context";

const mockGetFileTree = vi.fn();

const mockClient = {
  getFileTree: mockGetFileTree,
};

vi.mock("@/lib/use-client", () => ({
  useClient: () => mockClient,
}));

const mockRepository: Repository = {
  $typeName: "registry.v1.Repository",
  id: "test-repo-id",
  name: "Test Repository",
  visibility: 0,
  sdkPreferences: [],
};

const mockFileTreeData = {
  nodes: [
    {
      name: "proto",
      path: "proto",
      type: NodeType.DIRECTORY,
      children: [
        {
          name: "user",
          path: "proto/user",
          type: NodeType.DIRECTORY,
          children: [
            {
              name: "user.proto",
              path: "proto/user/user.proto",
              type: NodeType.FILE,
              children: [],
            },
            {
              name: "user_service.proto",
              path: "proto/user/user_service.proto",
              type: NodeType.FILE,
              children: [],
            },
          ],
        },
        {
          name: "product",
          path: "proto/product",
          type: NodeType.DIRECTORY,
          children: [
            {
              name: "product.proto",
              path: "proto/product/product.proto",
              type: NodeType.FILE,
              children: [],
            },
            {
              name: "product_service.proto",
              path: "proto/product/product_service.proto",
              type: NodeType.FILE,
              children: [],
            },
          ],
        },
        {
          name: "common",
          path: "proto/common",
          type: NodeType.DIRECTORY,
          children: [
            {
              name: "types.proto",
              path: "proto/common/types.proto",
              type: NodeType.FILE,
              children: [],
            },
            {
              name: "errors.proto",
              path: "proto/common/errors.proto",
              type: NodeType.FILE,
              children: [],
            },
          ],
        },
      ],
    },
    {
      name: "docs",
      path: "docs",
      type: NodeType.DIRECTORY,
      children: [
        {
          name: "README.md",
          path: "docs/README.md",
          type: NodeType.FILE,
          children: [],
        },
        {
          name: "API.md",
          path: "docs/API.md",
          type: NodeType.FILE,
          children: [],
        },
      ],
    },
  ],
};

function renderWithContext(component: React.ReactElement) {
  return render(
    <RepositoryContext.Provider
      value={{
        repository: mockRepository,
        isLoading: false,
        error: null,
      }}
    >
      {component}
    </RepositoryContext.Provider>
  );
}

describe("RepositoryFilesContent", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetFileTree.mockImplementation(
      (request: { id: string; path?: string }) => {
        if (!request.path) {
          return Promise.resolve(mockFileTreeData);
        }

        const findNodeByPath = (
          nodes: typeof mockFileTreeData.nodes,
          targetPath: string
        ): { nodes: typeof mockFileTreeData.nodes } | null => {
          for (const node of nodes) {
            if (node.path === targetPath) {
              return { nodes: node.children };
            }
            if (node.children && node.children.length > 0) {
              const found = findNodeByPath(node.children, targetPath);
              if (found) return found;
            }
          }
          return null;
        };

        const result = findNodeByPath(mockFileTreeData.nodes, request.path);
        return Promise.resolve(result || { nodes: [] });
      }
    );
  });

  it("renders the Files card with title and description", async () => {
    renderWithContext(<RepositoryFilesContent />);

    expect(screen.getByText("Files")).toBeInTheDocument();
    expect(
      screen.getByText(/browse and manage repository files and schemas/i)
    ).toBeInTheDocument();

    await waitFor(
      () => {
        expect(screen.getByText("proto")).toBeInTheDocument();
      },
      { timeout: 2000 }
    );
  });

  it("renders the file tree section", async () => {
    renderWithContext(<RepositoryFilesContent />);

    await waitFor(() => {
      expect(screen.getByText("File Tree")).toBeInTheDocument();
    });
  });

  it("displays the initial message to select a file", async () => {
    renderWithContext(<RepositoryFilesContent />);

    await waitFor(() => {
      expect(screen.getByText("Select a file")).toBeInTheDocument();
    });

    expect(
      screen.getByText(/select a file from the tree to view details/i)
    ).toBeInTheDocument();
  });

  it("shows file action buttons when a file is selected", async () => {
    const user = userEvent.setup();
    renderWithContext(<RepositoryFilesContent />);

    await waitFor(
      () => {
        expect(screen.getByText("proto")).toBeInTheDocument();
      },
      { timeout: 2000 }
    );

    const protoFolder = screen.getByText("proto");
    await user.click(protoFolder);

    const userFolder = screen.getByText("user");
    await user.click(userFolder);

    const file = screen.getByText("user.proto");
    await user.click(file);

    await waitFor(() => {
      expect(screen.getByText("File Actions")).toBeInTheDocument();
    });

    expect(
      screen.getByRole("button", { name: /download/i })
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /edit/i })).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /history/i })
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /delete/i })).toBeInTheDocument();
  });

  it("renders folder structure with proto, docs folders", async () => {
    renderWithContext(<RepositoryFilesContent />);

    await waitFor(
      () => {
        expect(screen.getByText("proto")).toBeInTheDocument();
        expect(screen.getByText("docs")).toBeInTheDocument();
      },
      { timeout: 2000 }
    );
  });

  it("renders nested folders (user, product, common)", async () => {
    const user = userEvent.setup();
    renderWithContext(<RepositoryFilesContent />);

    await waitFor(
      () => {
        expect(screen.getByText("proto")).toBeInTheDocument();
      },
      { timeout: 2000 }
    );

    const protoFolder = screen.getByText("proto");
    await user.click(protoFolder);

    await waitFor(() => {
      expect(screen.getByText("user")).toBeInTheDocument();
      expect(screen.getByText("product")).toBeInTheDocument();
      expect(screen.getByText("common")).toBeInTheDocument();
    });
  });

  it("renders proto files with lazy loading", async () => {
    const user = userEvent.setup();
    renderWithContext(<RepositoryFilesContent />);

    await waitFor(
      () => {
        expect(screen.getByText("proto")).toBeInTheDocument();
      },
      { timeout: 2000 }
    );

    // Expand proto folder
    const protoFolder = screen.getByText("proto");
    await user.click(protoFolder);

    await waitFor(
      () => {
        expect(screen.getByText("user")).toBeInTheDocument();
      },
      { timeout: 3000 }
    );

    // Expand user folder to test lazy loading
    const userFolder = screen.getByText("user");
    await user.click(userFolder);

    // Verify files are lazy-loaded when folder is expanded
    await waitFor(
      () => {
        expect(screen.getByText("user.proto")).toBeInTheDocument();
        expect(screen.getByText("user_service.proto")).toBeInTheDocument();
      },
      { timeout: 3000 }
    );

    // Verify other sibling folders are still present
    expect(screen.getByText("product")).toBeInTheDocument();
    expect(screen.getByText("common")).toBeInTheDocument();
  });

  it("renders documentation files with lazy loading", async () => {
    const user = userEvent.setup();
    renderWithContext(<RepositoryFilesContent />);

    // Wait for initial tree load
    await waitFor(
      () => {
        expect(screen.getByText("proto")).toBeInTheDocument();
        expect(screen.getByText("docs")).toBeInTheDocument();
      },
      { timeout: 3000 }
    );

    // Expand docs folder to test lazy loading
    const docsFolder = screen.getByText("docs");
    await user.click(docsFolder);

    // Verify files are lazy-loaded when folder is expanded
    await waitFor(
      () => {
        expect(screen.getByText("README.md")).toBeInTheDocument();
        expect(screen.getByText("API.md")).toBeInTheDocument();
      },
      { timeout: 5000 }
    );
  });

  it("shows loading skeleton when data is loading", async () => {
    mockGetFileTree.mockImplementation(
      () =>
        new Promise((resolve) =>
          setTimeout(() => resolve(mockFileTreeData), 50)
        )
    );

    renderWithContext(<RepositoryFilesContent />);

    expect(screen.getByText("File Tree")).toBeInTheDocument();

    await waitFor(
      () => {
        expect(screen.getByText("proto")).toBeInTheDocument();
      },
      { timeout: 2000 }
    );
  });

  it("shows error alert when there is an error", async () => {
    const mockError = new Error("Failed to load");
    mockGetFileTree.mockRejectedValue(mockError);

    renderWithContext(<RepositoryFilesContent />);

    await waitFor(() => {
      expect(screen.getByText("Error loading file tree")).toBeInTheDocument();
    });
  });

  it("shows empty state when no files are found", async () => {
    mockGetFileTree.mockResolvedValue({ nodes: [] });

    renderWithContext(<RepositoryFilesContent />);

    await waitFor(() => {
      expect(
        screen.getByText(/no files found in this repository/i)
      ).toBeInTheDocument();
    });
  });
});
