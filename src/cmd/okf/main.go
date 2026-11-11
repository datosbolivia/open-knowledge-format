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

	"github.com/datosbolivia/open-knowledge-format/src/adapters/outbound/manifests"
	"github.com/datosbolivia/open-knowledge-format/src/domain"
	"github.com/datosbolivia/open-knowledge-format/src/usecases"
)

const version = "0.3.0-dev.1"

func printUsage() {
	fmt.Fprintf(os.Stderr, "Open Knowledge Format (OKF) Engine & CLI - v%s\n"+
		"A universal, vendor-neutral standard for open knowledge bundles and data products.\n\n"+
		"Usage:\n"+
		"  okf <command> [arguments]\n\n"+
		"Commands:\n"+
		"  init        Initialize a new OKF v0.2 bundle directory structure\n"+
		"  add         Create a new concept Markdown document with frontmatter template\n"+
		"  link        Insert semantic link between two concepts in the knowledge graph\n"+
		"  run         Execute a pipeline script and record provenance in log.md\n"+
		"  audit       Audit bundle for broken links, orphan nodes, circular refs, and staleness\n"+
		"  graph       Extract directed adjacency graph of concepts and semantic relations (JSON)\n"+
		"  index       Scan bundle and generate/update index.md for progressive disclosure\n"+
		"  validate    Validate a concept document, bundle directory, or URL against OKF v0.2\n"+
		"  diff        Classify schema & concept changes between two bundle versions (PATCH/MINOR/MAJOR)\n"+
		"  attest      Verify an Attested Computation receipt and emit a JSON AttestationVerdict\n"+
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
	case "run":
		runRun(os.Args[2:])
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
	case "diff":
		runDiff(os.Args[2:])
	case "attest":
		runAttest(os.Args[2:])
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
	bundlePath := fs.String("bundle", "", "Path to the OKF bundle root directory")
	outPath := fs.String("out", "", "Output HTML path (default: <bundle>/viz.html)")
	bundleName := fs.String("name", "", "Display name for the bundle (default: directory name)")
	parseFlexibleFlags(fs, args)

	targetBundle := *bundlePath
	if targetBundle == "" && len(fs.Args()) > 0 {
		targetBundle = fs.Args()[0]
	}
	if targetBundle == "" {
		fmt.Fprintln(os.Stderr, "Error: bundle directory path or --bundle is required.")
		fs.Usage()
		os.Exit(1)
	}

	uc := usecases.NewVisualizeUseCase()
	stats, err := uc.Generate(targetBundle, *outPath, *bundleName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating visualization: %v\n", err)
		os.Exit(1)
	}

	targetOut := *outPath
	if targetOut == "" {
		targetOut = filepath.Join(targetBundle, "viz.html")
	}

	fmt.Fprintf(os.Stderr, "Generated visualization: %d concept(s), %d edge(s), %d bytes -> %s\n",
		stats.ConceptsCount, stats.EdgesCount, stats.BytesWritten, targetOut)
}

func runInspect(args []string) {
	fs := flag.NewFlagSet("inspect", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output inspected node metadata and validation report as JSON")
	parseFlexibleFlags(fs, args)

	if len(fs.Args()) == 0 {
		fmt.Fprintln(os.Stderr, "Error: missing concept file or bundle directory path (e.g. 'okf inspect index.md').")
		fs.Usage()
		os.Exit(1)
	}

	target := fs.Args()[0]
	reader := manifests.NewUnifiedMetadataReader("", 5*time.Second)
	dp, dpErr := reader.ReadNode(context.Background(), target)

	var report domain.ValidationReport
	if fi, statErr := os.Stat(target); statErr == nil && !fi.IsDir() {
		if rawBytes, rErr := os.ReadFile(target); rErr == nil {
			if r := domain.ValidateConceptContent(string(rawBytes), target); r != nil {
				report = *r
			}
		}
	} else {
		valUC := usecases.NewValidateUseCase(5 * time.Second)
		if r, vErr := valUC.Validate(context.Background(), target); vErr == nil && r != nil {
			report = *r
		}
	}

	if *jsonOutput {
		payload := map[string]interface{}{
			"target":     target,
			"valid":      report.Valid,
			"validation": report,
		}
		if dpErr == nil && dp != nil {
			payload["dataproduct"] = dp
		}
		outBytes, _ := json.MarshalIndent(payload, "", "  ")
		fmt.Println(string(outBytes))
		return
	}

	fmt.Printf("Target:  %s\n", target)
	if dpErr == nil && dp != nil {
		fmt.Printf("Title:   %s\n", dp.Manifest.Title)
		fmt.Printf("Type:    %s\n", dp.Manifest.Type)
		if dp.Manifest.Status != "" {
			fmt.Printf("Status:  %s\n", dp.Manifest.Status)
		}
		if len(dp.Manifest.Dimensions) > 0 {
			fmt.Printf("Dimensions: %s\n", strings.Join(dp.Manifest.Dimensions, ", "))
		}
		anchors := append([]string{}, dp.Manifest.SKOS...)
		anchors = append(anchors, dp.Manifest.ExactMatch...)
		if len(anchors) > 0 {
			fmt.Printf("Ontology Anchors (SKOS): %s\n", strings.Join(anchors, ", "))
		}
		if len(dp.Manifest.Contracts) > 0 {
			fmt.Printf("Contracts (%d):\n", len(dp.Manifest.Contracts))
			for _, c := range dp.Manifest.Contracts {
				ref := c.Path
				if ref == "" {
					ref = c.Ref
				}
				if c.Binary != "" {
					fmt.Printf("  - type=%s binary=\"%s %s\"\n", c.Type, c.Binary, c.Subcommand)
				} else {
					fmt.Printf("  - type=%s path=%s\n", c.Type, ref)
				}
			}
		}
		if len(dp.Resources) > 0 {
			fmt.Printf("Resources (%d):\n", len(dp.Resources))
			for _, r := range dp.Resources {
				fmt.Printf("  - %s (%s): %s [%d schema field(s)]\n", r.Name, r.Format, r.Path, len(r.Fields))
			}
		}
	}
	fmt.Printf("Valid:   %t (Errors: %d, Warnings: %d)\n", report.Valid, report.TotalErrors, report.TotalWarnings)
	for _, issue := range report.Issues {
		fmt.Printf("  - [%s] %s: %s\n", issue.Severity, issue.Code, issue.Message)
	}
}

func isBooleanFlag(arg string) bool {
	clean := strings.TrimLeft(strings.TrimSpace(arg), "-")
	if idx := strings.Index(clean, "="); idx != -1 {
		return true
	}
	switch clean {
	case "json", "strict", "v", "quiet", "h", "help":
		return true
	default:
		return false
	}
}

func parseFlexibleFlags(fs *flag.FlagSet, args []string) {
	var flags, nonFlags []string
	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "-") {
			flags = append(flags, args[i])
			// If not a bool flag and next token doesn't start with '-'
			if !isBooleanFlag(args[i]) && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				flags = append(flags, args[i+1])
				i++
			}
		} else {
			nonFlags = append(nonFlags, args[i])
		}
	}
	_ = fs.Parse(append(flags, nonFlags...))
}

