package engine

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func ParseMeta(path string) (Meta, error) {
	digest, data, err := DigestFile(path)
	if err != nil {
		return Meta{}, err
	}
	meta := Meta{
		Schema: "gooo/proof-aware-test-reuse/meta/v1",
		Precedence: []string{}, ReceiptStates: []string{}, ReceiptFields: []string{},
		Claims: []Claim{}, Obligations: []Obligation{}, Edges: []Edge{}, Activities: []string{},
		ForbiddenEffects: []string{}, SourcePath: path, SourceDigest: digest,
	}
	if err := scanLines(data, func(line string, lineNumber int) error {
		switch {
		case strings.HasPrefix(line, "program "):
			meta.Program = strings.TrimSpace(strings.TrimPrefix(line, "program "))
		case strings.HasPrefix(line, "namespace "):
			meta.Namespace = strings.TrimSpace(strings.TrimPrefix(line, "namespace "))
		case strings.HasPrefix(line, "precedence "):
			meta.Precedence = strings.Fields(strings.ReplaceAll(strings.TrimSpace(strings.TrimPrefix(line, "precedence ")), ">", " "))
		case strings.HasPrefix(line, "receipt_schema "):
			meta.ReceiptSchema = strings.TrimSpace(strings.TrimPrefix(line, "receipt_schema "))
		case strings.HasPrefix(line, "receipt_policy "):
			meta.ReceiptPolicy = strings.TrimSpace(strings.TrimPrefix(line, "receipt_policy "))
		case strings.HasPrefix(line, "receipt_state "):
			meta.ReceiptStates = append(meta.ReceiptStates, strings.TrimSpace(strings.TrimPrefix(line, "receipt_state ")))
		case strings.HasPrefix(line, "receipt_field "):
			meta.ReceiptFields = append(meta.ReceiptFields, strings.TrimSpace(strings.TrimPrefix(line, "receipt_field ")))
		case strings.HasPrefix(line, "claim "):
			claim, ok := parseClaim(strings.TrimSpace(strings.TrimPrefix(line, "claim ")))
			if !ok {
				return fmt.Errorf("line %d: invalid claim", lineNumber)
			}
			meta.Claims = append(meta.Claims, claim)
		case strings.HasPrefix(line, "obligation "):
			obligation, ok := parseObligation(strings.TrimSpace(strings.TrimPrefix(line, "obligation ")))
			if !ok {
				return fmt.Errorf("line %d: invalid obligation", lineNumber)
			}
			meta.Obligations = append(meta.Obligations, obligation)
		case strings.HasPrefix(line, "edge "):
			edge, ok := parseEdge(strings.TrimSpace(strings.TrimPrefix(line, "edge ")))
			if !ok {
				return fmt.Errorf("line %d: invalid edge", lineNumber)
			}
			meta.Edges = append(meta.Edges, edge)
		case strings.HasPrefix(line, "impact_policy "):
			meta.ImpactPolicy = strings.TrimSpace(strings.TrimPrefix(line, "impact_policy "))
		case strings.HasPrefix(line, "activity "):
			meta.Activities = append(meta.Activities, strings.TrimSpace(strings.TrimPrefix(line, "activity ")))
		case strings.HasPrefix(line, "forbid_effect "):
			meta.ForbiddenEffects = append(meta.ForbiddenEffects, strings.TrimSpace(strings.TrimPrefix(line, "forbid_effect ")))
		}
		return nil
	}); err != nil {
		return Meta{}, err
	}
	if err := validateMeta(meta); err != nil {
		return Meta{}, err
	}
	return meta, nil
}

