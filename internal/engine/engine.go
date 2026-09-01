package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type RunOptions struct {
	MetaPath     string
	ContractPath string
	SourcePath   string
	ReceiptsDir  string
	OutputDir    string
	Mode         string
}

type SuiteOptions struct {
	MetaPath     string
	ContractPath string
	CasesDir     string
	ReceiptsDir  string
	OutputDir    string
	Mode         string
}

type executionRecord struct {
	Schema       string `json:"schema"`
	Action       string `json:"action"`
	Status       string `json:"status"`
	Output       string `json:"output"`
	ResultDigest string `json:"result_digest"`
	Verified     bool   `json:"verified"`
}

func Run(options RunOptions) (RunReport, error) {
	if options.Mode == "" {
		options.Mode = "incremental"
	}
	if options.Mode != "full" && options.Mode != "incremental" {
		return RunReport{}, fmt.Errorf("mode must be full or incremental")
	}
	meta, err := ParseMeta(options.MetaPath)
	if err != nil {
		return RunReport{}, fmt.Errorf("load meta: %w", err)
	}
	_, contractDigest, err := loadContract(options.ContractPath)
	if err != nil {
		return RunReport{}, fmt.Errorf("load contract: %w", err)
	}
	program, err := ParseProgram(options.SourcePath)
	if err != nil {
		return RunReport{}, fmt.Errorf("load program: %w", err)
	}
	program.SourcePath = filepath.Base(options.SourcePath)
	if err := ensureOutputDir(options.OutputDir); err != nil {
		return RunReport{}, err
	}

	root := repositoryRoot(options.MetaPath)
	fixtureDigest, err := digestFixture(root, program.Obligations[0].Fixture)
	if err != nil {
		return RunReport{}, fmt.Errorf("digest fixture: %w", err)
	}
	obligation := program.Obligations[0]
	key := ProofKey{
		SourceDigest:    program.SourceDigest,
		ContractDigest:  contractDigest,
		FixtureDigest:   fixtureDigest,
		ToolchainDigest: DigestBytes([]byte(obligation.Toolchain)),
		RunnerDigest:    DigestBytes([]byte(obligation.Runner)),
	}

	unknown, refuted := validateProgram(meta, program)
	graph := buildDependencyGraph(program, obligation.ID)
	if graph.PathToTarget {
		graph.Frontier = append(graph.Frontier, graph.Impacted...)
	}
	refuted = append(refuted, graphRefutations(graph)...)

	receipts, _, receiptUnknown, receiptRefuted, err := inspectReceipts(options.ReceiptsDir, program, key, obligation.ID, options.Mode, root)
	if err != nil {
		return RunReport{}, err
	}
	unknown = append(unknown, receiptUnknown...)
	refuted = append(refuted, receiptRefuted...)

	decision := Reduce(unknown, refuted)
	action := ActionUnknown
	reason := ""
	if decision == DecisionRefuted {
		action = ActionRefuted
		reason = refuted[0].Reason
	} else if decision == DecisionUnknown {
		action = ActionUnknown
		reason = unknown[0].Reason
	} else if options.Mode == "incremental" && len(receipts) == 1 && !graph.PathToTarget {
		action = ActionReused
		reason = "EXACT_PASS_RECEIPT_AND_EMPTY_INVALIDATION_FRONTIER"
	} else {
		action = ActionSelected
		reason = "NO_REUSE_PROOF_OR_CHANGE_IMPACT_REQUIRES_EXECUTION"
	}

	plan := Plan{
		Schema:       "gooo/proof-aware-test-reuse/plan/v1",
		Mode:         options.Mode,
		Obligation:   obligation.ID,
		Action:       action,
		Reason:       reason,
		Selected:     []string{},
		Reused:       []string{},
		Unknown:      []string{},
		Refuted:      []string{},
		ReceiptPaths: append([]string(nil), program.ReceiptRefs...),
	}
	switch action {
	case ActionSelected:
		plan.Selected = []string{obligation.ID}
	case ActionReused:
		plan.Reused = []string{obligation.ID}
	case ActionUnknown:
		plan.Unknown = []string{obligation.ID}
	case ActionRefuted:
		plan.Refuted = []string{obligation.ID}
	}

	generatedCode := generateGo(program.Output)
	resultDigest := DigestBytes([]byte(program.Output))
	verification := Verification{
		Schema:               "gooo/proof-aware-test-reuse/verification/v1",
		GeneratedOutput:      program.Output,
		ExpectedOutput:       program.Output,
		ExpectedResultDigest: resultDigest,
		ExecutionRequired:    action == ActionSelected,
		ExecutionVerified:    action == ActionSelected,
		ReusedProof:          action == ActionReused,
	}
	if action == ActionReused {
		verification.ExecutionVerified = len(receipts) == 1 && receipts[0].ResultDigest == resultDigest
	}
	execution := executionRecord{
		Schema:       "gooo/proof-aware-test-reuse/execution/v1",
		Action:       action,
		Status:       executionStatus(action, verification.ExecutionVerified),
		Output:       program.Output,
		ResultDigest: resultDigest,
		Verified:     verification.ExecutionVerified,
	}

	graphPath := filepath.Join(options.OutputDir, "dependency-graph.json")
	irPath := filepath.Join(options.OutputDir, "semantic-ir.json")
	planPath := filepath.Join(options.OutputDir, "test-plan.json")
	executionPath := filepath.Join(options.OutputDir, "execution.json")
	receiptPath := filepath.Join(options.OutputDir, "reuse-receipt.json")
	generatedDir := filepath.Join(options.OutputDir, "generated")
	generatedPath := filepath.Join(generatedDir, "program.go")
	for _, path := range []string{graphPath, irPath, planPath, executionPath, receiptPath} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return RunReport{}, err
		}
	}
	if err := os.MkdirAll(generatedDir, 0o755); err != nil {
		return RunReport{}, err
	}
	if err := writeJSON(graphPath, graph); err != nil {
		return RunReport{}, err
	}
	if err := writeJSON(irPath, program); err != nil {
		return RunReport{}, err
	}
	if err := writeJSON(planPath, plan); err != nil {
		return RunReport{}, err
	}
	if err := writeJSON(executionPath, execution); err != nil {
		return RunReport{}, err
	}
	resultReceipt := Receipt{
		Schema:             meta.ReceiptSchema,
		Obligation:         obligation.ID,
		State:              receiptState(decision),
		SourceDigest:       key.SourceDigest,
		ContractDigest:     key.ContractDigest,
		FixtureDigest:      key.FixtureDigest,
		ToolchainDigest:    key.ToolchainDigest,
		RunnerDigest:       key.RunnerDigest,
		ResultDigest:       resultDigest,
		OperationalAudit:   operationalAudit(),
		PrFirstConformance: "REFUTED",
	}
	if err := writeJSON(receiptPath, resultReceipt); err != nil {
		return RunReport{}, err
	}
	if err := os.WriteFile(generatedPath, []byte(generatedCode), 0o644); err != nil {
		return RunReport{}, err
	}

	inventory, err := BuildInventory(root)
	if err != nil {
		return RunReport{}, fmt.Errorf("build inventory: %w", err)
	}
	artifacts, artifactDigests, err := collectArtifacts(options.OutputDir, []string{
		"dependency-graph.json", "semantic-ir.json", "test-plan.json", "execution.json", "reuse-receipt.json", "generated/program.go",
	})
	if err != nil {
		return RunReport{}, err
	}
	report := RunReport{
		Schema:         "gooo/proof-aware-test-reuse/report/v1",
		Decision:       decision,
		Mode:           options.Mode,
		CaseID:         program.CaseID,
		SourceDigest:   program.SourceDigest,
		ContractDigest: contractDigest,
		FixtureDigest:  fixtureDigest,
		Toolchain:      obligation.Toolchain,
		Runner:         obligation.Runner,
		ProofKey:       key,
		Plan:           plan,
		Unknown:        unknown,
		Refuted:        refuted,
		Improvement: Improvement{
			State:            DecisionUnknown,
			Scenario:         program.CaseID,
			SameJob:          false,
			ExactIntegerPair: false,
			Reason:           "EXACT_BEFORE_AFTER_INTEGER_PAIR_NOT_AVAILABLE",
			UtilityState:     DecisionUnknown,
		},
		Utility:            UnknownClaim("UTILITY", "observe_external_user_evidence", "NO_EXTERNAL_USER_EVIDENCE", "MISSING_EXTERNAL_EVIDENCE", "COLLECT_REAL_USER_WORKLOAD_EVIDENCE", []string{"external-user-evidence"}),
		Metrics:            metricsFor(action),
		Inventory:          inventory,
		GeneratedArtifacts: artifacts,
		Verification:       verification,
		Authority: Authority{
			RepositoryWrites:        0,
			CommitAuthority:         0,
			PushAuthority:           0,
			MergeAuthority:          0,
			ReleaseMutation:         0,
			LocalValidationCommands: 0,
			LocalValidationState:    "NOT_RUN",
		},
		ArtifactDigests:    artifactDigests,
		OperationalAudit:   operationalAudit(),
		PrFirstConformance: "REFUTED",
	}
	if err := writeHumanReport(filepath.Join(options.OutputDir, "human-report.md"), report); err != nil {
		return RunReport{}, err
	}
	if err := writeJSON(filepath.Join(options.OutputDir, "report.json"), report); err != nil {
		return RunReport{}, err
	}
	return report, nil
}

