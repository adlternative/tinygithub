package gitRepo

import (
	"context"
	"fmt"
	"github.com/adlternative/tinygithub/pkg/cmd"
	"strings"
)

func IsRepositoryEmpty(ctx context.Context, repoPath string) (bool, error) {
	var stderrBuf strings.Builder
	var stdoutBuf strings.Builder

	gitCmd := cmd.NewGitCommand("rev-parse").WithGitDir(repoPath).
		WithArgs("HEAD").
		WithStdout(&stdoutBuf).
		WithStderr(&stderrBuf)

	if err := gitCmd.Run(ctx); err != nil {
		return true, fmt.Errorf("gitCmd run failed with %w", err)
	}
	return stdoutBuf.String() == "HEAD", nil
}
