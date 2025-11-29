import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { Pagination } from "./pagination";

describe("Pagination", () => {
  it("does not render when totalPages is 1 or less", () => {
    const onPageChange = vi.fn();

    const { container: container1 } = render(
      <Pagination currentPage={1} totalPages={1} onPageChange={onPageChange} />
    );

    const { container: container2 } = render(
      <Pagination currentPage={1} totalPages={0} onPageChange={onPageChange} />
    );

    expect(container1.firstChild).toBeNull();
    expect(container2.firstChild).toBeNull();
  });

  it("renders pagination buttons for multiple pages", () => {
    const onPageChange = vi.fn();

    render(
      <Pagination currentPage={1} totalPages={3} onPageChange={onPageChange} />
    );

    expect(screen.getByRole("button", { name: "1" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "2" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "3" })).toBeInTheDocument();
  });

  it("highlights the current page", () => {
    const onPageChange = vi.fn();

    render(
      <Pagination currentPage={2} totalPages={3} onPageChange={onPageChange} />
    );

    const currentPageButton = screen.getByRole("button", { name: "2" });
    expect(currentPageButton.className).toContain("bg-primary");
  });

  it("calls onPageChange when a page button is clicked", async () => {
    const user = userEvent.setup();
    const onPageChange = vi.fn();

    render(
      <Pagination currentPage={1} totalPages={3} onPageChange={onPageChange} />
    );

    await user.click(screen.getByRole("button", { name: "3" }));

    expect(onPageChange).toHaveBeenCalledWith(3);
  });

  it("handles next and previous buttons correctly", async () => {
    const user = userEvent.setup();
    const onPageChange = vi.fn();

    render(
      <Pagination currentPage={2} totalPages={5} onPageChange={onPageChange} />
    );

    const buttons = screen.getAllByRole("button");
    const nextButton = buttons[buttons.length - 1]; // Last button should be next
    await user.click(nextButton);
    expect(onPageChange).toHaveBeenCalledWith(3);

    const prevButton = buttons[0]; // First button should be previous
    await user.click(prevButton);
    expect(onPageChange).toHaveBeenCalledWith(1);
  });

  it("disables previous button on first page", () => {
    const onPageChange = vi.fn();

    render(
      <Pagination currentPage={1} totalPages={3} onPageChange={onPageChange} />
    );

    const buttons = screen.getAllByRole("button");
    const prevButton = buttons[0]; // First button should be previous
    expect(prevButton).toBeDisabled();
  });

  it("disables next button on last page", () => {
    const onPageChange = vi.fn();

    render(
      <Pagination currentPage={3} totalPages={3} onPageChange={onPageChange} />
    );

    const buttons = screen.getAllByRole("button");
    const nextButton = buttons[buttons.length - 1]; // Last button should be next
    expect(nextButton).toBeDisabled();
  });

  it("disables all buttons when disabled prop is true", () => {
    const onPageChange = vi.fn();

    render(
      <Pagination
        currentPage={2}
        totalPages={5}
        onPageChange={onPageChange}
        disabled={true}
      />
    );

    const buttons = screen.getAllByRole("button");
    buttons.forEach((button) => {
      expect(button).toBeDisabled();
    });
  });

  it("shows ellipsis for large page ranges", () => {
    const onPageChange = vi.fn();

    render(
      <Pagination currentPage={1} totalPages={10} onPageChange={onPageChange} />
    );

    expect(screen.getByRole("button", { name: "1" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "2" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "3" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "10" })).toBeInTheDocument();

    const ellipsisElements = document.querySelectorAll(
      'div[class*="flex h-8 w-8"]'
    );
    expect(ellipsisElements.length).toBeGreaterThan(0);
  });

  it("shows correct page range when in middle of pagination", () => {
    const onPageChange = vi.fn();

    render(
      <Pagination currentPage={5} totalPages={10} onPageChange={onPageChange} />
    );

    expect(screen.getByRole("button", { name: "1" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "3" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "4" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "5" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "6" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "7" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "10" })).toBeInTheDocument();
  });
});