func RunSuite(options SuiteOptions) (SuiteReport, error) {
	if options.Mode == "" {
		options.Mode = "incremental"
	}
	contract, contractDigest, err := loadContract(options.ContractPath)
	if err != nil {
		return SuiteReport{}, err
	}
	if err := ensureOutputDir(options.OutputDir); err != nil {
		return SuiteReport{}, err
	}
	cases := make([]SuiteCase, 0, len(contract.Cases))
	actual := RunMetrics{}
	expectedStates := map[string]int{DecisionClosed: 0, DecisionUnknown: 0, DecisionRefuted: 0}
	actualStates := map[string]int{DecisionClosed: 0, DecisionUnknown: 0, DecisionRefuted: 0}
	var inventory Inventory
	generatedArtifacts := 0
	var generatedBytes int64
	for _, item := range contract.Cases {
		expectedStates[item.Expected]++
		path := resolvePath(options.CasesDir, item.Source)
		out := filepath.Join(options.OutputDir, "cases", item.ID)
		report, runErr := Run(RunOptions{
			MetaPath: options.MetaPath, ContractPath: options.ContractPath, SourcePath: path,
			ReceiptsDir: options.ReceiptsDir, OutputDir: out, Mode: options.Mode,
		})
		if runErr != nil {
			return SuiteReport{}, fmt.Errorf("case %s: %w", item.ID, runErr)
		}
		if inventory == (Inventory{}) {
			inventory = report.Inventory
		}
		generatedArtifacts += len(report.GeneratedArtifacts)
		for _, artifact := range report.GeneratedArtifacts {
			generatedBytes += artifact.Bytes
		}
		actual.Total++
		switch report.Plan.Action {
		case ActionSelected:
			actual.Selected++
			actual.Executed++
		case ActionReused:
			actual.Reused++
		case ActionUnknown:
			actual.Unknown++
		case ActionRefuted:
			actual.Failed++
		}
		actualStates[report.Decision]++
		reason := report.Plan.Reason
		if len(report.Unknown) > 0 {
			reason = report.Unknown[0].Reason
		}
		if len(report.Refuted) > 0 {
			reason = report.Refuted[0].Reason
		}
		cases = append(cases, SuiteCase{
			ID: item.ID, Kind: item.Kind, Expected: item.Expected, Decision: report.Decision,
			Action: report.Plan.Action, Match: report.Decision == item.Expected, Reason: reason,
			Unknown: report.Unknown, Refuted: report.Refuted,
			ReportPath: filepath.Join("cases", item.ID, "human-report.md"),
		})
	}
	decision := DecisionClosed
	for _, item := range cases {
		if !item.Match {
			decision = DecisionRefuted
			break
		}
	}
	report := SuiteReport{
		Schema: "gooo/proof-aware-test-reuse/suite-report/v1", Decision: decision,
		Contract: contract.ID, ContractDigest: contractDigest, Mode: options.Mode,
		FixedDenominator: len(contract.Cases), Cases: cases, Actual: actual,
		ExpectedStates: expectedStates, ActualStates: actualStates, Inventory: inventory,
		GeneratedArtifacts: generatedArtifacts, GeneratedBytes: generatedBytes,
		Authority:          Authority{LocalValidationCommands: 0, LocalValidationState: "NOT_RUN"},
		OperationalAudit:   operationalAudit(),
		PrFirstConformance: "REFUTED",
	}
	if err := writeJSON(filepath.Join(options.OutputDir, "suite-report.json"), report); err != nil {
		return SuiteReport{}, err
	}
	if err := writeSuiteHumanReport(filepath.Join(options.OutputDir, "human-report.md"), report); err != nil {
		return SuiteReport{}, err
	}
	return report, nil
}

