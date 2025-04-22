/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yichenj/gomanus/internal/pkg/agent"
	"github.com/yichenj/gomanus/internal/pkg/cli"
	"github.com/yichenj/gomanus/internal/pkg/llm"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := os.Getenv("API_KEY")
		baseURL := os.Getenv("BASE_URL")

		rm := llm.ModelSetting{
			ModelName: cmd.Flag("reason-model").Value.String(),
			BaseURL:   baseURL,
			APIKey:    apiKey,
		}
		lm := llm.ModelSetting{
			ModelName: cmd.Flag("llm-model").Value.String(),
			BaseURL:   baseURL,
			APIKey:    apiKey,
		}
		vm := llm.ModelSetting{
			ModelName: cmd.Flag("vision-model").Value.String(),
			BaseURL:   baseURL,
			APIKey:    apiKey,
		}

		ctx := context.Background()
		agent, err := agent.NewAgent(ctx, map[llm.ModelType]llm.ModelSetting{
			llm.ModelTypeLanguageModel:   lm,
			llm.ModelTypeReasonModel:     rm,
			llm.ModelTypeMultiModalModel: vm,
		})
		if err != nil {
			fmt.Printf("create agent err: %s\n", err.Error())
			return
		}
		cli.DoConversations(ctx, agent)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
