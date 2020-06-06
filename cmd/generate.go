package cmd

import (
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/openfaas/faasd/pkg/assets"
)

func init() {
	configureGenerateFlags(generateCmd.Flags())
}

// generateConfig are the CLI flags used by the `faasd generate` command to deploy the faasd service
type generateConfig struct {
	// output is the destination path that generate will write files to
	output string
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate faasd configuration files",
	RunE:  generateConfigFiles,
}

func generateConfigFiles(cmd *cobra.Command, _ []string) error {
	cfg, err := parseGenerateFlags(cmd)
	if err != nil {
		return err
	}

	return assets.WriteConfigFiles(cfg.output)
}

// configureGenerateFlags will define the flags for the `faasd generate` command. The flag struct, configure, and
// parse are split like this to simplify testability.
func configureGenerateFlags(flags *flag.FlagSet) {
	flags.StringP("out", "o", "", "output directory to write the configuration files to")
}

// parseGenerateFlags will load the flag values into an upFlags object. Errors will be underlying
// Get errors from the pflag library.
func parseGenerateFlags(cmd *cobra.Command) (generateConfig, error) {
	parsed := generateConfig{}
	path, err := cmd.Flags().GetString("out")
	if err != nil {
		return parsed, errors.Wrap(err, "can not parse output path flag")
	}

	if path == "" {
		path, err = os.Getwd()
		if err != nil {
			return parsed, err
		}
	}

	parsed.output = path
	return parsed, err
}