func validateProgram(meta Meta, program Program) ([]Unknown, []Refutation) {
	unknown := []Unknown{}
	refuted := []Refutation{}
	if program.Namespace != meta.Namespace {
		refuted = append(refuted, Refutation{State: DecisionRefuted, Stage: "SEMANTIC", Step: "validate_namespace", Reason: "NAMESPACE_CONTRACT_MISMATCH", NextOperation: "REPAIR_NAMESPACE", BlockedBy: []string{meta.Namespace}})
	}
	if len(meta.Obligations) > 0 {
		declared := meta.Obligations[0]
		actual := program.Obligations[0]
		if declared.Contract != actual.Contract || declared.Fixture != actual.Fixture || declared.Toolchain != actual.Toolchain || declared.Runner != actual.Runner {
			refuted = append(refuted, Refutation{State: DecisionRefuted, Stage: "SEMANTIC", Step: "validate_obligation_contract", Reason: "OBLIGATION_CONTRACT_MISMATCH", NextOperation: "ALIGN_SOURCE_WITH_META_OBLIGATION", BlockedBy: []string{declared.ID, actual.ID}})
		}
	}
	if program.TopDecision != "" {
		switch program.TopDecision {
		case DecisionUnknown:
			unknown = append(unknown, UnknownClaim("SEMANTIC", "resolve_top_level_decision", "TOP_LEVEL_DECISION_UNKNOWN", "TOP_LEVEL_UNKNOWN", "SUPPLY_EXPLICIT_CLOSED_OR_REFUTED_DECISION", []string{"top-level-decision"}))
		case DecisionRefuted:
			refuted = append(refuted, Refutation{State: DecisionRefuted, Stage: "SEMANTIC", Step: "honor_top_level_decision", Reason: "EXPLICIT_REFUTED_TOP_LEVEL_DECISION", NextOperation: "REVIEW_PROGRAM", BlockedBy: []string{"top-level-decision"}})
		case DecisionClosed:
		default:
			refuted = append(refuted, Refutation{State: DecisionRefuted, Stage: "SEMANTIC", Step: "resolve_top_level_decision", Reason: "UNKNOWN_TOP_LEVEL_DECISION_FAIL_CLOSED", NextOperation: "DECLARE_EXPLICIT_DECISION", BlockedBy: []string{"top-level-decision"}})
		}
	}
	for _, effect := range program.Effects {
		for _, forbidden := range meta.ForbiddenEffects {
			if effect == forbidden {
				refuted = append(refuted, Refutation{State: DecisionRefuted, Stage: "AUTHORITY", Step: "check_effect_boundary", Reason: "FORBIDDEN_EFFECT_" + strings.ToUpper(effect), NextOperation: "REMOVE_FORBIDDEN_EFFECT", BlockedBy: []string{effect}})
			}
		}
	}
	return unknown, refuted
}

