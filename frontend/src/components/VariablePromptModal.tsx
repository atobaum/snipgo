import { useState, useEffect } from "react";
import { Variable } from "../types";

interface VariablePromptModalProps {
  variables: Variable[];
  onSubmit: (values: Record<string, string>) => void;
  onCancel: () => void;
}

export function VariablePromptModal({
  variables,
  onSubmit,
  onCancel,
}: VariablePromptModalProps) {
  const [values, setValues] = useState<Record<string, string>>({});

  // Initialize values with defaults
  useEffect(() => {
    const initialValues: Record<string, string> = {};
    variables.forEach((variable) => {
      if (variable.default) {
        initialValues[variable.name] = variable.default;
      } else if (variable.choices && variable.choices.length > 0) {
        initialValues[variable.name] = variable.choices[0];
      } else {
        initialValues[variable.name] = "";
      }
    });
    setValues(initialValues);
  }, [variables]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(values);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Escape") {
      e.preventDefault();
      onCancel();
    } else if (e.key === "Enter" && e.target instanceof HTMLElement && e.target.tagName !== "INPUT" && e.target.tagName !== "SELECT") {
      e.preventDefault();
      onSubmit(values);
    }
  };

  return (
    <div
      className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
      onClick={onCancel}
      onKeyDown={handleKeyDown}
      tabIndex={-1}
    >
      <div
        className="bg-white rounded-lg shadow-xl max-w-md w-full mx-4"
        onClick={(e) => e.stopPropagation()}
      >
        <form onSubmit={handleSubmit}>
          {/* Header */}
          <div className="px-6 py-4 border-b border-gray-200">
            <h2 className="text-xl font-bold text-gray-800">
              Fill in Variables
            </h2>
          </div>

          {/* Body */}
          <div className="px-6 py-4 max-h-96 overflow-y-auto">
            <div className="space-y-4">
              {variables.map((variable) => (
                <VariableInput
                  key={variable.name}
                  variable={variable}
                  value={values[variable.name] || ""}
                  onChange={(value) =>
                    setValues((prev) => ({ ...prev, [variable.name]: value }))
                  }
                />
              ))}
            </div>
          </div>

          {/* Footer */}
          <div className="px-6 py-4 border-t border-gray-200 flex justify-end gap-2">
            <button
              type="button"
              onClick={onCancel}
              className="px-4 py-2 text-gray-700 bg-gray-100 rounded hover:bg-gray-200"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="px-4 py-2 text-white bg-blue-500 rounded hover:bg-blue-600"
            >
              Copy Expanded
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

interface VariableInputProps {
  variable: Variable;
  value: string;
  onChange: (value: string) => void;
}

function VariableInput({ variable, value, onChange }: VariableInputProps) {
  return (
    <div>
      <label
        htmlFor={variable.name}
        className="block text-sm font-bold text-gray-700 mb-1"
      >
        {variable.name}
      </label>
      {variable.description && (
        <p className="text-xs text-gray-500 mb-1">{variable.description}</p>
      )}
      {variable.choices ? (
        <select
          id={variable.name}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          {variable.choices.map((choice) => (
            <option key={choice} value={choice}>
              {choice}
            </option>
          ))}
        </select>
      ) : (
        <input
          type="text"
          id={variable.name}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
      )}
    </div>
  );
}
