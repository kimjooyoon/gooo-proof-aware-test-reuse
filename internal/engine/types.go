package engine

const (
	DecisionClosed  = "CLOSED"
	DecisionUnknown = "UNKNOWN"
	DecisionRefuted = "REFUTED"

	ActionSelected = "SELECTED"
	ActionReused   = "REUSED"
	ActionUnknown  = "UNKNOWN"
	ActionRefuted  = "REFUTED"

	Toolchain = "go1.27.0"
	Runner    = "ubuntu-latest"
)

type Claim struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type Obligation struct {
	ID        string `json:"id"`
	Contract  string `json:"contract"`
	Fixture   string `json:"fixture"`
	Toolchain string `json:"toolchain"`
	Runner    string `json:"runner"`
}

type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type IndicatorSpec struct {
	Name      string `json:"name"`
	Direction string `json:"direction"`
	Goal      string `json:"goal"`
}

type IndicatorPolicy struct {
	DeltaDefinition   string `json:"delta_definition"`
	ImprovementPolicy string `json:"improvement_policy"`
	UnchangedClaim    string `json:"unchanged_claim"`
	RegressionClaim   string `json:"regression_claim"`
}

type IndicatorObservation struct {
	Name        string `json:"name"`
	Before      int64  `json:"before"`
	After       int64  `json:"after"`
	SignedDelta int64  `json:"signed_delta"`
	Direction   string `json:"direction"`
	Observation string `json:"observation"`
	ClaimState  string `json:"claim_state"`
	Reason      string `json:"reason"`
}

type Impact struct {
	Node string `json:"node"`
}

type Meta struct {
	Schema           string          `json:"schema"`
	Program          string          `json:"program"`
	Namespace        string          `json:"namespace"`
	Precedence       []string        `json:"precedence"`
	ReceiptSchema    string          `json:"receipt_schema"`
	ReceiptPolicy    string          `json:"receipt_policy"`
	ReceiptStates    []string        `json:"receipt_states"`
	ReceiptFields    []string        `json:"receipt_fields"`
	Claims           []Claim         `json:"claims"`
	Obligations      []Obligation    `json:"obligations"`
	Edges            []Edge          `json:"edges"`
	ImpactPolicy     string          `json:"impact_policy"`
	Activities       []string        `json:"activities"`
	ForbiddenEffects []string        `json:"forbidden_effects"`
	Indicators       []IndicatorSpec `json:"indicators"`
	IndicatorPolicy  IndicatorPolicy `json:"indicator_policy"`
	SourcePath       string          `json:"source_path"`
	SourceDigest     string          `json:"source_digest"`
}

type Program struct {
	Schema       string       `json:"schema"`
	CaseID       string       `json:"case_id"`
	Program      string       `json:"program"`
	Namespace    string       `json:"namespace"`
	Claims       []Claim      `json:"claims"`
	Obligations  []Obligation `json:"obligations"`
	Edges        []Edge       `json:"edges"`
	Impacts      []Impact     `json:"impacts"`
	ReceiptRefs  []string     `json:"receipt_refs"`
	Output       string       `json:"output"`
	Effects      []string     `json:"effects"`
	TopDecision  string       `json:"top_decision,omitempty"`
	SourcePath   string       `json:"source_path"`
	SourceDigest string       `json:"source_digest"`
}

type Receipt struct {
	Schema             string           `json:"schema"`
	Obligation         string           `json:"obligation"`
	State              string           `json:"state"`
	SourceDigest       string           `json:"source_digest"`
	ContractDigest     string           `json:"contract_digest"`
	FixtureDigest      string           `json:"fixture_digest"`
	ToolchainDigest    string           `json:"toolchain_digest"`
	RunnerDigest       string           `json:"runner_digest"`
	ResultDigest       string           `json:"result_digest"`
	OperationalAudit   OperationalAudit `json:"operational_audit"`
	PrFirstConformance string           `json:"pr_first_conformance"`
}

type OperationalAudit struct {
	State      string `json:"state"`
	ExactCount int    `json:"exact_count"`
	Stage      string `json:"stage"`
	Step       string `json:"step"`
	Reason     string `json:"reason"`
}

type ProofKey struct {
	SourceDigest    string `json:"source_digest"`
	ContractDigest  string `json:"contract_digest"`
	FixtureDigest   string `json:"fixture_digest"`
	ToolchainDigest string `json:"toolchain_digest"`
	RunnerDigest    string `json:"runner_digest"`
}

