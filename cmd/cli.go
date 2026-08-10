package cmd

import (
	"flag"
	"fmt"
	"io"
	"os"
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

	// Auto-detect interactive TUI mode if stdin is terminal and no explicit pipe
	isTTYStdin := isatty.IsTerminal(os.Stdin.Fd())
	isTTYStdout := isatty.IsTerminal(os.Stdout.Fd())

	runInteractive := cfg.Interactive
	if !cfg.Interactive && isTTYStdin && isTTYStdout && cfg.FilePath == "" {
		// If user runs tquery directly without piping input, prompt help or default interactive if input available
	}

	if runInteractive {
		if err := tui.Run(rawJSON, cfg.Query, !cfg.NoUnwrap); err != nil {
			fmt.Fprintf(os.Stderr, "Interactive TUI error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Non-Interactive Pipe Output
	targetData := any(nil)
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

	flag.StringVar(&cfg.Format, "f", "table", "Output format: table, markdown, csv, tsv, tree, json")
	flag.StringVar(&cfg.Format, "format", "table", "Output format: table, markdown, csv, tsv, tree, json")
	flag.BoolVar(&cfg.Interactive, "i", false, "Launch interactive TUI mode")
	flag.BoolVar(&cfg.Interactive, "interactive", false, "Launch interactive TUI mode")
	flag.BoolVar(&cfg.NoHeaders, "n", false, "Hide headers")
	flag.BoolVar(&cfg.NoHeaders, "no-headers", false, "Hide headers")
	flag.BoolVar(&cfg.NoUnwrap, "no-unwrap", false, "Disable root array wrapper auto-unwrapping")
	flag.BoolVar(&cfg.NoColor, "no-color", false, "Disable ANSI color formatting")
	flag.BoolVar(&cfg.ShowVersion, "v", false, "Show version")
	flag.BoolVar(&cfg.ShowVersion, "version", false, "Show version")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: tquery [options] [jq_query] [file]\n\n")
		fmt.Fprintf(os.Stderr, "tquery converts raw JSON & JQ streams into human-readable tables, trees, and interactive UI.\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  curl https://api.example.com/models | tquery\n")
		fmt.Fprintf(os.Stderr, "  curl https://api.example.com/models | tquery '.data[] | {id, owned_by}'\n")
		fmt.Fprintf(os.Stderr, "  tquery -f markdown '.items' data.json\n")
		fmt.Fprintf(os.Stderr, "  tquery -i data.json\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	args := flag.Args()
	if len(args) > 0 {
		// First non-flag arg can be jq query or file path if it starts with file
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

