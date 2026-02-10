import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { VariablePromptModal } from "./VariablePromptModal";
import { Variable } from "../types";

describe("VariablePromptModal", () => {
  const mockVariables: Variable[] = [
    { name: "username", description: "GitHub username", default: "johndoe" },
    { name: "environment", choices: ["dev", "staging", "prod"] },
    { name: "port", default: "8080" },
  ];

  const defaultProps = {
    variables: mockVariables,
    onSubmit: vi.fn(),
    onCancel: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders all variable inputs", () => {
    render(<VariablePromptModal {...defaultProps} />);

    expect(screen.getByText("Fill in Variables")).toBeInTheDocument();
    expect(screen.getByLabelText("username")).toBeInTheDocument();
    expect(screen.getByLabelText("environment")).toBeInTheDocument();
    expect(screen.getByLabelText("port")).toBeInTheDocument();
  });

  it("shows description as helper text", () => {
    render(<VariablePromptModal {...defaultProps} />);
    expect(screen.getByText("GitHub username")).toBeInTheDocument();
  });

  it("pre-fills default values in text inputs", () => {
    render(<VariablePromptModal {...defaultProps} />);

    const usernameInput = screen.getByLabelText("username") as HTMLInputElement;
    const portInput = screen.getByLabelText("port") as HTMLInputElement;

    expect(usernameInput.value).toBe("johndoe");
    expect(portInput.value).toBe("8080");
  });

  it("shows dropdown for variables with choices", () => {
    render(<VariablePromptModal {...defaultProps} />);

    const envSelect = screen.getByLabelText("environment") as HTMLSelectElement;
    expect(envSelect.tagName).toBe("SELECT");

    const options = screen.getAllByRole("option");
    expect(options).toHaveLength(3);
    expect(screen.getByText("dev")).toBeInTheDocument();
    expect(screen.getByText("staging")).toBeInTheDocument();
    expect(screen.getByText("prod")).toBeInTheDocument();
  });

  it("submits with all values when Copy Expanded is clicked", async () => {
    const user = userEvent.setup();
    render(<VariablePromptModal {...defaultProps} />);

    const usernameInput = screen.getByLabelText("username") as HTMLInputElement;
    await user.clear(usernameInput);
    await user.type(usernameInput, "alice");

    const envSelect = screen.getByLabelText("environment") as HTMLSelectElement;
    await user.selectOptions(envSelect, "prod");

    const submitButton = screen.getByText("Copy Expanded");
    await user.click(submitButton);

    await waitFor(() => {
      expect(defaultProps.onSubmit).toHaveBeenCalledWith({
        username: "alice",
        environment: "prod",
        port: "8080",
      });
    });
  });

  it("calls onCancel when Cancel button is clicked", async () => {
    const user = userEvent.setup();
    render(<VariablePromptModal {...defaultProps} />);

    const cancelButton = screen.getByText("Cancel");
    await user.click(cancelButton);

    expect(defaultProps.onCancel).toHaveBeenCalled();
    expect(defaultProps.onSubmit).not.toHaveBeenCalled();
  });

  it("handles variables without defaults", () => {
    const varsWithoutDefaults: Variable[] = [
      { name: "apiKey" },
      { name: "region", choices: ["us-east", "us-west"] },
    ];

    render(
      <VariablePromptModal
        {...defaultProps}
        variables={varsWithoutDefaults}
      />
    );

    const apiKeyInput = screen.getByLabelText("apiKey") as HTMLInputElement;
    expect(apiKeyInput.value).toBe("");

    const regionSelect = screen.getByLabelText("region") as HTMLSelectElement;
    expect(regionSelect.value).toBe("us-east");
  });
});
