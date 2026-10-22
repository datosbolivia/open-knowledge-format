package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/datosbolivia/open-knowledge-format/src/domain"
	"github.com/datosbolivia/open-knowledge-format/src/usecases"
)

const version = "0.2.0"

func printUsage() {
	fmt.Fprintf(os.Stderr, "Open Knowledge Format (OKF) Engine & CLI - v%s\n"+
		"A universal, vendor-neutral standard for open knowledge bundles and data products.\n\n"+
		"Usage:\n"+
		"  okf <command> [arguments]\n\n"+
		"Commands:\n"+
		"  validate    Validate a concept document, bundle directory, or URL against OKF v0.2\n"+
		"  visualize   Generate a self-contained interactive HTML graph visualization\n"+
		"  inspect     Display parsed frontmatter, ontology anchors, and contracts\n"+
		"  version     Display version and specification information\n\n"+
		"Run 'okf <command> -h' for more details on each command.\n", version)
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "validate":
		runValidate(os.Args[2:])
	case "visualize":
		runVisualize(os.Args[2:])
	case "inspect":
		runInspect(os.Args[2:])
	case "version", "--version", "-v":
		fmt.Printf("okf version %s (OKF/ODKF v0.2 Specification Engine)\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\nRun 'okf help' for available commands.\n", command)
		os.Exit(1)
	}
}

func runValidate(args []string) {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output validation report as JSON")
	timeoutSec := fs.Int("timeout", 5, "HTTP request timeout in seconds")
	fs.Parse(args)

	targets := fs.Args()
	if len(targets) == 0 {
		fmt.Fprintln(os.Stderr, "Error: missing target to validate (specify file path, bundle dir, or URL).")
		os.Exit(1)
	}

	uc := usecases.NewValidateUseCase(time.Duration(*timeoutSec) * time.Second)
	ctx := context.Background()

	hasErrors := false
	for _, target := range targets {
		report, err := uc.Validate(ctx, target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error validating %s: %v\n", target, err)
			hasErrors = true
			continue
		}

		if *jsonOutput {
			outBytes, _ := json.MarshalIndent(report, "", "  ")
			fmt.Println(string(outBytes))
		} else {
			if report.Valid {
				fmt.Printf("[OK] %s: VALID (0 errors, %d warnings)\n", target, report.TotalWarnings)
			} else {
				fmt.Printf("[FAIL] %s: INVALID (%d errors, %d warnings)\n", target, report.TotalErrors, report.TotalWarnings)
				hasErrors = true
			}
			for _, issue := range report.Issues {
				loc := ""
				if issue.Path != "" {
					loc = issue.Path
					if issue.Line > 0 {
						loc = fmt.Sprintf("%s:%d", loc, issue.Line)
					}
					loc = " (" + loc + ")"
				}
				fmt.Printf("  - [%s] %s: %s%s\n", issue.Severity, issue.Code, issue.Message, loc)
			}
		}

		if !report.Valid {
			hasErrors = true
		}
	}

	if hasErrors {
		os.Exit(1)
	}
}

func runVisualize(args []string) {
	fs := flag.NewFlagSet("visualize", flag.ExitOnError)
	bundlePath := fs.String("bundle", "", "Path to the OKF bundle root directory (required)")
	outPath := fs.String("out", "", "Output HTML path (default: <bundle>/viz.html)")
	bundleName := fs.String("name", "", "Display name for the bundle (default: directory name)")
	fs.Parse(args)

	if *bundlePath == "" {
		fmt.Fprintln(os.Stderr, "Error: --bundle is required.")
		fs.Usage()
		os.Exit(1)
	}

	uc := usecases.NewVisualizeUseCase()
	stats, err := uc.Generate(*bundlePath, *outPath, *bundleName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating visualization: %v\n", err)
		os.Exit(1)
	}

	targetOut := *outPath
	if targetOut == "" {
		targetOut = filepath.Join(*bundlePath, "viz.html")
	}

	fmt.Fprintf(os.Stderr, "Generated visualization: %d concept(s), %d edge(s), %d bytes -> %s\n",
		stats.ConceptsCount, stats.EdgesCount, stats.BytesWritten, targetOut)
}

func runInspect(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: missing concept file path.")
		os.Exit(1)
	}

	target := args[0]
	bytes, err := os.ReadFile(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	parts := strings.Split(string(bytes), "---")
	if len(parts) < 3 {
		fmt.Fprintln(os.Stderr, "Error: no YAML frontmatter delimiters '---' found.")
		os.Exit(1)
	}

	report := domain.ValidateConceptContent(string(bytes), target)
	fmt.Printf("Concept: %s\n", target)
	fmt.Printf("Status:  Valid=%v (Errors: %d, Warnings: %d)\n\n", report.Valid, report.TotalErrors, report.TotalWarnings)
	fmt.Println("Frontmatter YAML:")
	fmt.Println("--------------------------------------------------")
	fmt.Println(strings.TrimSpace(parts[1]))
	fmt.Println("--------------------------------------------------")
}