func ParseProgram(path string) (Program, error) {
	digest, data, err := DigestFile(path)
	if err != nil {
		return Program{}, err
	}
	program := Program{
		Schema: "gooo/proof-aware-test-reuse/program/v1",
		Claims: []Claim{}, Obligations: []Obligation{}, Edges: []Edge{}, Impacts: []Impact{},
		ReceiptRefs: []string{}, Effects: []string{}, SourcePath: path, SourceDigest: digest,
	}
	if err := scanLines(data, func(line string, lineNumber int) error {
		switch {
		case strings.HasPrefix(line, "case "):
			program.CaseID = strings.TrimSpace(strings.TrimPrefix(line, "case "))
		case strings.HasPrefix(line, "program "):
			program.Program = strings.TrimSpace(strings.TrimPrefix(line, "program "))
		case strings.HasPrefix(line, "namespace "):
			program.Namespace = strings.TrimSpace(strings.TrimPrefix(line, "namespace "))
		case strings.HasPrefix(line, "claim "):
			claim, ok := parseClaim(strings.TrimSpace(strings.TrimPrefix(line, "claim ")))
			if !ok {
				return fmt.Errorf("line %d: invalid claim", lineNumber)
			}
			program.Claims = append(program.Claims, claim)
		case strings.HasPrefix(line, "obligation "):
			obligation, ok := parseObligation(strings.TrimSpace(strings.TrimPrefix(line, "obligation ")))
			if !ok {
				return fmt.Errorf("line %d: invalid obligation", lineNumber)
			}
			program.Obligations = append(program.Obligations, obligation)
		case strings.HasPrefix(line, "edge "):
			edge, ok := parseEdge(strings.TrimSpace(strings.TrimPrefix(line, "edge ")))
			if !ok {
				return fmt.Errorf("line %d: invalid edge", lineNumber)
			}
			program.Edges = append(program.Edges, edge)
		case strings.HasPrefix(line, "impact "):
			node := strings.TrimSpace(strings.TrimPrefix(line, "impact "))
			if node == "" {
				return fmt.Errorf("line %d: empty impact", lineNumber)
			}
			program.Impacts = append(program.Impacts, Impact{Node: node})
		case strings.HasPrefix(line, "receipt "):
			ref := strings.TrimSpace(strings.TrimPrefix(line, "receipt "))
			if ref == "" {
				return fmt.Errorf("line %d: empty receipt", lineNumber)
			}
			program.ReceiptRefs = append(program.ReceiptRefs, ref)
		case strings.HasPrefix(line, "output "):
			value, ok := parseQuoted(strings.TrimSpace(strings.TrimPrefix(line, "output ")))
			if !ok {
				return fmt.Errorf("line %d: invalid output", lineNumber)
			}
			program.Output = value
		case strings.HasPrefix(line, "effect "):
			program.Effects = append(program.Effects, strings.TrimSpace(strings.TrimPrefix(line, "effect ")))
		case strings.HasPrefix(line, "decision "):
			program.TopDecision = strings.TrimSpace(strings.TrimPrefix(line, "decision "))
		}
		return nil
	}); err != nil {
		return Program{}, err
	}
	if program.CaseID == "" || program.Program == "" || program.Namespace == "" || len(program.Obligations) != 1 || program.Output == "" {
		return Program{}, fmt.Errorf("program requires case, program, namespace, exactly one obligation, and output")
	}
	return program, nil
}

func parseClaim(value string) (Claim, bool) {
	quote := strings.IndexByte(value, '"')
	if quote < 1 {
		return Claim{}, false
	}
	id := strings.TrimSpace(value[:quote])
	text, err := strconv.Unquote(strings.TrimSpace(value[quote:]))
	if err != nil || id == "" || text == "" {
		return Claim{}, false
	}
	return Claim{ID: id, Text: text}, true
}

func parseObligation(value string) (Obligation, bool) {
	fields := strings.Fields(value)
	if len(fields) != 5 {
		return Obligation{}, false
	}
	return Obligation{ID: fields[0], Contract: fields[1], Fixture: fields[2], Toolchain: fields[3], Runner: fields[4]}, true
}

