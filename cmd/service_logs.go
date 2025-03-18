package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	goexecute "github.com/alexellis/go-execute/v2"
	"github.com/spf13/cobra"
)

func makeServiceLogsCmd() *cobra.Command {
	var command = &cobra.Command{
		Use:   "logs",
		Short: "View logs for a service",
		Long:  `View logs for a service created by faasd from the docker-compose.yml file.`,
		Example: `  ## View logs for the gateway for the last hour
  faasd service logs gateway --since 1h

  ## View logs for the cron-connector, and tail them
  faasd service logs cron-connector -f
`,
	}

	command.Flags().Duration("since", 10*time.Minute, "How far back in time to include logs")
	command.Flags().BoolP("follow", "f", false, "Follow the logs")

	command.RunE = runServiceLogsE
	command.PreRunE = preRunServiceLogsE

	return command
}

func runServiceLogsE(cmd *cobra.Command, args []string) error {
	name := args[0]

	namespace, _ := cmd.Flags().GetString("namespace")
	follow, _ := cmd.Flags().GetBool("follow")
	since, _ := cmd.Flags().GetDuration("since")

	journalTask := goexecute.ExecTask{
		Command:     "journalctl",
		Args:        []string{"-o", "cat", "-t", fmt.Sprintf("%s:%s", namespace, name)},
		StreamStdio: true,
	}

	if follow {
		journalTask.Args = append(journalTask.Args, "-f")
	}

	if since != 0 {
		// Calculate the timestamp that is 'age' duration ago
		sinceTime := time.Now().Add(-since)
		// Format according to journalctl's expected format: "2012-10-30 18:17:16"
		formattedTime := sinceTime.Format("2006-01-02 15:04:05")
		journalTask.Args = append(journalTask.Args, fmt.Sprintf("--since=%s", formattedTime))
	}

	res, err := journalTask.Execute(context.Background())
	if err != nil {
		return err
	}

	if res.ExitCode != 0 {
		return fmt.Errorf("failed to get logs for service %s: %s", name, res.Stderr)
	}

	return nil
}

func preRunServiceLogsE(cmd *cobra.Command, args []string) error {

	if os.Geteuid() != 0 {
		return errors.New("this command must be run as root")
	}

	if len(args) == 0 {
		return errors.New("service name is required as an argument")
	}

	namespace, _ := cmd.Flags().GetString("namespace")
	if namespace == "" {
		return errors.New("namespace is required")
	}

	return nil
}