func runRun(args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	target := fs.String("target", "", "Target triad or dataset name (e.g. 'my_catalog:dataset:resource')")
	jsonOutput := fs.Bool("json", false, "Output execution metrics as JSON")
	bundleDir := fs.String("bundle", "", "Path to bundle root (default: auto-detected)")
	parseFlexibleFlags(fs, args)

	if len(fs.Args()) == 0 {
		fmt.Fprintln(os.Stderr, "Error: missing script path to execute (e.g. 'okf run pipeline.py -target my_dataset').")
		os.Exit(1)
	}

	scriptPath := fs.Args()[0]
	scriptArgs := fs.Args()[1:]

	res, err := usecases.ExecuteScript(usecases.RunParams{
		ScriptPath: scriptPath,
		Target:     *target,
		Args:       scriptArgs,
		BundleDir:  *bundleDir,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing script: %v\n", err)
		os.Exit(1)
	}

	if *jsonOutput {
		outBytes, _ := json.MarshalIndent(res, "", "  ")
		fmt.Println(string(outBytes))
	} else {
		statusStr := "SUCCESS"
		if !res.Success {
			statusStr = fmt.Sprintf("FAILED (exit %d)", res.ExitCode)
		}
		fmt.Printf("[%s] %s -> %d ms (sha256: %s)\n", statusStr, res.ScriptPath, res.DurationMs, res.ScriptHash[:12])
		if res.Stdout != "" {
			fmt.Print(res.Stdout)
		}
		if res.Stderr != "" {
			fmt.Fprint(os.Stderr, res.Stderr)
		}
	}

	if !res.Success {
		os.Exit(res.ExitCode)
	}
}

func runDiff(args []string) {
	fs := flag.NewFlagSet("diff", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output diff report as JSON")
	parseFlexibleFlags(fs, args)

	if len(fs.Args()) < 2 {
		fmt.Fprintln(os.Stderr, "Error: 'okf diff' requires <bundle_v1> <bundle_v2>.")
		os.Exit(1)
	}

	report, err := usecases.DiffBundles(fs.Args()[0], fs.Args()[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error computing diff: %v\n", err)
		os.Exit(1)
	}

	if *jsonOutput {
		outBytes, _ := json.MarshalIndent(report, "", "  ")
		fmt.Println(string(outBytes))
		return
	}

	fmt.Printf("Semantic Diff: %s -> %s\n", report.OldTarget, report.NewTarget)
	fmt.Printf("Recommended Version Bump: %s (%d breaking, %d minor, %d patch)\n",
		report.RecommendedBump, report.BreakingChanges, report.MinorChanges, report.PatchChanges)
	for _, c := range report.Changes {
		fmt.Printf("  - [%s] (%s) %s: %s\n", c.Level, c.Category, c.Target, c.Description)
	}
}

func runAttest(args []string) {
	fs := flag.NewFlagSet("attest", flag.ExitOnError)
	sqlFlag := fs.String("sql", "", "Executed SQL or JSON receipt to verify against sanctioned computation")
	paramFlag := fs.String("param", "", "Comma-separated key=val parameter bindings (e.g. year=2026)")
	parseFlexibleFlags(fs, args)

	if len(fs.Args()) == 0 {
		fmt.Fprintln(os.Stderr, "Error: 'okf attest' requires a concept file path (e.g. 'okf attest computations/revenue-ytd.md').")
		os.Exit(1)
	}

	params := make(map[string]string)
	if strings.TrimSpace(*paramFlag) != "" {
		for _, kv := range strings.Split(*paramFlag, ",") {
			parts := strings.SplitN(strings.TrimSpace(kv), "=", 2)
			if len(parts) == 2 {
				params[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	verdict, err := usecases.AttestComputation(fs.Args()[0], *sqlFlag, params)
	outBytes, _ := json.MarshalIndent(verdict, "", "  ")
	fmt.Println(string(outBytes))
	if err != nil || verdict.Verdict == "refuted" || verdict.Verdict == "error" {
		os.Exit(1)
	}
}
