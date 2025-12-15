import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { MarkdownRenderer } from "./markdown-renderer";

describe("MarkdownRenderer", () => {
  describe("basic rendering", () => {
    it("renders plain text content", () => {
      render(<MarkdownRenderer content="Hello World" />);

      expect(screen.getByText("Hello World")).toBeInTheDocument();
    });

    it("applies prose classes to container", () => {
      const { container } = render(<MarkdownRenderer content="Test" />);

      const wrapper = container.firstChild as HTMLElement;
      expect(wrapper).toHaveClass("prose", "prose-sm", "dark:prose-invert");
    });

    it("accepts custom className", () => {
      const { container } = render(
        <MarkdownRenderer content="Test" className="custom-class" />
      );

      const wrapper = container.firstChild as HTMLElement;
      expect(wrapper).toHaveClass("custom-class");
    });
  });

  describe("headings", () => {
    it("renders h1 headings", () => {
      render(<MarkdownRenderer content="# Heading 1" />);

      const heading = screen.getByRole("heading", { level: 1 });
      expect(heading).toHaveTextContent("Heading 1");
    });

    it("renders h2 headings", () => {
      render(<MarkdownRenderer content="## Heading 2" />);

      const heading = screen.getByRole("heading", { level: 2 });
      expect(heading).toHaveTextContent("Heading 2");
    });

    it("renders h3 headings", () => {
      render(<MarkdownRenderer content="### Heading 3" />);

      const heading = screen.getByRole("heading", { level: 3 });
      expect(heading).toHaveTextContent("Heading 3");
    });
  });

  describe("links", () => {
    it("renders links with correct href", () => {
      render(<MarkdownRenderer content="[Click here](https://example.com)" />);

      const link = screen.getByRole("link", { name: "Click here" });
      expect(link).toHaveAttribute("href", "https://example.com");
    });

    it("applies link styling classes", () => {
      render(<MarkdownRenderer content="[Link](https://example.com)" />);

      const link = screen.getByRole("link");
      expect(link).toHaveClass("text-primary", "hover:underline");
    });
  });

  describe("tables", () => {
    const tableMarkdown = `
| Header 1 | Header 2 |
| -------- | -------- |
| Cell 1   | Cell 2   |
`;

    it("renders tables", () => {
      render(<MarkdownRenderer content={tableMarkdown} />);

      expect(screen.getByRole("table")).toBeInTheDocument();
    });

    it("renders table headers", () => {
      render(<MarkdownRenderer content={tableMarkdown} />);

      expect(screen.getByText("Header 1")).toBeInTheDocument();
      expect(screen.getByText("Header 2")).toBeInTheDocument();
    });

    it("renders table cells", () => {
      render(<MarkdownRenderer content={tableMarkdown} />);

      expect(screen.getByText("Cell 1")).toBeInTheDocument();
      expect(screen.getByText("Cell 2")).toBeInTheDocument();
    });
  });

  describe("code blocks", () => {
    it("renders inline code", () => {
      render(<MarkdownRenderer content="Use `const` for constants" />);

      const code = screen.getByText("const");
      expect(code.tagName).toBe("CODE");
    });

    it("renders fenced code blocks with syntax highlighting", () => {
      const codeBlock = `
\`\`\`javascript
const x = 1;
\`\`\`
`;
      render(<MarkdownRenderer content={codeBlock} />);

      expect(screen.getByText("const")).toBeInTheDocument();
      expect(screen.getByText("1")).toBeInTheDocument();
    });
  });

  describe("HTML passthrough", () => {
    it("renders raw HTML anchor tags", () => {
      render(<MarkdownRenderer content='<a name="top"></a>' />);

      const anchor = document.querySelector('a[name="top"]');
      expect(anchor).toBeInTheDocument();
    });

    it("renders HTML with alignment", () => {
      render(<MarkdownRenderer content='<p align="right">Aligned text</p>' />);

      const paragraph = screen.getByText("Aligned text");
      expect(paragraph).toHaveAttribute("align", "right");
    });
  });

  describe("GitHub Flavored Markdown", () => {
    it("renders strikethrough text", () => {
      render(<MarkdownRenderer content="~~deleted~~" />);

      const del = screen.getByText("deleted");
      expect(del.tagName).toBe("DEL");
    });

    it("renders task list items", () => {
      const taskList = `- [x] Done
- [ ] Todo`;
      render(<MarkdownRenderer content={taskList} />);

      const checkboxes = screen.getAllByRole("checkbox");
      expect(checkboxes).toHaveLength(2);
      expect(checkboxes[0]).toBeChecked();
      expect(checkboxes[1]).not.toBeChecked();
    });
  });
});
