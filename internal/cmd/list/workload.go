package list

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
	"strings"
	"text/tabwriter"
	"time"

	"github.com/docker/go-units"
	"github.com/spf13/cobra"

	"github.com/project-flotta/flotta-dev-cli/internal/resources"
)

// NewWorkloadCmd returns the workload command
func NewWorkloadCmd() *cobra.Command {
	workloadCmd := &cobra.Command{
		Use:     "workload",
		Aliases: []string{"workloads"},
		Short:   "List workloads",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := resources.NewClient()
			if err != nil {
				return err
			}

			workload, err := resources.NewEdgeWorkload(client)
			if err != nil {
				return err
			}

			workloads, err := workload.List()
			if err != nil {
				return err
			}

			if len(workloads.Items) == 0 {
				return fmt.Errorf("No resources were found.\n")
			}

			workloadsMap := make(map[string]string)

			// create a list of all registered devices
			device, err := resources.NewEdgeDevice(client, "")
			if err != nil {
				return err
			}
			devicesList, err := device.List()
			if err != nil {
				return err
			}

			writer := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 8, 2, '\t', tabwriter.AlignRight)
			defer writer.Flush()

			// loop over registered devices
			foundWorkload := false
			for _, dvc := range devicesList.Items {
				device, err := resources.NewEdgeDevice(client, dvc.Name)
				if err != nil {
					return err
				}

				// get workloads by device
				registeredDevice, err := device.Get()
				if err != nil {
					return err
				}
				workloads := registeredDevice.Status.Workloads
				for _, workload := range workloads {
					if !foundWorkload {
						foundWorkload = true
						fmt.Fprintf(writer, "%s\t%s\t%s\t\n", "NAME", "STATUS", "CREATED")
					}
					workloadsMap[workload.Name] = string(workload.Phase)
				}
			}

			errWorkloads := make([]string, 0)
			for _, workload := range workloads.Items {
				createdTime := workload.CreationTimestamp.Time
				formattedTime := units.HumanDuration(time.Now().UTC().Sub(createdTime)) + " ago"
				if ok := workloadsMap[workload.Name]; ok != "" {
					fmt.Fprintf(writer, "%s\t%v\t%s\t\n", workload.Name, workloadsMap[workload.Name], formattedTime)
				} else {
					errWorkloads = append(errWorkloads, workload.Name)
				}
			}

			fmt.Fprintf(writer, "\nfailed to get device status for workloads: %v\n", strings.Join(errWorkloads, ", "))

			return nil
		},
	}

	return workloadCmd
}

func init() {
	// subcommand of list
	listCmd.AddCommand(NewWorkloadCmd())
}
