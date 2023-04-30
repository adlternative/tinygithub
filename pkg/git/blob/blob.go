package blob

import (
	"context"
	"fmt"
	"github.com/adlternative/tinygithub/pkg/cmd"
	"go.uber.org/zap/buffer"
	"strings"
)

func ShowBlob(ctx context.Context, repoPath, revision string) ([]byte, error) {
	var stderrBuf strings.Builder
	var stdoutBuf buffer.Buffer

	gitCmd := cmd.NewGitCommand("cat-file").WithGitDir(repoPath).
		WithOptions("-p").
		WithArgs(revision).
		WithStderr(&stderrBuf).
		WithStdout(&stdoutBuf)

	if err := gitCmd.Start(ctx); err != nil {
		return nil, fmt.Errorf("gitCmd start failed with %w", err)
	}

	if err := gitCmd.Wait(); err != nil {
		return nil, fmt.Errorf("git command failed with stderr:%v, error:%w", stderrBuf.String(), err)
	}
	return stdoutBuf.Bytes(), nil
}
