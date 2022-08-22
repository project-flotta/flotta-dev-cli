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

// NewWorkloadCmd returns the workload command
func NewWorkloadCmd() *cobra.Command {
	workloadCmd := &cobra.Command{
		Use:     "workload",
		Aliases: []string{"workloads"},
		Short:   "Delete a workload from flotta",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := resources.NewClient()
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "NewClient failed: %v\n", err)
				return err
			}

			workload, err := resources.NewEdgeWorkload(client)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "NewEdgeWorkload failed: %v\n", err)
				return err
			}

			err = workload.Remove(workloadName)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Remove workload failed: %v\n", err)
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "workload '%v' was deleted \n", workloadName)
			return nil
		},
	}

	// define command flags
	workloadCmd.Flags().StringVarP(&workloadName, "name", "n", "", "name of the workload to delete")
	err := workloadCmd.MarkFlagRequired("name")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to set flag `name` as required: %v\n", err)
		os.Exit(1)
	}

	return workloadCmd
}

func init() {
	// subcommand of delete
	deleteCmd.AddCommand(NewWorkloadCmd())
}
