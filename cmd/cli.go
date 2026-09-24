package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/mattn/go-isatty"
	"github.com/abhi-vmlinuz/tquery/completions"
	"github.com/abhi-vmlinuz/tquery/pkg/engine"
	"github.com/abhi-vmlinuz/tquery/pkg/filter"
	"github.com/abhi-vmlinuz/tquery/pkg/pager"
	"github.com/abhi-vmlinuz/tquery/pkg/parser"
	"github.com/abhi-vmlinuz/tquery/pkg/render"
	"github.com/abhi-vmlinuz/tquery/pkg/tui"
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
	Columns     []string
	ListColumns bool
	SortEnabled bool
	SortBy      string
	SortDesc    bool
	Clipboard   bool
}

func Execute() {
	if len(os.Args) > 1 && (os.Args[1] == "completion" || os.Args[1] == "completions") {
		handleCompletion(os.Args[2:])
		return
	}

	cfg := parseFlags()

	if cfg.ShowVersion {
		fmt.Printf("tq version %s\n", Version)
		os.Exit(0)
	}

	rawJSON, err := readInput(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	isTTYStdin := isatty.IsTerminal(os.Stdin.Fd())
	isTTYStdout := isatty.IsTerminal(os.Stdout.Fd())

	runInteractive := cfg.Interactive
	if !cfg.Interactive && isTTYStdin && isTTYStdout && cfg.FilePath == "" && len(cfg.Patterns) == 0 && !cfg.Clipboard {
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

	// 2. Format Normalization & Smart Shape Auto-Detection
	normFmt, err := normalizeFormat(cfg.Format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	chosenFormat := normFmt
	if chosenFormat == "auto" {
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

	// Column listing when -c / --columns was passed without arguments
	if cfg.ListColumns && len(cfg.Columns) == 0 {
		if len(ds.Headers) == 0 {
			fmt.Fprintln(os.Stdout, "No columns detected.")
			return
		}
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "Columns (%d):\n", len(ds.Headers))
		for idx, h := range ds.Headers {
			fmt.Fprintf(&buf, "  %2d  %s\n", idx+1, h)
		}
		if err := pager.WriteOrPage(os.Stdout, buf.String(), cfg.NoPager); err != nil {
			fmt.Fprintf(os.Stderr, "Output error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// 4. In-memory Sorting (-s, --sort, --desc)
	if cfg.SortEnabled || cfg.SortBy != "" {
		sortCol := cfg.SortBy
		if sortCol == "" {
			if len(cfg.Columns) > 0 {
				sortCol = cfg.Columns[0]
			} else {
				sortCol = findPrimaryColumn(ds.Headers)
			}
		}
		ds = parser.SortDataStructure(ds, sortCol, cfg.SortDesc)
	}

	// 5. Column Projection (-c, --columns)
	if len(cfg.Columns) > 0 {
		ds = parser.ProjectColumns(ds, cfg.Columns)
	}

	// 6. Apply -<number> / --limit trimming
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

func normalizeFormat(fmtStr string) (string, error) {
	cleaned := strings.ToLower(strings.TrimSpace(fmtStr))
	switch cleaned {
	case "", "auto":
		return "auto", nil
	case "table":
		return "table", nil
	case "tree":
		return "tree", nil
	case "json", "raw":
		return "json", nil
	case "markdown", "md":
		return "markdown", nil
	case "csv":
		return "csv", nil
	case "tsv":
		return "tsv", nil
	default:
		return "", fmt.Errorf("unknown output format %q. Supported formats: auto, table, tree, json (raw), markdown (md), csv, tsv", fmtStr)
	}
}

func handleCompletion(args []string) {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		fmt.Println("Generate shell completion scripts for tq.")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  tq completion [bash|zsh|fish|install]")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  source <(tq completion bash)              # Load bash completion in current shell")
		fmt.Println("  tq completion zsh > ~/.zsh/completion/_tq  # Save zsh completion script")
		fmt.Println("  tq completion install                     # Install completion for your active shell")
		os.Exit(0)
	}

	shell := strings.ToLower(args[0])
	switch shell {
	case "bash":
		fmt.Print(completions.BashCompletion)
	case "zsh":
		fmt.Print(completions.ZshCompletion)
	case "fish":
		fmt.Print(completions.FishCompletion)
	case "install":
		installCompletions()
	default:
		fmt.Fprintf(os.Stderr, "Unknown shell %q. Supported options: bash, zsh, fish, install\n", shell)
		os.Exit(1)
	}
	os.Exit(0)
}

func installCompletions() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding home directory: %v\n", err)
		os.Exit(1)
	}

	activeShell := os.Getenv("SHELL")
	if strings.Contains(activeShell, "zsh") {
		paths := []string{
			filepath.Join(home, ".local/share/zsh/site-functions/_tq"),
			filepath.Join(home, ".zsh/completion/_tq"),
		}
		writeCompletionToPaths(paths, completions.ZshCompletion, "zsh")
	} else if strings.Contains(activeShell, "fish") {
		paths := []string{
			filepath.Join(home, ".config/fish/completions/tq.fish"),
		}
		writeCompletionToPaths(paths, completions.FishCompletion, "fish")
	} else {
		// Default to bash
		paths := []string{
			filepath.Join(home, ".local/share/bash-completion/completions/tq"),
			filepath.Join(home, ".bash_completion.d/tq"),
		}
		writeCompletionToPaths(paths, completions.BashCompletion, "bash")
	}
}

func writeCompletionToPaths(paths []string, content string, shellName string) {
	installed := false
	for _, p := range paths {
		dir := filepath.Dir(p)
		if err := os.MkdirAll(dir, 0755); err != nil {
			continue
		}
		if err := os.WriteFile(p, []byte(content), 0644); err == nil {
			fmt.Printf("[ok] Installed %s completion to %s\n", shellName, p)
			installed = true
		}
	}
	if !installed {
		fmt.Fprintf(os.Stderr, "Could not install %s completion automatically. Use: tq completion %s\n", shellName, shellName)
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
		case "--desc":
			cfg.SortDesc = true
			cfg.SortEnabled = true
			continue
		case "-cb", "--clipboard":
			cfg.Clipboard = true
			continue
		}

		// -c / --columns <val>
		if arg == "-c" || arg == "--columns" {
			if i+1 < len(rawArgs) && !isFlag(rawArgs[i+1]) && !isFile(rawArgs[i+1]) {
				for _, c := range strings.Split(rawArgs[i+1], ",") {
					if trimmed := strings.TrimSpace(c); trimmed != "" {
						cfg.Columns = append(cfg.Columns, trimmed)
					}
				}
				i++
			} else {
				cfg.ListColumns = true
			}
			continue
		}

		// -s / --sort <val>
		if arg == "-s" || arg == "--sort" {
			cfg.SortEnabled = true
			if i+1 < len(rawArgs) && !isFlag(rawArgs[i+1]) && !isFile(rawArgs[i+1]) {
				val := rawArgs[i+1]
				if strings.HasPrefix(val, "-") && len(val) > 1 {
					cfg.SortDesc = true
					cfg.SortBy = strings.TrimPrefix(val, "-")
				} else {
					cfg.SortBy = val
				}
				i++
			}
			continue
		}

		// -f / --format <val>
		if arg == "-f" || arg == "--format" {
			if i+1 < len(rawArgs) && !isFlag(rawArgs[i+1]) {
				cfg.Format = rawArgs[i+1]
				i++
			}
			continue
		}

		// -l / -L / --limit <val>
		if arg == "-l" || arg == "-L" || arg == "--limit" {
			if i+1 < len(rawArgs) && !isFlag(rawArgs[i+1]) {
				if num, err := strconv.Atoi(rawArgs[i+1]); err == nil && num > 0 {
					cfg.Limit = num
					i++
				}
			}
			continue
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
			if i+1 < len(rawArgs) && !isFlag(rawArgs[i+1]) {
				cfg.Patterns = append(cfg.Patterns, rawArgs[i+1])
				i++
			}
			continue
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

		// Long flags with = (e.g. --format=tree, --limit=10, --grep=meta, --columns=id,name)
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
			case "columns", "c":
				for _, col := range strings.Split(val, ",") {
					if trimmed := strings.TrimSpace(col); trimmed != "" {
						cfg.Columns = append(cfg.Columns, trimmed)
					}
				}
			case "sort", "s":
				cfg.SortEnabled = true
				if strings.HasPrefix(val, "-") && len(val) > 1 {
					cfg.SortDesc = true
					cfg.SortBy = strings.TrimPrefix(val, "-")
				} else {
					cfg.SortBy = val
				}
			case "clipboard", "cb":
				cfg.Clipboard = true
			case "desc":
				cfg.SortDesc = true
				cfg.SortEnabled = true
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
	fmt.Fprintf(os.Stderr, "Usage: tq [options] [--tree|--table|--json] [-<number>] [-c col1,col2] [-s col] [-cb] [-g <pat1> -g <pat2>] [--strict] [jq_query] [file]\n\n")
	fmt.Fprintf(os.Stderr, "tq (tquery) converts raw JSON & JQ streams into human-readable tables, trees, and interactive UI.\n\n")
	fmt.Fprintf(os.Stderr, "Data Shaping & Projection:\n")
	fmt.Fprintf(os.Stderr, "  -c, --columns <cols>    Cherry-pick columns by name (comma-separated, e.g. -c id,status)\n")
	fmt.Fprintf(os.Stderr, "  -s, --sort <column>     Sort records by column (numeric or alphabetic; use -s -col for descending)\n")
	fmt.Fprintf(os.Stderr, "  --desc                  Sort in descending order (when combined with -s)\n")
	fmt.Fprintf(os.Stderr, "  -cb, --clipboard        Read JSON data directly from system clipboard\n")
	fmt.Fprintf(os.Stderr, "  -<number>, -l <number>  Limit output rows (e.g. -10, -5, -l 20)\n\n")
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
	fmt.Fprintf(os.Stderr, "  curl https://integrate.api.nvidia.com/v1/models | tq -c id,owned_by -10\n")
	fmt.Fprintf(os.Stderr, "  docker inspect container | tq -g 'nginx' -g 'running' --strict\n")
	fmt.Fprintf(os.Stderr, "  tq -cb -s created --desc\n")
	fmt.Fprintf(os.Stderr, "  tq -u data.json\n\n")
	os.Exit(0)
}

func readInput(cfg Config) ([]byte, error) {
	if cfg.Clipboard {
		text, err := clipboard.ReadAll()
		if err != nil {
			return nil, fmt.Errorf("error reading clipboard: %w", err)
		}
		if strings.TrimSpace(text) == "" {
			return nil, fmt.Errorf("clipboard is empty")
		}
		return []byte(text), nil
	}

	if cfg.FilePath != "" {
		return os.ReadFile(cfg.FilePath)
	}

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return nil, fmt.Errorf("no input provided. Pipe JSON to stdin, specify a JSON file path, or use -cb to read from clipboard")
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

func findPrimaryColumn(headers []string) string {
	priority := []string{"id", "name", "title", "key", "model", "created", "timestamp", "date"}
	for _, p := range priority {
		for _, h := range headers {
			if strings.EqualFold(h, p) {
				return h
			}
		}
	}
	if len(headers) > 0 {
		return headers[0]
	}
	return ""
}

func isFlag(arg string) bool {
	if strings.HasPrefix(arg, "--") {
		return true
	}
	if !strings.HasPrefix(arg, "-") || len(arg) <= 1 {
		return false
	}
	if _, err := strconv.Atoi(arg[1:]); err == nil {
		return true // numeric row limit like -10, -5
	}
	knownFlags := []string{
		"-c", "-s", "-l", "-L", "-g", "-e", "-i", "-I", "-v", "-V", "-u", "-f", "-n", "-h",
		"-cb", "-gi", "-gv", "-gvi", "-giv", "-iv", "-vi",
	}
	for _, kf := range knownFlags {
		if arg == kf {
			return true
		}
	}
	return false
}

