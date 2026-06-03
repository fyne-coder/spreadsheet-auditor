package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"spreadsheet-auditor/internal/audit"
	"spreadsheet-auditor/internal/reviewpack"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: spreadsheet-auditor scan WORKBOOK [--output PATH] [--review-pack PATH] [--export-csv PATH] [--exported-at RFC3339]")
		return 2
	}

	if args[0] != "scan" {
		fmt.Fprintf(os.Stderr, "unknown command %q\n", args[0])
		return 2
	}
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "scan requires a workbook path")
		return 2
	}

	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	output := fs.String("output", "", "write JSON report to this path")
	outputShort := fs.String("o", "", "write JSON report to this path")
	reviewPack := fs.String("review-pack", "", "write HTML review pack to this path")
	exportCSV := fs.String("export-csv", "", "write CSV review pack to this path")
	exportedAtFlag := fs.String("exported-at", "", "RFC3339 timestamp for HTML/CSV exports (defaults to now in UTC)")
	scanArgs := reorderScanArgs(args[1:])
	if err := fs.Parse(scanArgs); err != nil {
		return 2
	}

	workbook := fs.Arg(0)
	if workbook == "" {
		fmt.Fprintln(os.Stderr, "scan requires a workbook path")
		return 2
	}

	report, err := audit.AuditWorkbook(workbook)
	if err != nil {
		fmt.Fprintf(os.Stderr, "audit failed: %v\n", err)
		return 1
	}

	jsonTarget := *output
	if jsonTarget == "" {
		jsonTarget = *outputShort
	}
	reviewPackTarget := *reviewPack
	csvTarget := *exportCSV
	wroteOutput := jsonTarget != "" || reviewPackTarget != "" || csvTarget != ""

	exportedAt := time.Now().UTC()
	if *exportedAtFlag != "" {
		exportedAt, err = time.Parse(time.RFC3339, *exportedAtFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid --exported-at: %v\n", err)
			return 2
		}
		exportedAt = exportedAt.UTC()
	}

	if jsonTarget != "" {
		payload, err := report.CanonicalJSON()
		if err != nil {
			fmt.Fprintf(os.Stderr, "encode failed: %v\n", err)
			return 1
		}
		if err := os.WriteFile(jsonTarget, payload, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "write failed: %v\n", err)
			return 1
		}
	}

	if reviewPackTarget != "" {
		if err := reviewpack.WriteExport(report, reviewPackTarget, reviewpack.ExportOptions{
			Format:     reviewpack.FormatHTML,
			ExportedAt: exportedAt,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "write failed: %v\n", err)
			return 1
		}
	}

	if csvTarget != "" {
		if err := reviewpack.WriteExport(report, csvTarget, reviewpack.ExportOptions{
			Format:     reviewpack.FormatCSV,
			ExportedAt: exportedAt,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "write failed: %v\n", err)
			return 1
		}
	}

	if !wroteOutput {
		payload, err := report.CanonicalJSON()
		if err != nil {
			fmt.Fprintf(os.Stderr, "encode failed: %v\n", err)
			return 1
		}
		_, err = os.Stdout.Write(payload)
		return boolToExit(err == nil)
	}

	return 0
}

func boolToExit(ok bool) int {
	if ok {
		return 0
	}
	return 1
}

func reorderScanArgs(args []string) []string {
	if len(args) == 0 {
		return args
	}
	flags := make([]string, 0, len(args))
	positionals := make([]string, 0, len(args))
	for index := 0; index < len(args); index++ {
		arg := args[index]
		switch arg {
		case "--output", "-o", "--review-pack", "--export-csv", "--exported-at":
			flags = append(flags, arg)
			if index+1 < len(args) {
				index++
				flags = append(flags, args[index])
			}
		default:
			positionals = append(positionals, arg)
		}
	}
	return append(flags, positionals...)
}