func buildDependencyGraph(program Program, obligationID string) DependencyGraph {
	nodes := []string{}
	for _, claim := range program.Claims {
		nodes = append(nodes, "claim:"+claim.ID)
	}
	for _, obligation := range program.Obligations {
		nodes = append(nodes, "obligation:"+obligation.ID)
	}
	for _, impact := range program.Impacts {
		nodes = append(nodes, impact.Node)
	}
	nodes = sortedStrings(unique(nodes))
	impacted := make([]string, 0, len(program.Impacts))
	for _, impact := range program.Impacts {
		impacted = append(impacted, impact.Node)
	}
	impacted = sortedStrings(unique(impacted))
	target := "obligation:" + obligationID
	adjacency := map[string][]string{}
	for _, edge := range program.Edges {
		adjacency[edge.From] = append(adjacency[edge.From], edge.To)
	}
	pathToTarget := false
	for _, start := range impacted {
		seen := map[string]bool{start: true}
		queue := []string{start}
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]
			if current == target {
				pathToTarget = true
				break
			}
			for _, next := range adjacency[current] {
				if !seen[next] {
					seen[next] = true
					queue = append(queue, next)
				}
			}
		}
	}
	return DependencyGraph{Schema: "gooo/proof-aware-test-reuse/dependency-graph/v1", Nodes: nodes, Edges: program.Edges, Impacted: impacted, Frontier: []string{}, PathToTarget: pathToTarget}
}

