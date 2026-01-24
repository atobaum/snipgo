import { javascript } from "@codemirror/lang-javascript";
import { python } from "@codemirror/lang-python";
import { yaml } from "@codemirror/lang-yaml";
import { json } from "@codemirror/lang-json";
import { markdown } from "@codemirror/lang-markdown";
import type { Extension } from "@codemirror/state";

export const AUTO_SAVE_DELAY = 1500; // 1.5 seconds

export const SUPPORTED_LANGUAGES = [
  "plaintext",
  "javascript",
  "typescript",
  "python",
  "yaml",
  "json",
  "markdown",
  "bash",
  "go",
  "rust",
  "java",
  "cpp",
  "c",
  "html",
  "css",
  "sql",
];

export const languageExtensions: Record<string, Extension> = {
  javascript: javascript(),
  typescript: javascript({ jsx: true }),
  python: python(),
  yaml: yaml(),
  json: json(),
  markdown: markdown(),
};
