package codex

// Service provides Codex integration functionality
type Service struct {
	generator *AgentsGenerator
	creator   *CommandCreator
}

// NewService creates a new Codex service
func NewService() *Service {
	return &Service{
		generator: NewAgentsGenerator(),
		creator:   NewCommandCreator(),
	}
}

// GenerateAGENTS generates AGENTS.md content from plan content
func (s *Service) GenerateAGENTS(planContent string, existingContent []byte) (string, error) {
	return s.generator.Generate(planContent, existingContent)
}

// WriteAGENTS writes AGENTS.md content to the project root
func (s *Service) WriteAGENTS(content string, repoRoot string) error {
	writer := NewWriter(repoRoot)
	return writer.WriteAGENTSmd(content)
}

// WriteCommandFiles writes command files to .codex/commands/
func (s *Service) WriteCommandFiles(force bool, repoRoot string) error {
	// Create commands using the CommandCreator
	var createFunc func(string, CommandSet) error
	if force {
		createFunc = s.creator.CreateCommandsForce
	} else {
		createFunc = s.creator.CreateCommands
	}

	return createFunc(repoRoot, CommandSetStandard)
}