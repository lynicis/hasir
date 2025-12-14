import type { Repository } from "@buf/hasir_hasir.bufbuild_es/registry/v1/registry_pb";

import { NodeType } from "@buf/hasir_hasir.bufbuild_es/registry/v1/registry_pb";
import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import userEvent from "@testing-library/user-event";

import { RepositoryContext } from "@/lib/repository-context";

import RepositoryFilesContent from "./repository-files-content";

vi.mock("react-syntax-highlighter", () => ({
  Prism: ({ children }: { children: string }) => <pre>{children}</pre>,
}));

vi.mock("react-syntax-highlighter/dist/esm/styles/prism", () => ({
  vscDarkPlus: {},
}));

const mockGetFileTree = vi.fn();
const mockGetFilePreview = vi.fn();

const mockClient = {
  getFileTree: mockGetFileTree,
  getFilePreview: mockGetFilePreview,
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

  it("shows file preview when a file is selected", async () => {
    const user = userEvent.setup();
    const mockFileContent =
      'syntax = "proto3";\n\nmessage User {\n  string id = 1;\n}';
    mockGetFilePreview.mockResolvedValue({
      content: mockFileContent,
      mimeType: "text/plain",
      size: BigInt(mockFileContent.length),
    });

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
      expect(mockGetFilePreview).toHaveBeenCalledWith({
        id: "test-repo-id",
        path: "proto/user/user.proto",
      });
    });

    await waitFor(() => {
      expect(screen.getByText(/syntax = "proto3"/)).toBeInTheDocument();
      expect(screen.getByText(/message User/)).toBeInTheDocument();
    });
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

    const protoFolder = screen.getByText("proto");
    await user.click(protoFolder);

    await waitFor(
      () => {
        expect(screen.getByText("user")).toBeInTheDocument();
      },
      { timeout: 3000 }
    );

    const userFolder = screen.getByText("user");
    await user.click(userFolder);

    await waitFor(
      () => {
        expect(screen.getByText("user.proto")).toBeInTheDocument();
        expect(screen.getByText("user_service.proto")).toBeInTheDocument();
      },
      { timeout: 3000 }
    );

    expect(screen.getByText("product")).toBeInTheDocument();
    expect(screen.getByText("common")).toBeInTheDocument();
  });

  it("renders documentation files with lazy loading", async () => {
    const user = userEvent.setup();
    renderWithContext(<RepositoryFilesContent />);

    await waitFor(
      () => {
        expect(screen.getByText("proto")).toBeInTheDocument();
        expect(screen.getByText("docs")).toBeInTheDocument();
      },
      { timeout: 3000 }
    );

    const docsFolder = screen.getByText("docs");
    await user.click(docsFolder);

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

  it("shows loading skeleton while fetching file preview", async () => {
    const user = userEvent.setup();
    mockGetFilePreview.mockImplementation(
      () =>
        new Promise((resolve) =>
          setTimeout(
            () =>
              resolve({
                content: "test content",
                mimeType: "text/plain",
                size: BigInt(12),
              }),
            100
          )
        )
    );

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

    expect(document.querySelector(".animate-pulse")).toBeInTheDocument();

    await waitFor(
      () => {
        expect(screen.getByText("test content")).toBeInTheDocument();
      },
      { timeout: 2000 }
    );
  });

  it("shows error alert when file preview fails to load", async () => {
    const user = userEvent.setup();
    const mockError = new Error("Failed to load file preview");
    mockGetFilePreview.mockRejectedValue(mockError);

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
      expect(
        screen.getByText("Error loading file preview")
      ).toBeInTheDocument();
    });
  });

  it("shows appropriate message for non-text files", async () => {
    const user = userEvent.setup();
    mockGetFilePreview.mockResolvedValue({
      content: "binary content",
      mimeType: "image/png",
      size: BigInt(1024),
    });

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
      expect(
        screen.getByText(/preview not available for this file type/i)
      ).toBeInTheDocument();
    });
  });

  it("displays file metadata (size and mime type)", async () => {
    const user = userEvent.setup();
    const fileSize = BigInt(2048);
    mockGetFilePreview.mockResolvedValue({
      content: "test content",
      mimeType: "text/plain",
      size: fileSize,
    });

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
      expect(screen.getByText(/2\.00 KB/)).toBeInTheDocument();
      expect(screen.getByText(/text\/plain/)).toBeInTheDocument();
    });
  });
});
