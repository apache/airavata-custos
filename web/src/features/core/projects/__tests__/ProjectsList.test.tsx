import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { ProjectsList } from "../components/ProjectsList";
import type { Project } from "../schemas";

const project: Project = {
  id: "project-001",
  originated_id: "BIO130000",
  title: "Genomic Sequencing Pipeline",
  origination: "ACCESS",
  project_pi_id: "user-pi-001",
  status: "ACTIVE",
  created_time: "2026-04-01T00:00:00.000Z",
};

function renderList(overrides: Partial<React.ComponentProps<typeof ProjectsList>> = {}) {
  const props = {
    rows: [project],
    isLoading: false,
    error: null,
    search: "",
    onSearchChange: vi.fn(),
    statusFilter: "all" as const,
    onStatusFilterChange: vi.fn(),
    ...overrides,
  };
  return { ...render(<ProjectsList {...props} />), props };
}

describe("<ProjectsList />", () => {
  it("renders the project title as a link to the detail page", () => {
    renderList();
    const link = screen.getByRole("link", { name: project.title });
    expect(link).toHaveAttribute("href", `/projects/${project.id}`);
  });

  it("shows the empty state when no rows match the filters", () => {
    renderList({ rows: [] });
    expect(screen.getByRole("heading", { name: /no projects yet/i })).toBeInTheDocument();
  });

  it("filters by status", () => {
    const inactive: Project = {
      ...project,
      id: "project-002",
      title: "Archived Project",
      status: "INACTIVE",
    };
    renderList({ rows: [project, inactive], statusFilter: "INACTIVE" });
    expect(screen.queryByRole("link", { name: project.title })).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: inactive.title })).toBeInTheDocument();
  });

  it("surfaces an error state with a retry callback", () => {
    const onRetry = vi.fn();
    renderList({ error: new Error("boom"), onRetry });
    expect(screen.getByText(/boom/)).toBeInTheDocument();
  });
});
