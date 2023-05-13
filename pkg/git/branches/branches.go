package branches

import (
	"bufio"
	"context"
	"fmt"
	"github.com/adlternative/tinygithub/pkg/cmd"
	"strings"
)

func GetAllBranch(ctx context.Context, repoPath string) ([]string, error) {
	var stderrBuf strings.Builder
	var branches []string

	gitCmd := cmd.NewGitCommand("branch").WithGitDir(repoPath).
		WithOptions("--format=%(refname:lstrip=2)").
		WithStderr(&stderrBuf)

	if err := gitCmd.Start(ctx); err != nil {
		return nil, fmt.Errorf("gitCmd start failed with %w", err)
	}

	scanner := bufio.NewScanner(gitCmd)

	for scanner.Scan() {
		branches = append(branches, scanner.Text())
	}

	if err := gitCmd.Wait(); err != nil {
		return nil, fmt.Errorf("git command failed with stderr:%v, error:%w", stderrBuf.String(), err)
	}
	return branches, nil
}