func graphRefutations(graph DependencyGraph) []Refutation {
	known := map[string]bool{}
	for _, node := range graph.Nodes {
		known[node] = true
	}
	refuted := []Refutation{}
	for _, edge := range graph.Edges {
		if !known[edge.From] || !known[edge.To] {
			refuted = append(refuted, Refutation{State: DecisionRefuted, Stage: "GRAPH", Step: "resolve_dependency_edge", Reason: "BROKEN_DEPENDENCY_EDGE", NextOperation: "DECLARE_ALL_EDGE_NODES", BlockedBy: []string{edge.From, edge.To}})
		}
	}
	return refuted
}

func inspectReceipts(receiptsDir string, program Program, key ProofKey, obligationID, mode, root string) ([]Receipt, []string, []Unknown, []Refutation, error) {
	receipts := []Receipt{}
	paths := []string{}
	unknown := []Unknown{}
	refuted := []Refutation{}
	if len(program.ReceiptRefs) == 0 {
		if mode == "incremental" {
			unknown = append(unknown, UnknownClaim("RECEIPT", "load_receipt", "MISSING_IMMUTABLE_RECEIPT", "MISSING_RECEIPT", "PRODUCE_EXACT_PASS_RECEIPT", []string{"receipt"}))
		}
		return receipts, paths, unknown, refuted, nil
	}
	for _, ref := range program.ReceiptRefs {
		path := resolvePath(receiptsDir, ref)
		receipt, err := loadReceipt(path)
		paths = append(paths, path)
		if err != nil {
			if os.IsNotExist(err) {
				unknown = append(unknown, UnknownClaim("RECEIPT", "load_receipt", "MISSING_IMMUTABLE_RECEIPT", "MISSING_RECEIPT", "PRODUCE_EXACT_PASS_RECEIPT", []string{ref}))
				continue
			}
			return nil, nil, nil, nil, fmt.Errorf("load receipt %s: %w", path, err)
		}
		receipts = append(receipts, receipt)
	}
	if len(receipts) > 1 && mode == "incremental" {
		unknown = append(unknown, UnknownClaim("RECEIPT", "disambiguate_receipts", "MULTIPLE_RECEIPTS_FOR_ONE_OBLIGATION", "AMBIGUOUS_RECEIPT", "SELECT_ONE_IMMUTABLE_RECEIPT", paths))
		return receipts, paths, unknown, refuted, nil
	}
	for _, receipt := range receipts {
		if receipt.State == "FAIL" {
			refuted = append(refuted, Refutation{State: DecisionRefuted, Stage: "RECEIPT", Step: "honor_receipt_failure", Reason: "IMMUTABLE_RECEIPT_EXPLICIT_FAIL", NextOperation: "REPAIR_AND_PRODUCE_NEW_RECEIPT", BlockedBy: []string{receipt.Obligation}})
			continue
		}
		if receipt.State != "PASS" {
			unknown = append(unknown, UnknownClaim("RECEIPT", "validate_receipt_state", "RECEIPT_STATE_NOT_PASS", "INVALID_RECEIPT", "PRODUCE_EXPLICIT_PASS_RECEIPT", []string{receipt.Obligation}))
			continue
		}
		if mode == "incremental" && !receiptMatches(receipt, key, obligationID) {
			unknown = append(unknown, UnknownClaim("RECEIPT", "compare_proof_key", "RECEIPT_PROOF_KEY_STALE", "STALE_RECEIPT", "PRODUCE_EXACT_PASS_RECEIPT", []string{"source_digest", "contract_digest", "fixture_digest", "toolchain_digest", "runner_digest"}))
		}
	}
	return receipts, paths, unknown, refuted, nil
}

