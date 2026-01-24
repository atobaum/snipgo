import CodeMirror from "@uiw/react-codemirror";
import { EditorView } from "@codemirror/view";
import { languageExtensions } from "../constants/editor";

interface CodeEditorProps {
  body: string;
  onBodyChange: (value: string) => void;
  language: string;
  wordWrap: boolean;
  rawMode: boolean;
  rawContent: string;
}

export function CodeEditor({
  body,
  onBodyChange,
  language,
  wordWrap,
  rawMode,
  rawContent,
}: CodeEditorProps) {
  const languageExtension = language
    ? languageExtensions[language.toLowerCase()]
    : undefined;

  if (rawMode) {
    return (
      <textarea
        value={rawContent}
        readOnly
        className="w-full h-full p-4 font-mono text-sm border-none outline-none resize-none"
        style={{ whiteSpace: wordWrap ? "pre-wrap" : "pre" }}
      />
    );
  }

  return (
    <CodeMirror
      value={body}
      onChange={onBodyChange}
      extensions={[
        ...(languageExtension ? [languageExtension] : []),
        ...(wordWrap ? [EditorView.lineWrapping] : []),
        EditorView.theme({
          "&": { height: "100%" },
          ".cm-scroller": { overflow: "auto" },
        }),
      ]}
      theme="light"
      basicSetup={{
        lineNumbers: true,
        foldGutter: true,
        dropCursor: false,
        allowMultipleSelections: false,
      }}
      style={{ height: "100%" }}
    />
  );
}
