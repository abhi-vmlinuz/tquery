package cmd

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/tquery/tquery/pkg/engine"
	"github.com/tquery/tquery/pkg/parser"
	"github.com/tquery/tquery/pkg/render"
	"github.com/tquery/tquery/pkg/tui"
)

const Version = "1.0.0"

type Config struct {
	Format      string
	Interactive bool
	NoHeaders   bool
	NoUnwrap    bool
	NoColor     bool
	ShowVersion bool
	Limit       int
	Query       string
	FilePath    string
}

func Execute() {
	cfg := parseFlags()

	if cfg.ShowVersion {
		fmt.Printf("tquery version %s\n", Version)
		os.Exit(0)
	}

	rawJSON, err := readInput(cfg.FilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	isTTYStdin := isatty.IsTerminal(os.Stdin.Fd())
	isTTYStdout := isatty.IsTerminal(os.Stdout.Fd())

	runInteractive := cfg.Interactive
	if !cfg.Interactive && isTTYStdin && isTTYStdout && cfg.FilePath == "" {
		// Default CLI behavior
	}

	if runInteractive {
		if err := tui.Run(rawJSON, cfg.Query, !cfg.NoUnwrap); err != nil {
			fmt.Fprintf(os.Stderr, "Interactive TUI error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Non-Interactive Pipe Output
	var targetData any
	if cfg.Query != "" && cfg.Query != "." {
		rawObj := parseRawInterface(rawJSON)
		res, err := engine.Evaluate(cfg.Query, rawObj)
		if err != nil {
			fmt.Fprintf(os.Stderr, "JQ evaluation error: %v\n", err)
			os.Exit(1)
		}
		targetData = res
	} else {
		targetData = parseRawInterface(rawJSON)
	}

	// Serialize target back to JSON bytes for parser
	jsonBytes, err := marshalInterface(targetData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Marshal error: %v\n", err)
		os.Exit(1)
	}

	ds, err := parser.Parse(jsonBytes, !cfg.NoUnwrap && cfg.Query == "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
	}

	// Apply -<number> / --limit trimming
	if cfg.Limit > 0 {
		if len(ds.Rows) > cfg.Limit {
			ds.Rows = ds.Rows[:cfg.Limit]
		}
		if s, ok := ds.Unwrapped.([]any); ok && len(s) > cfg.Limit {
			ds.Unwrapped = s[:cfg.Limit]
		}
	}

	opts := render.RenderOptions{
		Format:     render.Format(strings.ToLower(cfg.Format)),
		ShowHeader: !cfg.NoHeaders,
		UseColor:   !cfg.NoColor && isTTYStdout,
	}

	if err := render.Render(os.Stdout, ds, opts); err != nil {
		fmt.Fprintf(os.Stderr, "Render error: %v\n", err)
		os.Exit(1)
	}
}

func parseFlags() Config {
	var cfg Config
	var numericLimit int

	// Pre-process arguments to detect -<number> like -10, -50, etc.
	var filteredArgs []string
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			// Check if all remaining characters are digits
			if num, err := strconv.Atoi(arg[1:]); err == nil && num > 0 {
				numericLimit = num
				continue
			}
		}
		filteredArgs = append(filteredArgs, arg)
	}

	fs := flag.NewFlagSet("tq", flag.ExitOnError)

	var limitFlag int
	fs.StringVar(&cfg.Format, "f", "table", "Output format: table, markdown, csv, tsv, tree, json")
	fs.StringVar(&cfg.Format, "format", "table", "Output format: table, markdown, csv, tsv, tree, json")
	fs.BoolVar(&cfg.Interactive, "i", false, "Launch interactive TUI mode")
	fs.BoolVar(&cfg.Interactive, "interactive", false, "Launch interactive TUI mode")
	fs.BoolVar(&cfg.NoHeaders, "n", false, "Hide headers")
	fs.BoolVar(&cfg.NoHeaders, "no-headers", false, "Hide headers")
	fs.BoolVar(&cfg.NoUnwrap, "no-unwrap", false, "Disable root array wrapper auto-unwrapping")
	fs.BoolVar(&cfg.NoColor, "no-color", false, "Disable ANSI color formatting")
	fs.IntVar(&limitFlag, "l", 0, "Limit number of output rows (e.g. -l 10 or -10)")
	fs.IntVar(&limitFlag, "L", 0, "Limit number of output rows")
	fs.IntVar(&limitFlag, "limit", 0, "Limit number of output rows")
	fs.BoolVar(&cfg.ShowVersion, "v", false, "Show version")
	fs.BoolVar(&cfg.ShowVersion, "version", false, "Show version")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: tq [options] [-<number>] [jq_query] [file]\n\n")
		fmt.Fprintf(os.Stderr, "tq (tquery) converts raw JSON & JQ streams into human-readable tables, trees, and interactive UI.\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  curl https://api.example.com/models | tq\n")
		fmt.Fprintf(os.Stderr, "  curl https://api.example.com/models | tq -10\n")
		fmt.Fprintf(os.Stderr, "  curl https://api.example.com/models | tq '.data[] | {id, owned_by}'\n")
		fmt.Fprintf(os.Stderr, "  tq -5 -f markdown '.items' data.json\n")
		fmt.Fprintf(os.Stderr, "  tq -i data.json\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
	}

	_ = fs.Parse(filteredArgs)

	if limitFlag > 0 {
		cfg.Limit = limitFlag
	} else {
		cfg.Limit = numericLimit
	}

	args := fs.Args()
	if len(args) > 0 {
		if isFile(args[0]) {
			cfg.FilePath = args[0]
		} else {
			cfg.Query = args[0]
			if len(args) > 1 {
				cfg.FilePath = args[1]
			}
		}
	}

	return cfg
}

func readInput(filePath string) ([]byte, error) {
	if filePath != "" {
		return os.ReadFile(filePath)
	}

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return nil, fmt.Errorf("no input provided. Pipe JSON to stdin or specify a JSON file path")
	}

	return io.ReadAll(os.Stdin)
}

func parseRawInterface(b []byte) any {
	ds, err := parser.Parse(b, false)
	if err == nil && ds != nil {
		return ds.Raw
	}
	return nil
}

func marshalInterface(v any) ([]byte, error) {
	b, ok := v.([]byte)
	if ok {
		return b, nil
	}
	return parser.MarshalAny(v)
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
