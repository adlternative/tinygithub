package cmd

import (
	"context"
	"fmt"
	"github.com/adlternative/tinygithub/pkg/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io"
	"os"
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

	// if no stdout
	internalStdout io.ReadCloser
	// if no stdin
	internalStdin io.WriteCloser
	// if no stderr
	internalStderr io.ReadCloser
}

func (c *Command) Read(p []byte) (n int, err error) {
	return c.internalStdout.Read(p)
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

	gitBinPath := viper.GetString(config.GitBinPath)
	if gitBinPath == "" {
		gitBinPath = "git"
	}

	c.cmd = exec.CommandContext(ctx, gitBinPath, trueArgs...)

	if c.stdout != nil {
		c.cmd.Stdout = c.stdout
	} else {
		stdoutPipe, err := c.cmd.StdoutPipe()
		if err != nil {
			return err
		}
		c.internalStdout = stdoutPipe
	}
	if c.stderr != nil {
		c.cmd.Stderr = c.stderr
	} else {
		stderrPipe, err := c.cmd.StderrPipe()
		if err != nil {
			return err
		}
		c.internalStderr = stderrPipe
	}
	if c.stdin != nil {
		c.cmd.Stdin = c.stdin
	} else {
		stdinPipe, err := c.cmd.StdinPipe()
		if err != nil {
			return err
		}
		c.internalStdin = stdinPipe
	}
	if len(c.envs) > 0 {
		c.cmd.Env = append(c.cmd.Env, os.Environ()...)
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
	if c.internalStdin != nil {
		err := c.internalStdin.Close()
		if err != nil {
			return err
		}
	}
	if c.internalStdout != nil {
		_, _ = io.Copy(io.Discard, c.internalStdout)

		err := c.internalStdout.Close()
		if err != nil {
			return err
		}
	}
	if c.internalStderr != nil {
		_, _ = io.Copy(io.Discard, c.internalStderr)

		err := c.internalStderr.Close()
		if err != nil {
			return err
		}
	}

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