func parseEdge(value string) (Edge, bool) {
	parts := strings.Split(value, "->")
	if len(parts) != 2 {
		return Edge{}, false
	}
	from := strings.TrimSpace(parts[0])
	to := strings.TrimSpace(parts[1])
	return Edge{From: from, To: to}, from != "" && to != ""
}

func parseQuoted(value string) (string, bool) {
	if !strings.HasPrefix(value, "\"") {
		return "", false
	}
	result, err := strconv.Unquote(value)
	return result, err == nil && result != ""
}

func scanLines(data []byte, visit func(string, int) error) error {
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if index := strings.IndexByte(line, '#'); index >= 0 {
			line = strings.TrimSpace(line[:index])
		}
		if line == "" {
			continue
		}
		if err := visit(line, lineNumber); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func validateMeta(meta Meta) error {
	if meta.Program == "" || meta.Namespace == "" || meta.ReceiptSchema == "" || meta.ReceiptPolicy == "" || meta.ImpactPolicy == "" {
		return fmt.Errorf("meta requires program, namespace, receipt schema/policy, and impact policy")
	}
	if len(meta.Claims) == 0 || len(meta.Obligations) == 0 || len(meta.Edges) == 0 || len(meta.Activities) == 0 {
		return fmt.Errorf("meta must declare claims, obligations, dependency edges, and activities")
	}
	if strings.Join(meta.Precedence, ">") != "REFUTED>UNKNOWN>CLOSED" {
		return fmt.Errorf("meta precedence must be REFUTED > UNKNOWN > CLOSED")
	}
	if meta.ReceiptPolicy != "immutable-pass-only" {
		return fmt.Errorf("unsupported receipt policy %q", meta.ReceiptPolicy)
	}
	return nil
}

func loadContract(path string) (Contract, string, error) {
	digest, _, err := DigestFile(path)
	if err != nil {
		return Contract{}, "", err
	}
	var contract Contract
	if err := readJSON(path, &contract); err != nil {
		return Contract{}, "", err
	}
	if err := validateContract(contract); err != nil {
		return Contract{}, "", err
	}
	return contract, digest, nil
}

func validateContract(contract Contract) error {
	if contract.Schema != "gooo/proof-aware-test-reuse/denominator/v1" || !contract.Fixed || len(contract.Cases) != 9 {
		return fmt.Errorf("contract must be the fixed nine-case denominator")
	}
	if strings.Join(contract.Precedence, ">") != "REFUTED>UNKNOWN>CLOSED" {
		return fmt.Errorf("contract precedence is not fixed")
	}
	seen := map[string]bool{}
	counts := map[string]int{}
	for _, item := range contract.Cases {
		if item.ID == "" || item.Source == "" || item.Expected == "" || item.Kind == "" || seen[item.ID] {
			return fmt.Errorf("invalid or duplicate contract case %q", item.ID)
		}
		if item.Expected != DecisionClosed && item.Expected != DecisionUnknown && item.Expected != DecisionRefuted {
			return fmt.Errorf("invalid expected decision for %s", item.ID)
		}
		seen[item.ID] = true
		counts[item.Expected]++
	}
	if counts[DecisionClosed] != 3 || counts[DecisionUnknown] != 3 || counts[DecisionRefuted] != 3 {
		return fmt.Errorf("contract requires 3 CLOSED, 3 UNKNOWN, and 3 REFUTED cases")
	}
	return nil
}

func loadReceipt(path string) (Receipt, error) {
	var receipt Receipt
	if err := readJSON(path, &receipt); err != nil {
		return Receipt{}, err
	}
	return receipt, nil
}

func loadReceipts(root string, refs []string) ([]Receipt, []string, error) {
	receipts := make([]Receipt, 0, len(refs))
	paths := make([]string, 0, len(refs))
	for _, ref := range refs {
		path := resolvePath(root, ref)
		receipt, err := loadReceipt(path)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, []string{path}, nil
			}
			return nil, nil, err
		}
		receipts = append(receipts, receipt)
		paths = append(paths, path)
	}
	return receipts, paths, nil
}
