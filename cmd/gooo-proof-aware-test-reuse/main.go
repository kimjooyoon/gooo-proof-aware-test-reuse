package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/gooo-proof-aware-test-reuse/internal/engine"
)

func main() {
	if len(os.Args) < 2 {
		fatal("usage: gooo-proof-aware-test-reuse <run|suite> [flags]")
	}
	switch os.Args[1] {
	case "run":
		run(os.Args[2:])
	case "suite":
		suite(os.Args[2:])
	default:
		fatal("command must be run or suite")
	}
}

func run(args []string) {
	flags := flag.NewFlagSet("run", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	meta := flags.String("meta", ".gooo/proof-aware-test-reuse.gooo", "authoritative .gooo semantics")
	contract := flags.String("contract", "contracts/denominator-v1.json", "fixed canonical denominator")
	source := flags.String("source", "", "input .gooo program")
	receipts := flags.String("receipts", "fixtures/receipts", "immutable receipt directory")
	out := flags.String("out", "", "caller-owned output directory")
	mode := flags.String("mode", "incremental", "full or incremental")
	if err := flags.Parse(args); err != nil {
		os.Exit(2)
	}
	if *source == "" || *out == "" {
		fatal("--source and --out are required")
	}
	report, err := engine.Run(engine.RunOptions{
		MetaPath: *meta, ContractPath: *contract, SourcePath: *source,
		ReceiptsDir: *receipts, OutputDir: *out, Mode: *mode,
	})
	if err != nil {
		fatal(err.Error())
	}
	printJSON(struct {
		CaseID   string `json:"case_id"`
		Decision string `json:"decision"`
		Action   string `json:"action"`
	}{report.CaseID, report.Decision, report.Plan.Action})
}

func suite(args []string) {
	flags := flag.NewFlagSet("suite", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	meta := flags.String("meta", ".gooo/proof-aware-test-reuse.gooo", "authoritative .gooo semantics")
	contract := flags.String("contract", "contracts/denominator-v1.json", "fixed canonical denominator")
	cases := flags.String("cases", "fixtures/cases", "canonical .gooo case directory")
	receipts := flags.String("receipts", "fixtures/receipts", "immutable receipt directory")
	out := flags.String("out", "", "caller-owned output directory")
	mode := flags.String("mode", "incremental", "full or incremental")
	if err := flags.Parse(args); err != nil {
		os.Exit(2)
	}
	if *out == "" {
		fatal("--out is required")
	}
	report, err := engine.RunSuite(engine.SuiteOptions{
		MetaPath: *meta, ContractPath: *contract, CasesDir: *cases,
		ReceiptsDir: *receipts, OutputDir: *out, Mode: *mode,
	})
	if err != nil {
		fatal(err.Error())
	}
	printJSON(struct {
		Decision string `json:"decision"`
		Total    int    `json:"total"`
		Selected int    `json:"selected"`
		Reused   int    `json:"reused"`
		Unknown  int    `json:"unknown"`
		Refuted  int    `json:"refuted"`
	}{report.Decision, report.Actual.Total, report.Actual.Selected, report.Actual.Reused, report.Actual.Unknown, report.Actual.Failed})
}

func printJSON(value any) {
	data, err := json.Marshal(value)
	if err != nil {
		fatal(err.Error())
	}
	fmt.Println(string(data))
}

func fatal(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
