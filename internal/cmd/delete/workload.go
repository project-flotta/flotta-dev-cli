package delete

/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"fmt"
	"os"

	"github.com/project-flotta/flotta-dev-cli/internal/resources"
	"github.com/spf13/cobra"
)

var workloadName string

// workloadCmd represents the workload command
var workloadCmd = &cobra.Command{
	Use:   "workload",
	Aliases: []string{"workloads"},
	Short: "Delete a workload from flotta",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := resources.NewClient()
		if err != nil {
			fmt.Printf("NewClient failed: %v\n", err)
			return
		}

		workload, err := resources.NewEdgeWorkload(client)
		if err != nil {
			fmt.Printf("NewEdgeWorkload failed: %v\n", err)
			return
		}

		err = workload.Remove(workloadName)
		if err != nil {
			fmt.Printf("Remove workload failed: %v\n", err)
			return
		}

		fmt.Printf("workload '%v' was deleted \n", workloadName)
	},
}

func init() {
	// subcommand of delete
	deleteCmd.AddCommand(workloadCmd)

	// define command flags
	workloadCmd.Flags().StringVarP(&workloadName, "name", "n", "", "name of the workload to delete")
	err := workloadCmd.MarkFlagRequired("name")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to set flag `name` as required: %v\n", err)
		os.Exit(1)
	}
}
