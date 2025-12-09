import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { ModeToggle } from "./theme-toggle";

const mockSetTheme = vi.fn();

vi.mock("next-themes", () => ({
  useTheme: () => ({
    setTheme: mockSetTheme,
    theme: "light",
  }),
}));

describe("ModeToggle", () => {
  it("renders the theme toggle button", () => {
    render(<ModeToggle />);

    const button = screen.getByRole("button", { name: /toggle theme/i });
    expect(button).toBeInTheDocument();
  });

  it("opens the dropdown menu when clicked", async () => {
    const user = userEvent.setup();
    render(<ModeToggle />);

    const button = screen.getByRole("button", { name: /toggle theme/i });
    await user.click(button);

    expect(screen.getByText("Light")).toBeInTheDocument();
    expect(screen.getByText("Dark")).toBeInTheDocument();
    expect(screen.getByText("System")).toBeInTheDocument();
  });

  it("calls setTheme with 'light' when Light is clicked", async () => {
    const user = userEvent.setup();
    render(<ModeToggle />);

    const button = screen.getByRole("button", { name: /toggle theme/i });
    await user.click(button);

    const lightOption = screen.getByText("Light");
    await user.click(lightOption);

    expect(mockSetTheme).toHaveBeenCalledWith("light");
  });

  it("calls setTheme with 'dark' when Dark is clicked", async () => {
    const user = userEvent.setup();
    render(<ModeToggle />);

    const button = screen.getByRole("button", { name: /toggle theme/i });
    await user.click(button);

    const darkOption = screen.getByText("Dark");
    await user.click(darkOption);

    expect(mockSetTheme).toHaveBeenCalledWith("dark");
  });

  it("calls setTheme with 'system' when System is clicked", async () => {
    const user = userEvent.setup();
    render(<ModeToggle />);

    const button = screen.getByRole("button", { name: /toggle theme/i });
    await user.click(button);

    const systemOption = screen.getByText("System");
    await user.click(systemOption);

    expect(mockSetTheme).toHaveBeenCalledWith("system");
  });
});
