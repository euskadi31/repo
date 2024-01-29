package cmd

import (
	"fmt"
	"os"
	"path"

	"dario.cat/mergo"
	"github.com/euskadi31/repo/pkg/config"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	golangciFileName         = ".golangci.yml"
	golangciOverrideFileName = ".golangci.override.yml"
)

func fileExists(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}

	return true
}

// golangciLintCmd represents the golangciLint command
var golangciLintCmd = &cobra.Command{
	Use:   "golangci-lint",
	Short: "sync golangci-lint configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		repoDir, err := config.GetRepoDir()
		if err != nil {
			return fmt.Errorf("cannot get repo directory: %w", err)
		}

		cfg, err := config.Read()
		if err != nil {
			return fmt.Errorf("cannot read config: %w", err)
		}

		// check if template file exists if not return error
		if _, err := os.Stat(path.Join(repoDir, golangciFileName)); os.IsNotExist(err) {
			return fmt.Errorf(".golangci.yml template file does not exist")
		}

		b, err := os.ReadFile(path.Join(repoDir, golangciFileName))
		if err != nil {
			return fmt.Errorf("cannot read .golangci.yml template file: %w", err)
		}

		for _, rdir := range cfg.Repos {
			log.Info().Str("repo", rdir).Msg("sync golangci-lint configuration")

			var master map[string]interface{}

			if err := yaml.Unmarshal(b, &master); err != nil {
				return fmt.Errorf("cannot unmarshal .golangci.yml template file: %w", err)
			}

			content, err := getGolangCILintFile(rdir, master)
			if err != nil {
				return fmt.Errorf("cannot get golangci-lint file: %w", err)
			}

			f, err := os.Create(path.Join(rdir, golangciFileName))
			if err != nil {
				return fmt.Errorf("cannot open golangci-lint file: %w", err)
			}

			defer func() {
				if err := f.Close(); err != nil {
					log.Error().Err(err).Msg("cannot close golangci-lint file")
				}
			}()

			enc := yaml.NewEncoder(f)
			enc.SetIndent(2)

			if err := enc.Encode(content); err != nil {
				return fmt.Errorf("cannot encode golangci-lint file: %w", err)
			}

			if err := enc.Close(); err != nil {
				return fmt.Errorf("cannot close golangci-lint file: %w", err)
			}
		}

		return nil
	},
}

func getGolangCILintFile(dir string, master map[string]interface{}) (map[string]interface{}, error) {
	if !fileExists(path.Join(dir, golangciOverrideFileName)) {
		return master, nil
	}

	var override map[string]interface{}

	b, err := os.ReadFile(path.Join(dir, golangciOverrideFileName))
	if err != nil {
		return nil, fmt.Errorf("cannot read .golangci.override.yml file: %w", err)
	}

	if err := yaml.Unmarshal(b, &override); err != nil {
		return nil, fmt.Errorf("cannot unmarshal .golangci.override.yml file: %w", err)
	}

	// return mergeMaps2(master, override), nil
	return mergeMaps(master, override)
}

func mergeMaps(a, b map[string]interface{}) (map[string]interface{}, error) {
	if err := mergo.Merge(&a, b, mergo.WithOverride, mergo.WithAppendSlice); err != nil {
		return nil, fmt.Errorf("cannot merge maps: %w", err)
	}

	return a, nil
}

func mergeMaps2(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}

	for k, v := range b {
		if v2, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = mergeMaps2(bv, v2)

					continue
				}
			}
		} else if v2, ok := v.([]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.([]interface{}); ok {
					out[k] = append(bv, v2...)

					continue
				}
			}
		}

		out[k] = v
	}

	return out
}

/*
func mergeMaps2(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}

	for k, v := range b {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = mergeMaps2(bv, v)

					continue
				}
			}
		}

		out[k] = v
	}

	return out
}
*/

func init() {
	syncCmd.AddCommand(golangciLintCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// golangciLintCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// golangciLintCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
