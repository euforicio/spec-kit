package services

import (
	"fmt"
	"os/exec"
	"strings"
)

type GitService struct{}

type GitServiceInterface interface {
	GetRepoRoot() (string, error)
	GetCurrentBranch() (string, error)
	CreateBranch(branchName string) error
	CheckoutBranch(branchName string) error
	BranchExists(branchName string) (bool, error)
}

func NewGitService() *GitService {
	return &GitService{}
}

func (g *GitService) GetRepoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get repository root: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func (g *GitService) GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func (g *GitService) CreateBranch(branchName string) error {
	cmd := exec.Command("git", "checkout", "-b", branchName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create branch '%s': %w", branchName, err)
	}
	return nil
}

func (g *GitService) CheckoutBranch(branchName string) error {
	cmd := exec.Command("git", "checkout", branchName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to checkout branch '%s': %w", branchName, err)
	}
	return nil
}

func (g *GitService) BranchExists(branchName string) (bool, error) {
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branchName)
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if branch exists: %w", err)
	}
	return true, nil
}