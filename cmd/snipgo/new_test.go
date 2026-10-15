package main

import (
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestScanMultiLineFromReader(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single line with two blank lines",
			input:    "echo hello\n\n\n",
			expected: "echo hello",
		},
		{
			name:     "multiple lines with two blank lines to finish",
			input:    "docker system prune -af\ndocker volume prune -f\ndocker network prune -f\n\n\n",
			expected: "docker system prune -af\ndocker volume prune -f\ndocker network prune -f",
		},
		{
			name:     "lines with one blank line in between",
			input:    "line1\n\nline2\n\n\n",
			expected: "line1\n\nline2",
		},
		{
			name:     "empty input (just two blank lines)",
			input:    "\n\n",
			expected: "",
		},
		{
			name:     "input with trailing content after two blanks",
			input:    "first\nsecond\n\n\nignored",
			expected: "first\nsecond",
		},
		{
			name:     "three lines no intermediate blanks",
			input:    "line1\nline2\nline3\n\n\n",
			expected: "line1\nline2\nline3",
		},
		{
			name:     "single blank line does not terminate",
			input:    "before\n\nafter\n\n\n",
			expected: "before\n\nafter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			result, err := scanMultiLineFromReader(reader, "")
			if err != nil {
				t.Errorf("scanMultiLineFromReader() error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("scanMultiLineFromReader() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestScanMultiLineFromReader_EOF(t *testing.T) {
	// Test EOF without two blank lines (user sends EOF with Ctrl+D)
	input := "line1\nline2"
	reader := strings.NewReader(input)
	result, err := scanMultiLineFromReader(reader, "")
	if err != nil {
		t.Errorf("scanMultiLineFromReader() error = %v", err)
		return
	}
	if result != "line1\nline2" {
		t.Errorf("scanMultiLineFromReader() = %q, want %q", result, "line1\nline2")
	}
}

func TestParseTagsFlag(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{name: "empty", input: "", want: []string{}},
		{name: "single", input: "net", want: []string{"net"}},
		{name: "trims spaces", input: " net , mac ", want: []string{"net", "mac"}},
		{name: "drops empties", input: "net,,mac,", want: []string{"net", "mac"}},
		{name: "dedupes preserving order", input: "mac,net,mac,go,net", want: []string{"mac", "net", "go"}},
		{name: "only separators", input: " , ,", want: []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTagsFlag(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseTagsFlag(%q) = %#v, want %#v", tt.input, got, tt.want)
			}
		})
	}
}

func TestTrimSingleTrailingNewline(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "", want: ""},
		{input: "ls", want: "ls"},
		{input: "ls\n", want: "ls"},
		{input: "ls\n\n", want: "ls\n"},
		{input: "  indented\n", want: "  indented"},
		{input: "a\nb\n", want: "a\nb"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := trimSingleTrailingNewline(tt.input); got != tt.want {
				t.Errorf("trimSingleTrailingNewline(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestChooseBodySource(t *testing.T) {
	tests := []struct {
		name       string
		opts       newOptions
		stdinPiped bool
		want       bodySource
		wantErr    string
	}{
		{name: "interactive by default", want: bodyFromPrompt},
		{name: "piped stdin", stdinPiped: true, want: bodyFromStdin},
		{name: "clipboard", opts: newOptions{clipboard: true}, want: bodyFromClipboard},
		{name: "clipboard with editor", opts: newOptions{clipboard: true, editor: true}, want: bodyFromClipboard},
		{name: "stdin with editor", opts: newOptions{editor: true}, stdinPiped: true, want: bodyFromStdin},
		{name: "multiline and editor", opts: newOptions{multiline: true, editor: true}, wantErr: "mutually exclusive"},
		{name: "clipboard and piped stdin", opts: newOptions{clipboard: true}, stdinPiped: true, wantErr: "--clipboard"},
		{name: "clipboard and multiline", opts: newOptions{clipboard: true, multiline: true}, wantErr: "--multiline"},
		{name: "piped stdin and multiline", opts: newOptions{multiline: true}, stdinPiped: true, wantErr: "--multiline"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := chooseBodySource(tt.opts, tt.stdinPiped)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("chooseBodySource() error = %v, want containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("chooseBodySource() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("chooseBodySource() = %v, want %v", got, tt.want)
			}
		})
	}
}

type fakeNewIO struct {
	stdin       string
	stdinPiped  bool
	clipboard   string
	clipErr     error
	tty         string
	noTTY       bool
	lines       []string
	editResult  func(content string) string
	editedInput string
}

func (f *fakeNewIO) build() newIO {
	return newIO{
		stdin:      strings.NewReader(f.stdin),
		stdinPiped: f.stdinPiped,
		readClipboard: func() (string, error) {
			return f.clipboard, f.clipErr
		},
		openTTY: func() (io.ReadCloser, error) {
			if f.noTTY {
				return nil, errors.New("no tty")
			}
			return io.NopCloser(strings.NewReader(f.tty)), nil
		},
		promptLine: func(string) (string, error) {
			if len(f.lines) == 0 {
				return "", io.EOF
			}
			line := f.lines[0]
			f.lines = f.lines[1:]
			return line, nil
		},
		editContent: func(content []byte) ([]byte, error) {
			f.editedInput = string(content)
			if f.editResult == nil {
				return nil, errors.New("file was not modified")
			}
			return []byte(f.editResult(string(content))), nil
		},
	}
}

func TestBuildNewSnippet(t *testing.T) {
	tests := []struct {
		name         string
		opts         newOptions
		io           fakeNewIO
		wantTitle    string
		wantBody     string
		wantTags     []string
		wantLanguage string
		wantErr      string
	}{
		{
			name:         "piped stdin with all flags is non-interactive",
			opts:         newOptions{title: "List ports", tags: "net,mac", language: "bash"},
			io:           fakeNewIO{stdin: "lsof -i :8080\n", stdinPiped: true, noTTY: true},
			wantTitle:    "List ports",
			wantBody:     "lsof -i :8080",
			wantTags:     []string{"net", "mac"},
			wantLanguage: "bash",
		},
		{
			name:      "piped stdin preserves inner and trailing blank lines except one newline",
			opts:      newOptions{title: "Script"},
			io:        fakeNewIO{stdin: "  a\n\nb\n\n", stdinPiped: true, noTTY: true},
			wantTitle: "Script",
			wantBody:  "  a\n\nb\n",
			wantTags:  []string{},
		},
		{
			name:         "clipboard with title",
			opts:         newOptions{title: "List ports", clipboard: true, lang: "zsh"},
			io:           fakeNewIO{clipboard: "lsof -i :8080", noTTY: true},
			wantTitle:    "List ports",
			wantBody:     "lsof -i :8080",
			wantTags:     []string{},
			wantLanguage: "zsh",
		},
		{
			name:         "language wins over lang alias",
			opts:         newOptions{title: "t", clipboard: true, language: "bash", lang: "zsh"},
			io:           fakeNewIO{clipboard: "x"},
			wantTitle:    "t",
			wantBody:     "x",
			wantTags:     []string{},
			wantLanguage: "bash",
		},
		{
			name:    "empty piped stdin",
			opts:    newOptions{title: "t"},
			io:      fakeNewIO{stdin: "\n", stdinPiped: true},
			wantErr: "empty",
		},
		{
			name:    "whitespace-only clipboard",
			opts:    newOptions{title: "t", clipboard: true},
			io:      fakeNewIO{clipboard: " \n\t\n"},
			wantErr: "empty",
		},
		{
			name:    "clipboard read failure",
			opts:    newOptions{title: "t", clipboard: true},
			io:      fakeNewIO{clipErr: errors.New("boom")},
			wantErr: "clipboard",
		},
		{
			name:    "clipboard conflicts with piped stdin",
			opts:    newOptions{title: "t", clipboard: true},
			io:      fakeNewIO{stdin: "x", stdinPiped: true, clipboard: "y"},
			wantErr: "--clipboard",
		},
		{
			name:      "piped stdin without title prompts on tty",
			io:        fakeNewIO{stdin: "ls\n", stdinPiped: true, tty: "  From TTY  \n"},
			wantTitle: "From TTY",
			wantBody:  "ls",
			wantTags:  []string{},
		},
		{
			name:    "piped stdin without title and without tty",
			io:      fakeNewIO{stdin: "ls\n", stdinPiped: true, noTTY: true},
			wantErr: "--title",
		},
		{
			name:    "empty title from tty",
			io:      fakeNewIO{stdin: "ls\n", stdinPiped: true, tty: "\n"},
			wantErr: "cannot be empty",
		},
		{
			name:      "interactive prompts for title and command",
			io:        fakeNewIO{lines: []string{" Title ", " echo hi "}},
			wantTitle: "Title",
			wantBody:  "echo hi",
			wantTags:  []string{},
		},
		{
			name:      "title flag skips title prompt in interactive mode",
			opts:      newOptions{title: "Given", tags: "a"},
			io:        fakeNewIO{lines: []string{"echo hi"}},
			wantTitle: "Given",
			wantBody:  "echo hi",
			wantTags:  []string{"a"},
		},
		{
			name: "editor prefilled from stdin and frontmatter flags",
			opts: newOptions{title: "List ports", tags: "net", language: "bash", editor: true},
			io: fakeNewIO{
				stdin:      "lsof -i :8080\n",
				stdinPiped: true,
				editResult: func(content string) string {
					return strings.Replace(content, "lsof -i :8080", "lsof -nP -i :8080", 1) + "\n"
				},
			},
			wantTitle:    "List ports",
			wantBody:     "lsof -nP -i :8080",
			wantTags:     []string{"net"},
			wantLanguage: "bash",
		},
		{
			name:    "editor cancelled",
			opts:    newOptions{title: "t", editor: true, clipboard: true},
			io:      fakeNewIO{clipboard: "x"},
			wantErr: "cancelled",
		},
		{
			name: "editor result with empty body",
			opts: newOptions{title: "t", editor: true, clipboard: true},
			io: fakeNewIO{
				clipboard: "x",
				editResult: func(content string) string {
					return strings.TrimSuffix(content, "x") + "\n"
				},
			},
			wantErr: "empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := tt.io
			snippet, err := buildNewSnippet(tt.opts, fake.build())
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("buildNewSnippet() error = %v, want containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("buildNewSnippet() unexpected error: %v", err)
			}
			if snippet.ID == "" {
				t.Error("snippet ID is empty")
			}
			if snippet.Title != tt.wantTitle {
				t.Errorf("Title = %q, want %q", snippet.Title, tt.wantTitle)
			}
			if snippet.Body != tt.wantBody {
				t.Errorf("Body = %q, want %q", snippet.Body, tt.wantBody)
			}
			if !reflect.DeepEqual(snippet.Tags, tt.wantTags) {
				t.Errorf("Tags = %#v, want %#v", snippet.Tags, tt.wantTags)
			}
			if snippet.Language != tt.wantLanguage {
				t.Errorf("Language = %q, want %q", snippet.Language, tt.wantLanguage)
			}
		})
	}
}

func TestBuildNewSnippet_EditorShowsFrontmatterAndBody(t *testing.T) {
	fake := fakeNewIO{
		clipboard:  "kubectl get pods",
		editResult: func(content string) string { return content + "\n" },
	}
	opts := newOptions{title: "Pods", tags: "k8s", language: "bash", editor: true, clipboard: true}

	if _, err := buildNewSnippet(opts, fake.build()); err != nil {
		t.Fatalf("buildNewSnippet() unexpected error: %v", err)
	}

	for _, want := range []string{"title: Pods", "- k8s", "language: bash", "kubectl get pods"} {
		if !strings.Contains(fake.editedInput, want) {
			t.Errorf("editor content missing %q:\n%s", want, fake.editedInput)
		}
	}
}