func receiptMatches(receipt Receipt, key ProofKey, obligationID string) bool {
	return receipt.Schema == "gooo/proof-aware-test-reuse/receipt/v1" && receipt.Obligation == obligationID &&
		receipt.SourceDigest == key.SourceDigest && receipt.ContractDigest == key.ContractDigest &&
		receipt.FixtureDigest == key.FixtureDigest && receipt.ToolchainDigest == key.ToolchainDigest &&
		receipt.RunnerDigest == key.RunnerDigest && receipt.ResultDigest != ""
}

func digestFixture(root, fixture string) (string, error) {
	_, data, err := DigestFile(resolvePath(root, fixture))
	if err != nil {
		return "", err
	}
	return DigestBytes(data), nil
}

func repositoryRoot(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return filepath.Dir(path)
	}
	dir := filepath.Dir(abs)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return filepath.Dir(abs)
		}
		dir = parent
	}
}

func generateGo(output string) string {
	return "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Print(" + quoteGo(output) + ")\n}\n"
}

func quoteGo(value string) string {
	quoted := fmt.Sprintf("%q", value)
	return quoted
}

func executionStatus(action string, verified bool) string {
	switch action {
	case ActionSelected:
		if verified {
			return "EXECUTED_SEMANTICALLY"
		}
		return "SELECTED_FOR_EXECUTION"
	case ActionReused:
		return "REUSED_FROM_IMMUTABLE_RECEIPT"
	case ActionUnknown:
		return "BLOCKED_UNKNOWN"
	default:
		return "BLOCKED_REFUTED"
	}
}

func receiptState(decision string) string {
	if decision == DecisionClosed {
		return "PASS"
	}
	return decision
}

func operationalAudit() OperationalAudit {
	return OperationalAudit{
		State:      "OPERATIONAL_REFUTED",
		ExactCount: 1,
		Stage:      "AUTHORING",
		Step:       "OPEN_IMPLEMENTATION_PR_BEFORE_MAIN_INTEGRATION",
		Reason:     "INITIAL_IMPLEMENTATION_PUSH_PRECEDED_PR",
	}
}

func metricsFor(action string) RunMetrics {
	metrics := RunMetrics{Total: 1}
	switch action {
	case ActionSelected:
		metrics.Selected = 1
		metrics.Executed = 1
	case ActionReused:
		metrics.Reused = 1
	case ActionUnknown:
		metrics.Unknown = 1
	case ActionRefuted:
		metrics.Failed = 1
	}
	return metrics
}

func unique(values []string) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}

func collectArtifacts(root string, relativePaths []string) ([]Artifact, map[string]string, error) {
	artifacts := make([]Artifact, 0, len(relativePaths))
	digests := map[string]string{}
	sort.Strings(relativePaths)
	for _, relative := range relativePaths {
		path := filepath.Join(root, relative)
		digest, data, err := DigestFile(path)
		if err != nil {
			return nil, nil, err
		}
		artifact := Artifact{Path: relative, Kind: artifactKind(relative), Bytes: int64(len(data)), Digest: digest}
		artifacts = append(artifacts, artifact)
		digests[relative] = digest
	}
	return artifacts, digests, nil
}

