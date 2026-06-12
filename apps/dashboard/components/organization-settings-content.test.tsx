import { render, screen } from "@testing-library/react";

import OrganizationSettingsContent from "./organization-settings-content";

describe("OrganizationSettingsContent", () => {
  it("renders the page heading", () => {
    render(<OrganizationSettingsContent />);

    expect(
      screen.getByRole("heading", { name: /organization settings/i })
    ).toBeInTheDocument();
  });

  it("renders the page description", () => {
    render(<OrganizationSettingsContent />);

    expect(
      screen.getByText(/manage your organization details and preferences/i)
    ).toBeInTheDocument();
  });

  it("renders the organization settings form", () => {
    render(<OrganizationSettingsContent />);

    expect(screen.getByText("General Settings")).toBeInTheDocument();
  });
});
