package cmd

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os/exec"
)

type Command struct {
	cmd *exec.Cmd

	subCmd     string
	preOptions []string
	envs       []string
	options    []string
	args       []string

	stdout io.Writer
	stdin  io.Reader
	stderr io.Writer
}

func NewGitCommand(subCmd string) *Command {
	return &Command{
		subCmd: subCmd,
	}
}

func (c *Command) WithStdout(stdout io.Writer) *Command {
	c.stdout = stdout
	return c
}

func (c *Command) WithStdin(stdin io.Reader) *Command {
	c.stdin = stdin
	return c
}

func (c *Command) WithStderr(stderr io.Writer) *Command {
	c.stderr = stderr
	return c
}

func (c *Command) WithEnv(env ...string) *Command {
	c.envs = append(c.envs, env...)
	return c
}

func (c *Command) WithOptions(opts ...string) *Command {
	c.options = append(c.options, opts...)
	return c
}

func (c *Command) WithArgs(args ...string) *Command {
	c.args = append(c.args, args...)
	return c
}

func (c *Command) WithGitDir(gitDir string) *Command {
	c.preOptions = append(c.preOptions, fmt.Sprintf("--git-dir=%s", gitDir))
	return c
}

func (c *Command) Start(ctx context.Context) error {
	var trueArgs []string

	trueArgs = append(trueArgs, c.preOptions...)
	trueArgs = append(trueArgs, c.subCmd)
	trueArgs = append(trueArgs, c.options...)
	trueArgs = append(trueArgs, c.args...)

	c.cmd = exec.CommandContext(ctx, "git", trueArgs...)

	if c.stdout != nil {
		c.cmd.Stdout = c.stdout
	}
	if c.stderr != nil {
		c.cmd.Stderr = c.stderr
	}
	if c.stdin != nil {
		c.cmd.Stdin = c.stdin
	}
	if len(c.envs) > 0 {
		c.cmd.Env = append(c.cmd.Env, c.envs...)
	}

	log.Debug("git command: ", c.cmd.String())

	err := c.cmd.Start()
	if err != nil {
		return err
	}
	return nil
}

func (c *Command) Wait() error {
	err := c.cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}

func (c *Command) Run(ctx context.Context) error {
	err := c.Start(ctx)
	if err != nil {
		return err
	}

	err = c.cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}
