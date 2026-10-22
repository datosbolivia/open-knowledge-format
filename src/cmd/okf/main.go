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
		"  init        Initialize a new OKF v0.2 bundle directory structure\n"+
		"  add         Create a new concept Markdown document with frontmatter template\n"+
		"  link        Insert semantic link between two concepts in the knowledge graph\n"+
		"  audit       Audit bundle for broken links, orphan nodes, circular refs, and staleness\n"+
		"  graph       Extract directed adjacency graph of concepts and semantic relations (JSON)\n"+
		"  index       Scan bundle and generate/update index.md for progressive disclosure\n"+
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
	case "init":
		runInit(os.Args[2:])
	case "add":
		runAdd(os.Args[2:])
	case "link":
		runLink(os.Args[2:])
	case "audit":
		runAudit(os.Args[2:])
	case "graph":
		runGraph(os.Args[2:])
	case "index":
		runIndex(os.Args[2:])
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

func runInit(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	title := fs.String("title", "", "Human-readable bundle title")
	desc := fs.String("desc", "", "Short description of the knowledge bundle")
	parseFlexibleFlags(fs, args)

	targetDir := "."
	if len(fs.Args()) > 0 {
		targetDir = fs.Args()[0]
	}

	err := usecases.InitBundle(targetDir, *title, *desc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing bundle: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[OK] Initialized OKF v0.2 bundle at '%s'\n", targetDir)
	fmt.Println("  Created index.md (root index)")
	fmt.Println("  Created log.md (append-only changelog)")
	fmt.Println("  Created concepts/sample_dataset.md (conforming concept)")
}

func runAdd(args []string) {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	cType := fs.String("type", "concept", "Concept type (concept, dataset, service, schema, index)")
	title := fs.String("title", "", "Human-readable title")
	desc := fs.String("desc", "", "Short description")
	author := fs.String("author", "Agent <agent@datamesh.local>", "Author or agent identifier")
	cadence := fs.String("cadence", "P1M", "Freshness cadence ISO-8601 (e.g. P1D, P1W, P1M)")
	parseFlexibleFlags(fs, args)

	if len(fs.Args()) == 0 {
		fmt.Fprintln(os.Stderr, "Error: missing concept file path (e.g. 'okf add concepts/my_topic.md').")
		os.Exit(1)
	}

	targetPath := fs.Args()[0]
	createdPath, err := usecases.AddConcept(usecases.AddConceptParams{
		Path:        targetPath,
		Type:        *cType,
		Title:       *title,
		Description: *desc,
		Author:      *author,
		Cadence:     *cadence,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating concept: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[OK] Created concept node '%s'\n", createdPath)
}

func runLink(args []string) {
	fs := flag.NewFlagSet("link", flag.ExitOnError)
	rel := fs.String("rel", "", "Semantic predicate or link anchor text")
	parseFlexibleFlags(fs, args)

	if len(fs.Args()) < 2 {
		fmt.Fprintln(os.Stderr, "Error: 'okf link' requires source and target concept files (e.g. 'okf link source.md target.md -rel \"Correlates With\"').")
		os.Exit(1)
	}

	source := fs.Args()[0]
	target := fs.Args()[1]

	err := usecases.LinkConcepts(source, target, *rel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error linking concepts: %v\n", err)
		os.Exit(1)
	}

	relText := *rel
	if relText == "" {
		relText = "Related concept"
	}
	fmt.Printf("[OK] Linked '%s' -> '%s' (predicate: %s)\n", source, target, relText)
}

func runAudit(args []string) {
	fs := flag.NewFlagSet("audit", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output audit report as JSON")
	parseFlexibleFlags(fs, args)

	targetDir := "."
	if len(fs.Args()) > 0 {
		targetDir = fs.Args()[0]
	}

	report, err := usecases.AuditBundle(targetDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error auditing bundle: %v\n", err)
		os.Exit(1)
	}

	if *jsonOutput {
		outBytes, _ := json.MarshalIndent(report, "", "  ")
		fmt.Println(string(outBytes))
	} else {
		if report.Passed {
			fmt.Printf("[OK] Bundle audit PASSED for '%s' (%d files scanned)\n", targetDir, report.TotalFiles)
		} else {
			fmt.Printf("[FAIL] Bundle audit FAILED for '%s' (%d files scanned)\n", targetDir, report.TotalFiles)
		}
		fmt.Printf("  Broken links:         %d\n", report.BrokenLinks)
		fmt.Printf("  Orphan nodes:         %d\n", report.OrphanNodes)
		fmt.Printf("  Circular references:  %d\n", report.CircularRefs)
		fmt.Printf("  Stale concepts:       %d\n", report.StaleNodes)

		if len(report.Issues) > 0 {
			fmt.Println("\nIssues identified:")
			for _, issue := range report.Issues {
				fmt.Printf("  - [%s] %s (%s): %s\n", issue.Severity, issue.Rule, issue.File, issue.Message)
			}
		}
	}

	if !report.Passed {
		os.Exit(1)
	}
}

func runGraph(args []string) {
	fs := flag.NewFlagSet("graph", flag.ExitOnError)
	parseFlexibleFlags(fs, args)

	targetDir := "."
	if len(fs.Args()) > 0 {
		targetDir = fs.Args()[0]
	}

	report, err := usecases.BuildGraphAdjacency(targetDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building graph: %v\n", err)
		os.Exit(1)
	}

	outBytes, _ := json.MarshalIndent(report, "", "  ")
	fmt.Println(string(outBytes))
}

func runIndex(args []string) {
	fs := flag.NewFlagSet("index", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output index summary as JSON")
	parseFlexibleFlags(fs, args)

	targetDir := "."
	if len(fs.Args()) > 0 {
		targetDir = fs.Args()[0]
	}

	report, err := usecases.GenerateIndex(targetDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating index: %v\n", err)
		os.Exit(1)
	}

	if *jsonOutput {
		outBytes, _ := json.MarshalIndent(report, "", "  ")
		fmt.Println(string(outBytes))
	} else {
		fmt.Printf("[OK] Re-indexed bundle '%s' -> %s\n", targetDir, report.IndexFilePath)
		fmt.Printf("  Concepts indexed: %d\n", report.ConceptsIndexed)
		fmt.Printf("  Types found: %s\n", strings.Join(report.TypesFound, ", "))
	}
}

func runValidate(args []string) {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output validation report as JSON")
	timeoutSec := fs.Int("timeout", 5, "HTTP request timeout in seconds")
	parseFlexibleFlags(fs, args)

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
	parseFlexibleFlags(fs, args)

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
	fmt.Printf("Valid: %t (Errors: %d, Warnings: %d)\n", report.Valid, report.TotalErrors, report.TotalWarnings)
	for _, issue := range report.Issues {
		fmt.Printf("  - [%s] %s: %s\n", issue.Severity, issue.Code, issue.Message)
	}
}

func parseFlexibleFlags(fs *flag.FlagSet, args []string) {
	var flags, nonFlags []string
	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "-") {
			flags = append(flags, args[i])
			// If not a bool flag and next token doesn't start with '-'
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				flags = append(flags, args[i+1])
				i++
			}
		} else {
			nonFlags = append(nonFlags, args[i])
		}
	}
	_ = fs.Parse(append(flags, nonFlags...))
}
