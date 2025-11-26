import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";

import { Dashboard } from "./dashboard";

describe("Dashboard", () => {
  it("renders organizations list with an option for all organizations", () => {
    render(<Dashboard />);

    expect(
      screen.getByRole("button", { name: /all organizations/i })
    ).toBeInTheDocument();

    expect(
      screen.getByRole("button", { name: "Acme Corp" })
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Hasir Labs" })
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Proto Systems" })
    ).toBeInTheDocument();
  });

  it("shows all repositories and correct count by default", () => {
    render(<Dashboard />);

    // There are 4 mock repositories total
    expect(screen.getByText("4 repos")).toBeInTheDocument();

    expect(screen.getByText("payment-protos")).toBeInTheDocument();
    expect(screen.getByText("analytics-protos")).toBeInTheDocument();
    expect(screen.getByText("core-registry")).toBeInTheDocument();
    expect(screen.getByText("internal-tools")).toBeInTheDocument();
  });

  it("filters repositories when an organization is selected", async () => {
    const user = userEvent.setup();

    render(<Dashboard />);

    await user.click(screen.getByRole("button", { name: "Hasir Labs" }));

    // Only repositories belonging to "Hasir Labs" should be visible
    expect(screen.getByText("analytics-protos")).toBeInTheDocument();
    expect(screen.getByText("internal-tools")).toBeInTheDocument();

    expect(screen.queryByText("payment-protos")).not.toBeInTheDocument();
    expect(screen.queryByText("core-registry")).not.toBeInTheDocument();

    // Count should reflect filtered repositories
    expect(screen.getByText("2 repos")).toBeInTheDocument();

    // Helper text about which organization is being shown should appear
    expect(
      screen.getByText(/showing repositories in Hasir Labs/i)
    ).toBeInTheDocument();
  });

  it("resets to all repositories when All organizations is selected", async () => {
    const user = userEvent.setup();

    render(<Dashboard />);

    // First filter down to a single organization
    await user.click(screen.getByRole("button", { name: "Acme Corp" }));
    expect(screen.getByText("payment-protos")).toBeInTheDocument();
    expect(screen.queryByText("analytics-protos")).not.toBeInTheDocument();
    expect(screen.getByText("1 repos")).toBeInTheDocument();

    // Then go back to all organizations
    await user.click(
      screen.getByRole("button", { name: /all organizations/i })
    );

    expect(screen.getByText("payment-protos")).toBeInTheDocument();
    expect(screen.getByText("analytics-protos")).toBeInTheDocument();
    expect(screen.getByText("core-registry")).toBeInTheDocument();
    expect(screen.getByText("internal-tools")).toBeInTheDocument();
    expect(screen.getByText("4 repos")).toBeInTheDocument();
  });
});
