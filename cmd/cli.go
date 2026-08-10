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

const Version = "0.1.0"

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

	// 1. Evaluate JQ Query if provided
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
		if !cfg.NoUnwrap {
			targetData = parser.UnwrapRoot(targetData)
		}
	}

	// 2. Smart Shape Auto-Detection
	chosenFormat := strings.ToLower(cfg.Format)
	if chosenFormat == "auto" || chosenFormat == "" {
		if parser.IsComplexStructure(targetData) {
			chosenFormat = "tree"
		} else {
			chosenFormat = "table"
		}
	}

	// 3. Universal Grep across Tree, JSON, and Tabular Formats
	if cfg.Grep != "" && (chosenFormat == "tree" || chosenFormat == "json") {
		pruned, matched := filter.PruneJSON(targetData, cfg.Grep, cfg.IgnoreCase, cfg.InvertMatch)
		if !matched {
			if chosenFormat == "tree" {
				fmt.Fprintln(os.Stdout, "[] (no matching branches)")
			} else {
				fmt.Fprintln(os.Stdout, "{}")
			}
			return
		}
		targetData = pruned
	}

	// Serialize target back to JSON bytes for parser
	jsonBytes, err := marshalInterface(targetData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Marshal error: %v\n", err)
		os.Exit(1)
	}

	ds, err := parser.Parse(jsonBytes, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
	}

	// 4. Tabular Grep Filter (for Table, Markdown, CSV)
	if cfg.Grep != "" && chosenFormat != "tree" && chosenFormat != "json" {
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

	// 5. Apply -<number> / --limit trimming
	if cfg.Limit > 0 {
		if len(ds.Rows) > cfg.Limit {
			ds.Rows = ds.Rows[:cfg.Limit]
		}
		if s, ok := ds.Unwrapped.([]any); ok && len(s) > cfg.Limit {
			ds.Unwrapped = s[:cfg.Limit]
		}
	}

	opts := render.RenderOptions{
		Format:              render.Format(chosenFormat),
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
	cfg.IgnoreCase = true // Smart case-insensitive by default
	cfg.Format = "auto"   // Auto-detect table vs tree based on JSON structure

	rawArgs := os.Args[1:]
	var filteredArgs []string

	for i := 0; i < len(rawArgs); i++ {
		arg := rawArgs[i]

		// 1. Direct Shape Shortcuts (--tree, --table, --json, --raw, --markdown, --csv, --tsv)
		switch arg {
		case "--tree":
			cfg.Format = "tree"
			continue
		case "--table":
			cfg.Format = "table"
			continue
		case "--json", "--raw":
			cfg.Format = "json"
			continue
		case "--markdown", "--md":
			cfg.Format = "markdown"
			continue
		case "--csv":
			cfg.Format = "csv"
			continue
		case "--tsv":
			cfg.Format = "tsv"
			continue
		}

		// 2. Detect -<number> (e.g. -10, -5)
		if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			if num, err := strconv.Atoi(arg[1:]); err == nil && num > 0 {
				numericLimit = num
				continue
			}
		}

		// 3. Detect combined grep flags like -gi, -gv, -gvi, -giv
		if strings.HasPrefix(arg, "-g") && len(arg) > 2 {
			subFlags := arg[2:]
			if strings.Contains(subFlags, "i") {
				cfg.IgnoreCase = true
			}
			if strings.Contains(subFlags, "v") {
				cfg.InvertMatch = true
			}

			if i+1 < len(rawArgs) && !strings.HasPrefix(rawArgs[i+1], "-") {
				cfg.Grep = rawArgs[i+1]
				i++
				continue
			}
			continue
		}

		// 4. Detect -e <pattern>
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
	var formatFlag string

	fs.StringVar(&formatFlag, "f", "", "Output format: table, tree, markdown, csv, tsv, json")
	fs.StringVar(&formatFlag, "format", "", "Output format: table, tree, markdown, csv, tsv, json")
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
	fs.StringVar(&grepFlag, "g", "", "Filter rows or tree branches matching regex/pattern")
	fs.StringVar(&grepFlag, "grep", "", "Filter rows or tree branches matching regex/pattern")
	fs.BoolVar(&invertFlag, "v", false, "Invert match (select non-matching rows)")
	fs.BoolVar(&invertFlag, "V", false, "Invert match (select non-matching rows)")
	fs.BoolVar(&invertFlag, "invert", false, "Invert match (select non-matching rows)")
	fs.BoolVar(&invertFlag, "invert-match", false, "Invert match (select non-matching rows)")
	fs.BoolVar(&ignoreCaseFlag, "I", false, "Case-insensitive grep match")
	fs.BoolVar(&ignoreCaseFlag, "ignore-case", false, "Case-insensitive grep match")
	fs.BoolVar(&cfg.ShowVersion, "version", false, "Show version")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: tq [options] [--tree|--table|--json] [-<number>] [-g <pattern>] [jq_query] [file]\n\n")
		fmt.Fprintf(os.Stderr, "tq (tquery) converts raw JSON & JQ streams into human-readable tables, trees, and interactive UI.\n\n")
		fmt.Fprintf(os.Stderr, "Shape Shortcuts:\n")
		fmt.Fprintf(os.Stderr, "  --tree                  Force hierarchical tree view\n")
		fmt.Fprintf(os.Stderr, "  --table                 Force tabular view\n")
		fmt.Fprintf(os.Stderr, "  --json, --raw           Force formatted JSON output\n")
		fmt.Fprintf(os.Stderr, "  --markdown, --md        Force markdown table output\n")
		fmt.Fprintf(os.Stderr, "  --csv, --tsv            Force delimited data output\n\n")
		fmt.Fprintf(os.Stderr, "Grep Options:\n")
		fmt.Fprintf(os.Stderr, "  -g, --grep <pattern>    Filter rows/tree branches by regex or string\n")
		fmt.Fprintf(os.Stderr, "  -e <pattern>            Alias for -g pattern\n")
		fmt.Fprintf(os.Stderr, "  -gi, -gv, -gvi          Combined grep flags (ignore case, invert match)\n")
		fmt.Fprintf(os.Stderr, "  -v, -V, --invert        Invert grep match\n")
		fmt.Fprintf(os.Stderr, "  -<number>, -l <number>  Limit output rows (e.g. -10 or -l 5)\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  curl https://integrate.api.nvidia.com/v1/models | tq\n")
		fmt.Fprintf(os.Stderr, "  docker inspect container | tq -g 'port'\n")
		fmt.Fprintf(os.Stderr, "  kubectl get pod my-pod -o json | tq -g 'nginx'\n")
		fmt.Fprintf(os.Stderr, "  curl https://integrate.api.nvidia.com/v1/models | tq -g 'google' -5\n")
		fmt.Fprintf(os.Stderr, "  tq -i data.json\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
	}

	_ = fs.Parse(filteredArgs)

	if formatFlag != "" {
		cfg.Format = formatFlag
	}
	if grepFlag != "" {
		cfg.Grep = grepFlag
	}
	if invertFlag {
		cfg.InvertMatch = true
	}
	if ignoreCaseFlag {
		cfg.IgnoreCase = true
	}

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
