package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const modulePath = "github.com/117503445/ai-coding-webui"

type buildInfo struct {
	BuildTime  string
	GitBranch  string
	GitCommit  string
	GitTag     string
	GitDirty   string
	GitVersion string
	BuildDir   string
}

func mustProjectRoot() string {
	out, err := commandOutput("", "git", "rev-parse", "--show-toplevel")
	if err != nil {
		panic(fmt.Errorf("failed to detect git root: %w", err))
	}
	return strings.TrimSpace(out)
}

func collectBuildInfo() buildInfo {
	gitCommit := gitValue("rev-parse", "HEAD")
	gitBranch := gitValue("rev-parse", "--abbrev-ref", "HEAD")
	gitTag := gitValue("describe", "--tags", "--exact-match")
	gitVersion := gitTag
	if gitVersion == "" {
		gitVersion = shortCommit(gitCommit)
	}

	return buildInfo{
		BuildTime:  time.Now().UTC().Format("20060102T150405Z"),
		GitBranch:  gitBranch,
		GitCommit:  gitCommit,
		GitTag:     gitTag,
		GitDirty:   fmt.Sprintf("%t", isGitDirty()),
		GitVersion: gitVersion,
		BuildDir:   dirProjectRoot,
	}
}

func buildLDFlags(info buildInfo) string {
	return strings.Join([]string{
		"-s",
		"-w",
		fmt.Sprintf("-X %s/internal/buildinfo.BuildTime=%s", modulePath, info.BuildTime),
		fmt.Sprintf("-X %s/internal/buildinfo.GitBranch=%s", modulePath, info.GitBranch),
		fmt.Sprintf("-X %s/internal/buildinfo.GitCommit=%s", modulePath, info.GitCommit),
		fmt.Sprintf("-X %s/internal/buildinfo.GitTag=%s", modulePath, info.GitTag),
		fmt.Sprintf("-X %s/internal/buildinfo.GitDirty=%s", modulePath, info.GitDirty),
		fmt.Sprintf("-X %s/internal/buildinfo.GitVersion=%s", modulePath, info.GitVersion),
		fmt.Sprintf("-X %s/internal/buildinfo.BuildDir=%s", modulePath, info.BuildDir),
	}, " ")
}

func gitValue(args ...string) string {
	out, err := commandOutput(dirProjectRoot, "git", args...)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

func isGitDirty() bool {
	out, err := commandOutput(dirProjectRoot, "git", "status", "--porcelain")
	if err != nil {
		return false
	}
	return strings.TrimSpace(out) != ""
}

func shortCommit(commit string) string {
	if len(commit) > 7 {
		return commit[:7]
	}
	return commit
}

func commandOutput(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func runCommand(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	return cmd.Run()
}
