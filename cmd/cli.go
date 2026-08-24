package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/tquery/tquery/pkg/engine"
	"github.com/tquery/tquery/pkg/filter"
	"github.com/tquery/tquery/pkg/pager"
	"github.com/tquery/tquery/pkg/parser"
	"github.com/tquery/tquery/pkg/render"
	"github.com/tquery/tquery/pkg/tui"
)

const Version = "0.1.3"

type Config struct {
	Format      string
	Interactive bool
	NoHeaders   bool
	NoUnwrap    bool
	NoColor     bool
	NoPager     bool
	ShowVersion bool
	Limit       int
	Patterns    []string
	Strict      bool
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
	if !cfg.Interactive && isTTYStdin && isTTYStdout && cfg.FilePath == "" && len(cfg.Patterns) == 0 {
		// Default CLI behavior when launched interactively with no args
	}

	if runInteractive {
		if err := tui.Run(rawJSON, cfg.Query, !cfg.NoUnwrap); err != nil {
			fmt.Fprintf(os.Stderr, "Interactive TUI error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// 1. Parse JSON / NDJSON input stream
	var targetData any
	rawObj, err := parser.DecodeStream(rawJSON)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if cfg.Query != "" && cfg.Query != "." {
		res, err := engine.Evaluate(cfg.Query, rawObj)
		if err != nil {
			fmt.Fprintf(os.Stderr, "JQ evaluation error: %v\n", err)
			os.Exit(1)
		}
		targetData = res
	} else {
		targetData = rawObj
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

	// 3. Universal Multi-Pattern Search & Path Pruning
	if len(cfg.Patterns) > 0 {
		pruned, matched := filter.PruneJSON(targetData, filter.MultiOptions{
			Patterns:    cfg.Patterns,
			InvertMatch: cfg.InvertMatch,
			IgnoreCase:  cfg.IgnoreCase,
			Strict:      cfg.Strict,
		})
		if !matched {
			if chosenFormat == "tree" {
				fmt.Fprintln(os.Stdout, "[] (no matching branches)")
			} else if chosenFormat == "json" {
				fmt.Fprintln(os.Stdout, "{}")
			} else {
				fmt.Fprintln(os.Stdout, "[] (no matching records)")
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

	// 4. Apply -<number> / --limit trimming
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
		Limit:               cfg.Limit,
		HighlightPatterns:   cfg.Patterns,
		HighlightIgnoreCase: cfg.IgnoreCase,
	}

	var buf bytes.Buffer
	if err := render.Render(&buf, ds, opts); err != nil {
		fmt.Fprintf(os.Stderr, "Render error: %v\n", err)
		os.Exit(1)
	}

	if err := pager.WriteOrPage(os.Stdout, buf.String(), cfg.NoPager); err != nil {
		fmt.Fprintf(os.Stderr, "Output error: %v\n", err)
		os.Exit(1)
	}
}

func parseFlags() Config {
	var cfg Config
	cfg.IgnoreCase = true // Smart case-insensitive by default
	cfg.Format = "auto"   // Auto-detect table vs tree based on JSON structure

	rawArgs := os.Args[1:]
	var positionals []string

	for i := 0; i < len(rawArgs); i++ {
		arg := rawArgs[i]

		if arg == "-h" || arg == "--help" {
			printUsageAndExit()
		}
		if arg == "--version" {
			cfg.ShowVersion = true
			continue
		}

		// Shape Shortcuts
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
		case "--strict":
			cfg.Strict = true
			continue
		case "--no-pager":
			cfg.NoPager = true
			continue
		case "--pager":
			cfg.NoPager = false
			continue
		case "-n", "--no-headers":
			cfg.NoHeaders = true
			continue
		case "--no-unwrap":
			cfg.NoUnwrap = true
			continue
		case "--no-color":
			cfg.NoColor = true
			continue
		case "-u", "--ui", "--interactive":
			cfg.Interactive = true
			continue
		}

		// -f / --format <val>
		if arg == "-f" || arg == "--format" {
			if i+1 < len(rawArgs) {
				cfg.Format = rawArgs[i+1]
				i++
				continue
			}
		}

		// -l / -L / --limit <val>
		if arg == "-l" || arg == "-L" || arg == "--limit" {
			if i+1 < len(rawArgs) {
				if num, err := strconv.Atoi(rawArgs[i+1]); err == nil && num > 0 {
					cfg.Limit = num
					i++
					continue
				}
			}
		}

		// -<number> shortcut (e.g. -10, -5)
		if strings.HasPrefix(arg, "-") && len(arg) > 1 && !strings.HasPrefix(arg, "--") {
			if num, err := strconv.Atoi(arg[1:]); err == nil && num > 0 {
				cfg.Limit = num
				continue
			}
		}

		// -g / -e / --grep <pattern>
		if arg == "-g" || arg == "-e" || arg == "--grep" {
			if i+1 < len(rawArgs) {
				cfg.Patterns = append(cfg.Patterns, rawArgs[i+1])
				i++
				continue
			}
		}

		// Combined -g flags like -gi, -gv, -gvi, -giv, -gI, -gV
		if strings.HasPrefix(arg, "-g") && len(arg) > 2 {
			subFlags := arg[2:]
			if strings.ContainsAny(subFlags, "iI") {
				cfg.IgnoreCase = true
			}
			if strings.ContainsAny(subFlags, "vV") {
				cfg.InvertMatch = true
			}

			if i+1 < len(rawArgs) && !strings.HasPrefix(rawArgs[i+1], "-") {
				cfg.Patterns = append(cfg.Patterns, rawArgs[i+1])
				i++
				continue
			}
			continue
		}

		// POSIX combined flags like -iv, -vi, -i, -v, -u, -I, -V
		if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") && len(arg) > 1 {
			flags := arg[1:]
			isFlagCluster := true
			for _, r := range flags {
				if !strings.ContainsRune("iIvVuU", r) {
					isFlagCluster = false
					break
				}
			}
			if isFlagCluster {
				if strings.ContainsAny(flags, "iI") {
					cfg.IgnoreCase = true
				}
				if strings.ContainsAny(flags, "vV") {
					cfg.InvertMatch = true
				}
				if strings.ContainsAny(flags, "uU") {
					cfg.Interactive = true
				}
				continue
			}
		}

		// Long flags with = (e.g. --format=tree, --limit=10, --grep=meta)
		if strings.HasPrefix(arg, "--") && strings.Contains(arg, "=") {
			parts := strings.SplitN(arg[2:], "=", 2)
			key, val := parts[0], parts[1]
			switch key {
			case "format", "f":
				cfg.Format = val
			case "limit", "l", "L":
				if num, err := strconv.Atoi(val); err == nil && num > 0 {
					cfg.Limit = num
				}
			case "grep", "g", "e":
				cfg.Patterns = append(cfg.Patterns, val)
			}
			continue
		}

		// If it's an unrecognized option starting with -, ignore or skip
		if strings.HasPrefix(arg, "-") {
			continue
		}

		// Positional argument (JQ query or File path)
		positionals = append(positionals, arg)
	}

	if len(positionals) > 0 {
		if isFile(positionals[0]) {
			cfg.FilePath = positionals[0]
		} else {
			cfg.Query = positionals[0]
			if len(positionals) > 1 {
				cfg.FilePath = positionals[1]
			}
		}
	}

	return cfg
}

func printUsageAndExit() {
	fmt.Fprintf(os.Stderr, "Usage: tq [options] [--tree|--table|--json] [-<number>] [-g <pat1> -g <pat2>] [--strict] [jq_query] [file]\n\n")
	fmt.Fprintf(os.Stderr, "tq (tquery) converts raw JSON & JQ streams into human-readable tables, trees, and interactive UI.\n\n")
	fmt.Fprintf(os.Stderr, "Multi-Pattern Search:\n")
	fmt.Fprintf(os.Stderr, "  -g, --grep <pattern>    Filter by pattern (can be specified multiple times for OR)\n")
	fmt.Fprintf(os.Stderr, "  -e <pattern>            Alias for -g pattern\n")
	fmt.Fprintf(os.Stderr, "  --strict                Strict matching: require all supplied patterns to match (AND)\n")
	fmt.Fprintf(os.Stderr, "  -i, -I, --ignore-case   Case-insensitive search\n")
	fmt.Fprintf(os.Stderr, "  -v, -V, --invert        Invert grep match\n")
	fmt.Fprintf(os.Stderr, "  -gi, -gv, -gvi, -giv    Combined grep flags\n\n")
	fmt.Fprintf(os.Stderr, "Shape Shortcuts:\n")
	fmt.Fprintf(os.Stderr, "  --tree                  Force hierarchical tree view\n")
	fmt.Fprintf(os.Stderr, "  --table                 Force tabular view\n")
	fmt.Fprintf(os.Stderr, "  --json, --raw           Force formatted JSON output\n")
	fmt.Fprintf(os.Stderr, "  --markdown, --md        Force markdown table output\n")
	fmt.Fprintf(os.Stderr, "  --csv, --tsv            Force delimited data output\n\n")
	fmt.Fprintf(os.Stderr, "Examples:\n")
	fmt.Fprintf(os.Stderr, "  curl https://integrate.api.nvidia.com/v1/models | tq -g 'google' -g 'deepseek'\n")
	fmt.Fprintf(os.Stderr, "  docker inspect container | tq -g 'nginx' -g 'running' --strict\n")
	fmt.Fprintf(os.Stderr, "  tq -u data.json\n\n")
	os.Exit(0)
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
