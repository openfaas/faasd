package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/containerd/containerd/runtime/v2/logging"
	"github.com/coreos/go-systemd/journal"
	"github.com/spf13/cobra"
)

func CollectCommand() *cobra.Command {
	return collectCmd
}

var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "Collect logs to the journal",
	RunE:  runCollect,
}

func runCollect(_ *cobra.Command, _ []string) error {
	logging.Run(logStdio)
	return nil
}

// logStdio copied from
// https://github.com/containerd/containerd/pull/3085
// https://github.com/stellarproject/orbit
func logStdio(ctx context.Context, config *logging.Config, ready func() error) error {
	// construct any log metadata for the container
	vars := map[string]string{
		"SYSLOG_IDENTIFIER": fmt.Sprintf("%s:%s", config.Namespace, config.ID),
	}
	var wg sync.WaitGroup
	wg.Add(2)
	// forward both stdout and stderr to the journal
	go copy(&wg, config.Stdout, journal.PriInfo, vars)
	go copy(&wg, config.Stderr, journal.PriErr, vars)
	// signal that we are ready and setup for the container to be started
	if err := ready(); err != nil {
		return err
	}
	wg.Wait()
	return nil
}

func copy(wg *sync.WaitGroup, r io.Reader, pri journal.Priority, vars map[string]string) {
	defer wg.Done()
	s := bufio.NewScanner(r)
	for s.Scan() {
		if s.Err() != nil {
			return
		}
		journal.Send(s.Text(), pri, vars)
	}
}
