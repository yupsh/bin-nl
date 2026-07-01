package main

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/spf13/afero"
)

func TestRun(t *testing.T) {
	cases := []struct {
		files      map[string]string
		name       string
		version    string
		stdin      string
		wantOut    string
		wantErrSub string
		args       []string
		wantCode   int
	}{
		{
			name:    "default numbering",
			args:    []string{"nl"},
			stdin:   "alpha\nbeta\n",
			wantOut: "     1\talpha\n     2\tbeta\n",
		},
		{
			name:    "body all",
			args:    []string{"nl", "-b", "a"},
			stdin:   "alpha\n\nbeta\n",
			wantOut: "     1\talpha\n     2\t\n     3\tbeta\n",
		},
		{
			name:    "body non-empty",
			args:    []string{"nl", "-b", "t"},
			stdin:   "alpha\n\nbeta\n",
			wantOut: "     1\talpha\n\n     2\tbeta\n",
		},
		{
			name:    "body none",
			args:    []string{"nl", "-b", "n"},
			stdin:   "alpha\nbeta\n",
			wantOut: "      \talpha\n      \tbeta\n",
		},
		{
			name:       "body invalid",
			args:       []string{"nl", "-b", "z"},
			stdin:      "alpha\n",
			wantCode:   1,
			wantErrSub: "nl: invalid body-numbering style",
		},
		{
			name:    "separator",
			args:    []string{"nl", "-s", ": "},
			stdin:   "alpha\nbeta\n",
			wantOut: "     1: alpha\n     2: beta\n",
		},
		{
			name:    "starting line number",
			args:    []string{"nl", "-v", "10"},
			stdin:   "alpha\nbeta\n",
			wantOut: "    10\talpha\n    11\tbeta\n",
		},
		{
			name:    "line increment",
			args:    []string{"nl", "-i", "5"},
			stdin:   "alpha\nbeta\n",
			wantOut: "     1\talpha\n     6\tbeta\n",
		},
		{
			name:    "number width",
			args:    []string{"nl", "-w", "3"},
			stdin:   "alpha\nbeta\n",
			wantOut: "  1\talpha\n  2\tbeta\n",
		},
		{
			name:    "format left justified",
			args:    []string{"nl", "-n", "ln"},
			stdin:   "alpha\n",
			wantOut: "1     \talpha\n",
		},
		{
			name:    "format right zero padded",
			args:    []string{"nl", "-n", "rz"},
			stdin:   "alpha\n",
			wantOut: "000001\talpha\n",
		},
		{
			name:       "format invalid",
			args:       []string{"nl", "-n", "xx"},
			stdin:      "alpha\n",
			wantCode:   1,
			wantErrSub: "nl: invalid number-format",
		},
		{
			name:    "file source",
			args:    []string{"nl", "/in.txt"},
			files:   map[string]string{"/in.txt": "one\ntwo\n"},
			wantOut: "     1\tone\n     2\ttwo\n",
		},
		{
			name:    "version flag reports injected version",
			version: "1.2.3",
			args:    []string{"nl", "--version"},
			wantOut: "nl version 1.2.3\n",
		},
		{
			name:       "unknown flag errors",
			args:       []string{"nl", "--nope"},
			wantCode:   1,
			wantErrSub: "nl:",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			for path, content := range tc.files {
				if err := afero.WriteFile(fs, path, []byte(content), 0o644); err != nil {
					t.Fatalf("write fixture %s: %v", path, err)
				}
			}

			var out, errOut bytes.Buffer
			code := run(tc.version, tc.args, strings.NewReader(tc.stdin), &out, &errOut, fs)

			if code != tc.wantCode {
				t.Fatalf("exit code = %d, want %d (stderr=%q)", code, tc.wantCode, errOut.String())
			}
			if tc.wantErrSub == "" && out.String() != tc.wantOut {
				t.Fatalf("stdout = %q, want %q", out.String(), tc.wantOut)
			}
			if tc.wantErrSub != "" && !strings.Contains(errOut.String(), tc.wantErrSub) {
				t.Fatalf("stderr = %q, want substring %q", errOut.String(), tc.wantErrSub)
			}
		})
	}
}

func Test_main(t *testing.T) {
	origExit, origRun := osExit, runCLI
	t.Cleanup(func() { osExit, runCLI = origExit, origRun })

	gotCode := -1
	osExit = func(code int) { gotCode = code }
	runCLI = func(string, []string, io.Reader, io.Writer, io.Writer, afero.Fs) int { return 7 }

	main()

	if gotCode != 7 {
		t.Fatalf("main propagated exit code %d, want 7", gotCode)
	}
}
