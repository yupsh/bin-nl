package main

import (
	"context"
	"fmt"
	"io"

	command "github.com/gloo-foo/cmd-nl"
	gloo "github.com/gloo-foo/framework"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v3"
)

const name = "nl"

const (
	flagBody      = "body-numbering"
	flagSep       = "number-separator"
	flagStart     = "starting-line-number"
	flagIncrement = "line-increment"
	flagWidth     = "number-width"
	flagFormat    = "number-format"
)

// usageText is the command's multi-line usage synopsis, shown in --help.
// cli/v3 indents the whole block by 3 spaces, so these lines are flush-left to
// stay aligned in the rendered output.
const usageText = `nl [OPTIONS] [FILE...]

Write each FILE to standard output, with line numbers added.
With no FILE, or when FILE is -, read standard input.`

// Error is the package sentinel error type. Every error this package emits is
// declared as a const of this type so callers can match with errors.Is.
type Error string

func (e Error) Error() string { return string(e) }

const (
	// ErrInvalidBody is returned for an unrecognized -b style.
	ErrInvalidBody Error = "invalid body-numbering style (want a, t, or n)"
	// ErrInvalidFormat is returned for an unrecognized -n format.
	ErrInvalidFormat Error = "invalid number-format (want ln, rn, or rz)"
)

// init replaces urfave/cli's default --version/-v flag with a --version-only
// flag, freeing the single-letter -v for command flags (e.g. nl -v) while
// still exposing the injected build version.
func init() {
	cli.VersionFlag = &cli.BoolFlag{Name: "version", Usage: "print version information and exit"}
}

// bodyOptions maps each supported -b style to its library option constant.
var bodyOptions = map[string]any{
	"a": command.NlBodyAll,
	"t": command.NlBodyNonEmpty,
	"n": command.NlBodyNone,
}

// formatValues is the set of supported -n format strings.
var formatValues = map[string]struct{}{
	"ln": {}, "rn": {}, "rz": {},
}

// run builds and executes the nl CLI against the injected version, I/O, and
// filesystem, returning the process exit code.
func run(version string, args []string, stdin io.Reader, stdout, stderr io.Writer, fs afero.Fs) int {
	cmd := newApp(version, stdin, stdout, fs)
	cmd.Writer = stdout
	cmd.ErrWriter = stderr
	if err := cmd.Run(context.Background(), args); err != nil {
		_, _ = fmt.Fprintf(stderr, name+": %v\n", err)
		return 1
	}
	return 0
}

func newApp(version string, stdin io.Reader, stdout io.Writer, fs afero.Fs) *cli.Command {
	return &cli.Command{
		Name:            name,
		Version:         version,
		Usage:           "number lines of files",
		UsageText:       usageText,
		HideHelpCommand: true,
		// Keep exit handling in run() rather than letting urfave/cli call
		// os.Exit, so the exit code stays testable.
		ExitErrHandler: func(context.Context, *cli.Command, error) {},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    flagBody,
				Aliases: []string{"b"},
				Usage:   "use STYLE for numbering body lines (a=all, t=non-empty, n=none)",
			},
			&cli.StringFlag{Name: flagSep, Aliases: []string{"s"}, Usage: "add STRING after (possible) line number"},
			&cli.IntFlag{Name: flagStart, Aliases: []string{"v"}, Usage: "first line number for each section"},
			&cli.IntFlag{Name: flagIncrement, Aliases: []string{"i"}, Usage: "line number increment at each line"},
			&cli.IntFlag{Name: flagWidth, Aliases: []string{"w"}, Usage: "use NUMBER columns for line numbers"},
			&cli.StringFlag{
				Name:    flagFormat,
				Aliases: []string{"n"},
				Usage:   "insert line numbers according to FORMAT (ln, rn, rz)",
			},
		},
		Action: action(stdin, stdout, fs),
	}
}

func action(stdin io.Reader, stdout io.Writer, fs afero.Fs) cli.ActionFunc {
	return func(_ context.Context, c *cli.Command) error {
		opts, err := options(c)
		if err != nil {
			return err
		}
		_, err = gloo.Run(source(c, stdin, fs), gloo.ByteWriteTo(stdout), command.Nl(opts...))
		return err
	}
}

func source(c *cli.Command, stdin io.Reader, fs afero.Fs) any {
	if c.NArg() == 0 {
		return gloo.ByteReaderSource([]io.Reader{stdin})
	}
	files := make([]gloo.File, c.NArg())
	for i := range files {
		files[i] = gloo.File(c.Args().Get(i))
	}
	return gloo.ByteFileSource(fs, files)
}

func options(c *cli.Command) ([]any, error) {
	opts, err := valueOptions(c)
	if err != nil {
		return nil, err
	}
	opts = append(opts, scalarOptions(c)...)
	return opts, nil
}

// valueOptions collects the flags whose values must be validated against a
// fixed set (-b style, -n format) before mapping to library options.
func valueOptions(c *cli.Command) ([]any, error) {
	var opts []any
	if c.IsSet(flagBody) {
		opt, ok := bodyOptions[c.String(flagBody)]
		if !ok {
			return nil, ErrInvalidBody
		}
		opts = append(opts, opt)
	}
	if c.IsSet(flagFormat) {
		format := c.String(flagFormat)
		if _, ok := formatValues[format]; !ok {
			return nil, ErrInvalidFormat
		}
		opts = append(opts, command.NlFormat(format))
	}
	return opts, nil
}

// scalarOptions collects the flags that map directly to a library option with
// no validation (-s, -v, -i, -w).
func scalarOptions(c *cli.Command) []any {
	var opts []any
	if c.IsSet(flagSep) {
		opts = append(opts, command.NlSep(c.String(flagSep)))
	}
	if c.IsSet(flagStart) {
		opts = append(opts, command.NlStart(c.Int(flagStart)))
	}
	if c.IsSet(flagIncrement) {
		opts = append(opts, command.NlIncrement(c.Int(flagIncrement)))
	}
	if c.IsSet(flagWidth) {
		opts = append(opts, command.NlWidth(c.Int(flagWidth)))
	}
	return opts
}
