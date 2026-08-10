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
	"github.com/tquery/tquery/pkg/filter"
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
	Grep        string
	InvertMatch bool
	IgnoreCase  bool
	Query       string
	FilePath    string
}

func Execute() {
	cfg := parseFlags()

	if cfg.ShowVersion {
		fmt.Printf("tq version %s\n", Version)
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

	// Apply -g / -e / --grep regex & literal filtering
	if cfg.Grep != "" {
		filteredDS, err := filter.Filter(ds, filter.Options{
			Pattern:     cfg.Grep,
			InvertMatch: cfg.InvertMatch,
			IgnoreCase:  cfg.IgnoreCase,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Grep error: %v\n", err)
			os.Exit(1)
		}
		ds = filteredDS
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
		Format:              render.Format(strings.ToLower(cfg.Format)),
		ShowHeader:          !cfg.NoHeaders,
		UseColor:            !cfg.NoColor,
		HighlightPattern:    cfg.Grep,
		HighlightIgnoreCase: cfg.IgnoreCase,
	}

	if err := render.Render(os.Stdout, ds, opts); err != nil {
		fmt.Fprintf(os.Stderr, "Render error: %v\n", err)
		os.Exit(1)
	}
}

func parseFlags() Config {
	var cfg Config
	var numericLimit int
	cfg.IgnoreCase = true // Smart case-insensitive by default for seamless grep experience

	rawArgs := os.Args[1:]
	var filteredArgs []string

	for i := 0; i < len(rawArgs); i++ {
		arg := rawArgs[i]

		// 1. Detect -<number> (e.g. -10, -5)
		if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			if num, err := strconv.Atoi(arg[1:]); err == nil && num > 0 {
				numericLimit = num
				continue
			}
		}

		// 2. Detect combined grep flags like -gi, -gv, -gvi, -giv
		if strings.HasPrefix(arg, "-g") && len(arg) > 2 {
			subFlags := arg[2:]
			hasI := strings.Contains(subFlags, "i")
			hasV := strings.Contains(subFlags, "v")

			if hasI {
				cfg.IgnoreCase = true
			}
			if hasV {
				cfg.InvertMatch = true
			}

			// If next arg exists and doesn't start with '-', it's the pattern
			if i+1 < len(rawArgs) && !strings.HasPrefix(rawArgs[i+1], "-") {
				cfg.Grep = rawArgs[i+1]
				i++
				continue
			}
			continue
		}

		// 3. Detect -e <pattern> (standard grep flag)
		if arg == "-e" && i+1 < len(rawArgs) {
			cfg.Grep = rawArgs[i+1]
			i++
			continue
		}

		filteredArgs = append(filteredArgs, arg)
	}

	fs := flag.NewFlagSet("tq", flag.ExitOnError)

	var limitFlag int
	var grepFlag string
	var invertFlag bool
	var ignoreCaseFlag bool
	var interactiveFlag bool

	fs.StringVar(&cfg.Format, "f", "table", "Output format: table, markdown, csv, tsv, tree, json")
	fs.StringVar(&cfg.Format, "format", "table", "Output format: table, markdown, csv, tsv, tree, json")
	fs.BoolVar(&interactiveFlag, "i", false, "Launch interactive TUI mode (or ignore-case when using -g)")
	fs.BoolVar(&interactiveFlag, "interactive", false, "Launch interactive TUI mode")
	fs.BoolVar(&interactiveFlag, "ui", false, "Launch interactive TUI mode")
	fs.BoolVar(&cfg.NoHeaders, "n", false, "Hide headers")
	fs.BoolVar(&cfg.NoHeaders, "no-headers", false, "Hide headers")
	fs.BoolVar(&cfg.NoUnwrap, "no-unwrap", false, "Disable root array wrapper auto-unwrapping")
	fs.BoolVar(&cfg.NoColor, "no-color", false, "Disable ANSI color formatting")
	fs.IntVar(&limitFlag, "l", 0, "Limit number of output rows (e.g. -l 10 or -10)")
	fs.IntVar(&limitFlag, "L", 0, "Limit number of output rows")
	fs.IntVar(&limitFlag, "limit", 0, "Limit number of output rows")
	fs.StringVar(&grepFlag, "g", "", "Filter rows matching regex or string pattern")
	fs.StringVar(&grepFlag, "grep", "", "Filter rows matching regex or string pattern")
	fs.BoolVar(&invertFlag, "v", false, "Invert match (select non-matching rows)")
	fs.BoolVar(&invertFlag, "V", false, "Invert match (select non-matching rows)")
	fs.BoolVar(&invertFlag, "invert", false, "Invert match (select non-matching rows)")
	fs.BoolVar(&invertFlag, "invert-match", false, "Invert match (select non-matching rows)")
	fs.BoolVar(&ignoreCaseFlag, "I", false, "Case-insensitive grep match")
	fs.BoolVar(&ignoreCaseFlag, "ignore-case", false, "Case-insensitive grep match")
	fs.BoolVar(&cfg.ShowVersion, "version", false, "Show version")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: tq [options] [-<number>] [-g <pattern>] [jq_query] [file]\n\n")
		fmt.Fprintf(os.Stderr, "tq (tquery) converts raw JSON & JQ streams into human-readable tables, trees, and interactive UI.\n\n")
		fmt.Fprintf(os.Stderr, "Grep Options:\n")
		fmt.Fprintf(os.Stderr, "  -g, --grep <pattern>    Filter rows by regex or string pattern\n")
		fmt.Fprintf(os.Stderr, "  -e <pattern>            Alias for -g pattern\n")
		fmt.Fprintf(os.Stderr, "  -gi, -gv, -gvi          Combined grep flags (ignore case, invert match)\n")
		fmt.Fprintf(os.Stderr, "  -v, -V, --invert        Invert grep match (like grep -v)\n")
		fmt.Fprintf(os.Stderr, "  -<number>, -l <number>  Limit output rows (e.g. -10 or -l 5)\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  curl https://integrate.api.nvidia.com/v1/models | tq\n")
		fmt.Fprintf(os.Stderr, "  curl https://integrate.api.nvidia.com/v1/models | tq -10\n")
		fmt.Fprintf(os.Stderr, "  curl https://integrate.api.nvidia.com/v1/models | tq -g 'google|deepseek'\n")
		fmt.Fprintf(os.Stderr, "  curl https://integrate.api.nvidia.com/v1/models | tq -gi 'llama' -5\n")
		fmt.Fprintf(os.Stderr, "  curl https://integrate.api.nvidia.com/v1/models | tq -gv 'google'\n")
		fmt.Fprintf(os.Stderr, "  tq -i data.json\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
	}

	_ = fs.Parse(filteredArgs)

	if grepFlag != "" {
		cfg.Grep = grepFlag
	}
	if invertFlag {
		cfg.InvertMatch = true
	}
	if ignoreCaseFlag {
		cfg.IgnoreCase = true
	}

	// Contextual -i handling:
	// If -i is passed and a grep filter is active, -i acts as ignore-case;
	// if -i is passed without grep, it acts as interactive TUI mode!
	if interactiveFlag {
		if cfg.Grep != "" {
			cfg.IgnoreCase = true
		} else {
			cfg.Interactive = true
		}
	}

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