type Unknown struct {
	State         string   `json:"state"`
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

type Refutation struct {
	State         string   `json:"state"`
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

type DependencyGraph struct {
	Schema       string   `json:"schema"`
	Nodes        []string `json:"nodes"`
	Edges        []Edge   `json:"edges"`
	Impacted     []string `json:"impacted"`
	Frontier     []string `json:"invalidation_frontier"`
	PathToTarget bool     `json:"path_to_target"`
}

type Plan struct {
	Schema       string   `json:"schema"`
	Mode         string   `json:"mode"`
	Obligation   string   `json:"obligation"`
	Action       string   `json:"action"`
	Reason       string   `json:"reason"`
	Selected     []string `json:"selected"`
	Reused       []string `json:"reused"`
	Unknown      []string `json:"unknown"`
	Refuted      []string `json:"refuted"`
	ReceiptPaths []string `json:"receipt_paths"`
}

type Verification struct {
	Schema               string `json:"schema"`
	GeneratedOutput      string `json:"generated_output"`
	ExpectedOutput       string `json:"expected_output"`
	ExpectedResultDigest string `json:"expected_result_digest"`
	ExecutionRequired    bool   `json:"execution_required"`
	ExecutionVerified    bool   `json:"execution_verified"`
	ReusedProof          bool   `json:"reused_proof"`
}

type Inventory struct {
	RootReadmeExcluded   bool `json:"root_readme_excluded"`
	GitExcluded          bool `json:"git_excluded"`
	CallerOutputExcluded bool `json:"caller_output_excluded"`
	CacheExcluded        bool `json:"cache_excluded"`
	VendorExcluded       bool `json:"vendor_excluded"`
	ToolchainExcluded    bool `json:"toolchain_excluded"`
	DescendantDirs       int  `json:"descendant_dirs"`
	RegularFiles         int  `json:"regular_files"`
	GoFiles              int  `json:"go_files"`
	GoPhysicalLines      int  `json:"go_physical_lines"`
	GoooFiles            int  `json:"gooo_files"`
	GoooPhysicalLines    int  `json:"gooo_physical_lines"`
	PhysicalLines        int  `json:"physical_lines"`
}

type Artifact struct {
	Path   string `json:"path"`
	Kind   string `json:"kind"`
	Bytes  int64  `json:"bytes"`
	Digest string `json:"digest"`
}

type Authority struct {
	RepositoryWrites        int    `json:"repository_writes"`
	CommitAuthority         int    `json:"commit_authority"`
	PushAuthority           int    `json:"push_authority"`
	MergeAuthority          int    `json:"merge_authority"`
	ReleaseMutation         int    `json:"release_mutation"`
	LocalValidationCommands int    `json:"local_validation_commands"`
	LocalValidationState    string `json:"local_validation_state"`
}

type RunMetrics struct {
	Total    int `json:"total"`
	Selected int `json:"selected"`
	Executed int `json:"executed"`
	Reused   int `json:"reused"`
	Failed   int `json:"failed"`
	Unknown  int `json:"unknown"`
}

type RunReport struct {
	Schema             string            `json:"schema"`
	Decision           string            `json:"decision"`
	Mode               string            `json:"mode"`
	CaseID             string            `json:"case_id"`
	SourceDigest       string            `json:"source_digest"`
	ContractDigest     string            `json:"contract_digest"`
	FixtureDigest      string            `json:"fixture_digest"`
	Toolchain          string            `json:"toolchain"`
	Runner             string            `json:"runner"`
	ProofKey           ProofKey          `json:"proof_key"`
	Plan               Plan              `json:"plan"`
	Unknown            []Unknown         `json:"unknown"`
	Refuted            []Refutation      `json:"refuted"`
	IndicatorSpecs     []IndicatorSpec   `json:"indicator_specs"`
	IndicatorPolicy    IndicatorPolicy   `json:"indicator_policy"`
	Utility            Unknown           `json:"utility"`
	Metrics            RunMetrics        `json:"metrics"`
	Inventory          Inventory         `json:"inventory"`
	GeneratedArtifacts []Artifact        `json:"generated_artifacts"`
	Verification       Verification      `json:"verification"`
	Authority          Authority         `json:"authority"`
	ArtifactDigests    map[string]string `json:"artifact_digests"`
	OperationalAudit   OperationalAudit  `json:"operational_audit"`
	PrFirstConformance string            `json:"pr_first_conformance"`
}

type ContractCase struct {
	ID       string `json:"id"`
	Source   string `json:"source"`
	Expected string `json:"expected"`
	Kind     string `json:"kind"`
}

type Contract struct {
	Schema     string         `json:"schema"`
	ID         string         `json:"id"`
	Version    string         `json:"version"`
	Fixed      bool           `json:"fixed"`
	Precedence []string       `json:"precedence"`
	Cases      []ContractCase `json:"cases"`
}

type SuiteCase struct {
	ID         string       `json:"id"`
	Kind       string       `json:"kind"`
	Expected   string       `json:"expected"`
	Decision   string       `json:"decision"`
	Action     string       `json:"action"`
	Match      bool         `json:"match"`
	Reason     string       `json:"reason"`
	Unknown    []Unknown    `json:"unknown"`
	Refuted    []Refutation `json:"refuted"`
	ReportPath string       `json:"report_path"`
}

type SuiteReport struct {
	Schema             string           `json:"schema"`
	Decision           string           `json:"decision"`
	Contract           string           `json:"contract"`
	ContractDigest     string           `json:"contract_digest"`
	Mode               string           `json:"mode"`
	FixedDenominator   int              `json:"fixed_denominator"`
	Cases              []SuiteCase      `json:"cases"`
	Actual             RunMetrics       `json:"actual"`
	ExpectedStates     map[string]int   `json:"expected_states"`
	ActualStates       map[string]int   `json:"actual_states"`
	Inventory          Inventory        `json:"inventory"`
	GeneratedArtifacts int              `json:"generated_artifacts"`
	GeneratedBytes     int64            `json:"generated_bytes"`
	Authority          Authority        `json:"authority"`
	OperationalAudit   OperationalAudit `json:"operational_audit"`
	PrFirstConformance string           `json:"pr_first_conformance"`
}

func UnknownClaim(stage, step, reason, class, next string, blockedBy []string) Unknown {
	return Unknown{State: DecisionUnknown, Stage: stage, Step: step, Reason: reason, UnknownClass: class, NextOperation: next, BlockedBy: blockedBy}
}

func Reduce(unknown []Unknown, refuted []Refutation) string {
	if len(refuted) > 0 {
		return DecisionRefuted
	}
	if len(unknown) > 0 {
		return DecisionUnknown
	}
	return DecisionClosed
}
