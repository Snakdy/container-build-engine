package cmd

import (
	"context"
	"fmt"
	cbev1 "github.com/Snakdy/container-build-engine/pkg/api/v1"
	"github.com/Snakdy/container-build-engine/pkg/builder"
	"github.com/Snakdy/container-build-engine/pkg/containers"
	"github.com/Snakdy/container-build-engine/pkg/pipelines"
	"github.com/go-logr/logr"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/yaml"
	"os"
	"path/filepath"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "build an image",
	RunE:  build,
}

const (
	flagConfig = "config"

	flagSave  = "save"
	flagImage = "image"
	flagTag   = "tag"

	flagPlatform = "platform"
)

func init() {
	buildCmd.Flags().StringP(flagConfig, "c", "", "path to an image configuration file")

	buildCmd.Flags().String(flagSave, "", "path to save the image as a tar archive")
	buildCmd.Flags().String(flagImage, "", "oci image path (without tag) to push the image")
	buildCmd.Flags().StringArrayP(flagTag, "t", nil, "tags to push")

	buildCmd.Flags().String(flagPlatform, "linux/amd64", "build platform")

	_ = buildCmd.MarkFlagRequired(flagConfig)
	_ = buildCmd.MarkFlagFilename(flagConfig, ".yaml", ".yml")

	buildCmd.MarkFlagsMutuallyExclusive(flagSave, flagImage)
	buildCmd.MarkFlagsRequiredTogether(flagImage, flagTag)
}

func build(cmd *cobra.Command, _ []string) error {
	log := logr.FromContextOrDiscard(cmd.Context())

	configPath, _ := cmd.Flags().GetString(flagConfig)
	localPath, _ := cmd.Flags().GetString(flagSave)
	ociPath, _ := cmd.Flags().GetString(flagImage)
	tags, _ := cmd.Flags().GetStringArray(flagTag)

	platform, _ := cmd.Flags().GetString(flagPlatform)
	imgPlatform, err := v1.ParsePlatform(platform)
	if err != nil {
		log.Error(err, "failed to parse platform")
		return err
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	// read the config file
	cfg, err := readConfig(configPath)
	if err != nil {
		return err
	}

	b, err := newBuilder(cmd.Context(), cfg, nil, wd)
	if err != nil {
		return err
	}
	img, err := b.Build(cmd.Context(), imgPlatform)
	if err != nil {
		return err
	}

	if localPath != "" {
		return containers.Save(cmd.Context(), img, "image", localPath)
	}
	// push all tags
	for _, t := range tags {
		if err := containers.Push(cmd.Context(), img, fmt.Sprintf("%s:%s", ociPath, t)); err != nil {
			return err
		}
	}

	return nil
}

func readConfig(s string) (cbev1.Pipeline, error) {
	f, err := os.Open(filepath.Clean(s))
	if err != nil {
		return cbev1.Pipeline{}, err
	}

	var config cbev1.Pipeline
	if err := yaml.NewYAMLOrJSONDecoder(f, 4).Decode(&config); err != nil {
		return cbev1.Pipeline{}, err
	}
	return config, nil
}

// newBuilder converts our cbev1.Pipeline into the underlying pipeline
// resources.
func newBuilder(ctx context.Context, pipeline cbev1.Pipeline, statementFinder pipelines.StatementFinder, workingDir string) (*builder.Builder, error) {
	// if the user didn't specify a statement finder, we
	// need to use the default one
	if statementFinder == nil {
		statementFinder = pipelines.Find
	}

	orderedStatements := make([]pipelines.OrderedPipelineStatement, len(pipeline.Statements))
	for i := range pipeline.Statements {
		statement := statementFinder(pipeline.Statements[i].Name, pipeline.Statements[i].Options)
		if statement == nil {
			return nil, fmt.Errorf("could not find statement '%s'", pipeline.Statements[i].Name)
		}
		orderedStatements[i] = pipelines.OrderedPipelineStatement{
			ID:        pipeline.Statements[i].ID,
			Statement: statement,
			DependsOn: pipeline.Statements[i].DependsOn,
		}
	}

	return builder.NewBuilder(ctx, pipeline.Base, orderedStatements, builder.Options{
		WorkingDir: workingDir,
		Entrypoint: pipeline.Config.Entrypoint,
		Command:    pipeline.Config.Command,
	})
}
