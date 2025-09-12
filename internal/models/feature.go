package models

// ValidAgents are provided by models.ListAgents()/IsValidAgent

// FeatureCreateResult represents the result of creating a new feature
type FeatureCreateResult struct {
	BranchName string `json:"branch_name"`
	SpecFile   string `json:"spec_file"`
	FeatureNum string `json:"feature_num"`
}

// FeaturePlanResult represents the result of setting up a feature plan
type FeaturePlanResult struct {
	FeatureSpec string `json:"feature_spec"`
	ImplPlan    string `json:"impl_plan"`
	SpecsDir    string `json:"specs_dir"`
	Branch      string `json:"branch"`
}

// FeatureCheckResult represents the result of checking feature prerequisites
type FeatureCheckResult struct {
	FeatureDir    string   `json:"feature_dir"`
	AvailableDocs []string `json:"available_docs"`
}

// ContextUpdate represents an update to an agent context file
type ContextUpdate struct {
	Agent string `json:"agent"`
}

// FeatureContextResult represents the result of updating agent context files
type FeatureContextResult struct {
	Branch  string          `json:"branch"`
	Updates []ContextUpdate `json:"updates"`
	Summary []string        `json:"summary"`
}

// FeaturePathsResult represents the paths for a feature branch
type FeaturePathsResult struct {
	RepoRoot    string `json:"repo_root"`
	Branch      string `json:"branch"`
	FeatureDir  string `json:"feature_dir"`
	FeatureSpec string `json:"feature_spec"`
	ImplPlan    string `json:"impl_plan"`
	Tasks       string `json:"tasks"`
}
