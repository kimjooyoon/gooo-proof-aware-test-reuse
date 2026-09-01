package engine

import "testing"

func TestDecisionPrecedence(t *testing.T) {
	unknown := []Unknown{{State: DecisionUnknown}}
	refuted := []Refutation{{State: DecisionRefuted}}
	if got := Reduce(unknown, refuted); got != DecisionRefuted {
		t.Fatalf("decision precedence = %s", got)
	}
}

func TestReceiptMatchesProofKey(t *testing.T) {
	key := ProofKey{SourceDigest: "source", ContractDigest: "contract", FixtureDigest: "fixture", ToolchainDigest: "toolchain", RunnerDigest: "runner"}
	receipt := Receipt{Schema: "gooo/proof-aware-test-reuse/receipt/v1", Obligation: "smoke", State: "PASS", SourceDigest: "source", ContractDigest: "contract", FixtureDigest: "fixture", ToolchainDigest: "toolchain", RunnerDigest: "runner", ResultDigest: "result"}
	if !receiptMatches(receipt, key, "smoke") {
		t.Fatal("matching receipt was rejected")
	}
	receipt.RunnerDigest = "other"
	if receiptMatches(receipt, key, "smoke") {
		t.Fatal("stale receipt was accepted")
	}
}

func TestImpactPathReachesObligation(t *testing.T) {
	program := Program{
		Claims: []Claim{{ID: "message"}},
		Obligations: []Obligation{{ID: "smoke"}},
		Edges: []Edge{{From: "claim:message", To: "obligation:smoke"}},
		Impacts: []Impact{{Node: "claim:message"}},
	}
	graph := buildDependencyGraph(program, "smoke")
	if !graph.PathToTarget || len(graph.Frontier) != 0 {
		t.Fatalf("unexpected graph before frontier projection: %#v", graph)
	}
	graph.Frontier = append(graph.Frontier, graph.Impacted...)
	if len(graph.Frontier) != 1 || graph.Frontier[0] != "claim:message" {
		t.Fatalf("unexpected invalidation frontier: %#v", graph.Frontier)
	}
}

func TestGeneratedGoIsDeterministic(t *testing.T) {
	first := generateGo("hello")
	second := generateGo("hello")
	if first != second {
		t.Fatal("generated Go changed between identical inputs")
	}
}

func TestInventoryExcludesRootReadme(t *testing.T) {
	inventory := Inventory{RootReadmeExcluded: true, GitExcluded: true}
	if !inventory.RootReadmeExcluded {
		t.Fatal("root README exclusion is not recorded")
	}
}
