import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";

import RepositoryFilesContent from "./repository-files-content";

describe("RepositoryFilesContent", () => {
  it("renders the Files card with title and description", () => {
    render(<RepositoryFilesContent />);

    expect(screen.getByText("Files")).toBeInTheDocument();
    expect(
      screen.getByText(/browse and manage repository files and schemas/i)
    ).toBeInTheDocument();
  });

  it("renders the file tree section", () => {
    render(<RepositoryFilesContent />);

    expect(screen.getByText("File Tree")).toBeInTheDocument();
  });

  it("displays the initial message to select a file", () => {
    render(<RepositoryFilesContent />);

    expect(screen.getByText("Select a file")).toBeInTheDocument();
    expect(
      screen.getByText(/select a file from the tree to view details/i)
    ).toBeInTheDocument();
  });

  it("shows file action buttons when a file is selected", async () => {
    const user = userEvent.setup();
    render(<RepositoryFilesContent />);

    // First expand the proto folder
    const protoFolder = screen.getByText("proto");
    await user.click(protoFolder);

    // Then expand the user folder
    const userFolder = screen.getByText("user");
    await user.click(userFolder);

    // Find and click a file in the tree (user.proto)
    const file = screen.getByText("user.proto");
    await user.click(file);

    // Wait for file actions to appear after state update
    await waitFor(() => {
      expect(screen.getByText("File Actions")).toBeInTheDocument();
    });

    expect(screen.getByRole("button", { name: /download/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /edit/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /history/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /delete/i })).toBeInTheDocument();
  });

  it("renders folder structure with proto, docs folders", () => {
    render(<RepositoryFilesContent />);

    expect(screen.getByText("proto")).toBeInTheDocument();
    expect(screen.getByText("docs")).toBeInTheDocument();
  });

  it("renders nested folders (user, product, common)", async () => {
    const user = userEvent.setup();
    render(<RepositoryFilesContent />);

    // Expand the proto folder to see nested folders
    const protoFolder = screen.getByText("proto");
    await user.click(protoFolder);

    expect(screen.getByText("user")).toBeInTheDocument();
    expect(screen.getByText("product")).toBeInTheDocument();
    expect(screen.getByText("common")).toBeInTheDocument();
  });

  it("renders proto files", async () => {
    const user = userEvent.setup();
    render(<RepositoryFilesContent />);

    // Expand proto folder
    const protoFolder = screen.getByText("proto");
    await user.click(protoFolder);

    // Expand user folder
    const userFolder = screen.getByText("user");
    await user.click(userFolder);

    expect(screen.getByText("user.proto")).toBeInTheDocument();
    expect(screen.getByText("user_service.proto")).toBeInTheDocument();

    // Expand product folder
    const productFolder = screen.getByText("product");
    await user.click(productFolder);

    expect(screen.getByText("product.proto")).toBeInTheDocument();
    expect(screen.getByText("product_service.proto")).toBeInTheDocument();

    // Expand common folder
    const commonFolder = screen.getByText("common");
    await user.click(commonFolder);

    expect(screen.getByText("types.proto")).toBeInTheDocument();
    expect(screen.getByText("errors.proto")).toBeInTheDocument();
  });

  it("renders documentation files", async () => {
    const user = userEvent.setup();
    render(<RepositoryFilesContent />);

    // Expand docs folder
    const docsFolder = screen.getByText("docs");
    await user.click(docsFolder);

    expect(screen.getByText("README.md")).toBeInTheDocument();
    expect(screen.getByText("API.md")).toBeInTheDocument();
  });
});
