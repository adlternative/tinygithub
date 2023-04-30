package tree

import (
	"bufio"
	"context"
	"fmt"
	"github.com/adlternative/tinygithub/pkg/cmd"
	"github.com/adlternative/tinygithub/pkg/service/git/object"
	"strings"
)

type Entry struct {
	Mode object.Mode
	Type object.Type
	Oid  object.ID
	Path string
}

func ParseTreeLine(treeLine string) (*Entry, error) {
	parts := strings.Split(treeLine, "\t")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid tree line")
	}

	firstPart := parts[0]
	firstParts := strings.Fields(firstPart)
	if len(firstParts) != 3 {
		return nil, fmt.Errorf("invalid tree line")
	}

	mode, err := object.ParseMode(firstParts[0])
	if err != nil {
		return nil, err
	}

	parseType, err := object.ParseType(firstParts[1])
	if err != nil {
		return nil, err
	}

	oid, err := object.ParseID(firstParts[2])
	if err != nil {
		return nil, err
	}

	return &Entry{
		Mode: mode,
		Type: parseType,
		Oid:  oid,
		Path: parts[1],
	}, nil
}

func ParseTree(ctx context.Context, repoPath, revision string, treePath string) ([]*Entry, error) {
	var stderrBuf strings.Builder

	gitCmd := cmd.NewGitCommand("ls-tree").WithGitDir(repoPath).
		WithArgs(fmt.Sprintf("%s:%s", revision, treePath)).
		WithStderr(&stderrBuf)

	if err := gitCmd.Start(ctx); err != nil {
		return nil, fmt.Errorf("gitCmd start failed with %w", err)
	}

	var entries []*Entry
	scanner := bufio.NewScanner(gitCmd)

	for scanner.Scan() {
		entry, err := ParseTreeLine(scanner.Text())
		if err != nil {
			return nil, fmt.Errorf("parseTreeLine failed with %w", err)
		}
		entries = append(entries, entry)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner failed with %w", err)
	}

	if err := gitCmd.Wait(); err != nil {
		return nil, fmt.Errorf("git command failed with stderr:%v, error:%w", stderrBuf.String(), err)
	}
	return entries, nil
}