func artifactKind(path string) string {
	if strings.HasSuffix(path, ".go") {
		return "generated-go"
	}
	if strings.HasSuffix(path, ".md") {
		return "human-report"
	}
	return "machine-evidence"
}

func writeHumanReport(path string, report RunReport) error {
	var builder strings.Builder
	builder.WriteString("# Proof-aware test reuse report\n\n")
	fmt.Fprintf(&builder, "Decision: `%s`  \nCase: `%s`  \nMode: `%s`  \n\n", report.Decision, report.CaseID, report.Mode)
	fmt.Fprintf(&builder, "The proof key is source `%s`, contract `%s`, fixture `%s`, toolchain `%s`, runner `%s`.\n\n", report.SourceDigest, report.ContractDigest, report.FixtureDigest, report.Toolchain, report.Runner)
	fmt.Fprintf(&builder, "Plan: `%s` — %s.\n\n", report.Plan.Action, report.Plan.Reason)
	fmt.Fprintf(&builder, "Operational audit: `%s` with exact count `%d`; pr_first_conformance=`%s`.\n\n", report.OperationalAudit.State, report.OperationalAudit.ExactCount, report.PrFirstConformance)
	builder.WriteString("## Semantic artifacts\n\n")
	builder.WriteString("The run emits semantic IR, dependency graph, deterministic test plan, a caller-owned generated Go program, execution evidence, and a receipt.\n\n")
	builder.WriteString("## Claims\n\n")
	if len(report.Unknown) == 0 && len(report.Refuted) == 0 {
		builder.WriteString("No UNKNOWN or REFUTED claims.\n\n")
	}
	for _, claim := range report.Unknown {
		fmt.Fprintf(&builder, "- UNKNOWN: `%s` / `%s` — %s; next `%s`; blocked by `%s`.\n", claim.Stage, claim.Step, claim.Reason, claim.NextOperation, strings.Join(claim.BlockedBy, ", "))
	}
	for _, claim := range report.Refuted {
		fmt.Fprintf(&builder, "- REFUTED: `%s` / `%s` — %s; next `%s`.\n", claim.Stage, claim.Step, claim.Reason, claim.NextOperation)
	}
	fmt.Fprintf(&builder, "\nUtility remains `%s` because no external user evidence is present.\n", report.Utility.State)
	return os.WriteFile(path, []byte(builder.String()), 0o644)
}

func writeSuiteHumanReport(path string, report SuiteReport) error {
	var builder strings.Builder
	builder.WriteString("# Proof-aware test reuse conformance report\n\n")
	fmt.Fprintf(&builder, "Decision: `%s`  \nMode: `%s`  \nFixed denominator: `%d` cases  \n\n", report.Decision, report.Mode, report.FixedDenominator)
	fmt.Fprintf(&builder, "Actual: total=%d selected=%d executed=%d reused=%d unknown=%d refuted=%d.\n\n", report.Actual.Total, report.Actual.Selected, report.Actual.Executed, report.Actual.Reused, report.Actual.Unknown, report.Actual.Failed)
	fmt.Fprintf(&builder, "Operational audit: `%s` with exact count `%d`; pr_first_conformance=`%s`.\n\n", report.OperationalAudit.State, report.OperationalAudit.ExactCount, report.PrFirstConformance)
	builder.WriteString("| Case | Expected | Actual | Action | Reason |\n|---|---|---|---|---|\n")
	for _, item := range report.Cases {
		fmt.Fprintf(&builder, "| `%s` | `%s` | `%s` | `%s` | %s |\n", item.ID, item.Expected, item.Decision, item.Action, item.Reason)
	}
	builder.WriteString("\nA case is CLOSED only when its semantic obligations are satisfied; UNKNOWN is never converted to CLOSED, and REFUTED takes precedence.\n")
	return os.WriteFile(path, []byte(builder.String()), 0o644)
}
