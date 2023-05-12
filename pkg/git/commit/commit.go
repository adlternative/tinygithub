package commit

import (
	"bytes"
	"context"
	"fmt"
	"github.com/adlternative/tinygithub/pkg/cmd"
	"github.com/adlternative/tinygithub/pkg/git/object"
	"go.uber.org/zap/buffer"
	"strings"
)

type Object struct {
	ID *object.ID `json:"id"`

	RelativeDate string `json:"date"`
	Header       string `json:"header"`
	Message      string `json:"message"`
}

func ParseCommit(ctx context.Context, repoPath string, commitID *object.ID) (*Object, error) {
	var stderrBuf strings.Builder
	var stdoutBuf buffer.Buffer

	gitCmd := cmd.NewGitCommand("log").WithGitDir(repoPath).
		WithOptions("--max-count=1", "--date=relative", "--format=%cd%x00%s%x00%B").
		WithArgs(commitID.String()).
		WithStderr(&stderrBuf).WithStdout(&stdoutBuf)

	if err := gitCmd.Start(ctx); err != nil {
		return nil, fmt.Errorf("git log start failed with %w", err)
	}

	if err := gitCmd.Wait(); err != nil {
		return nil, fmt.Errorf("git log failed with stderr:%v, error:%w", stderrBuf.String(), err)
	}

	parts := bytes.Split(stdoutBuf.Bytes(), []byte{0})
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid git log output format: %v", stdoutBuf)
	}

	return &Object{
		ID:           commitID,
		RelativeDate: string(parts[0]),
		Header:       string(parts[1]),
		Message:      string(parts[2]),
	}, nil
}
