package kor

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yonahd/kor/pkg/kor"
	"github.com/yonahd/kor/pkg/utils"
)

var argoRolloutsCmd = &cobra.Command{
	Use:     "argorollouts",
	Aliases: []string{"argorollouts"},
	Short:   "Gets unused argo rollouts",
	Args:    cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		clientset, _ := kor.GetKubeClient(kubeconfig)
		if response, err := kor.GetUnusedArgoRollouts(filterOptions, clientset, outputFormat, opts); err != nil {
			fmt.Println(err)
		} else {
			utils.PrintLogo(outputFormat)
			fmt.Println(response)
		}
	},
}

func init() {
	rootCmd.AddCommand(argoRolloutsCmd)
}
