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
	"os"
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
		RunE: func(cmd *cobra.Command, args []string) error {

			writer := tabwriter.NewWriter(os.Stdout, 0, 8, 2, '\t', tabwriter.AlignRight)
			defer writer.Flush()
			fmt.Fprintf(writer, "%s\t%s\t%s\t\n", "NAME", "STATUS", "CREATED")

			client, err := resources.NewClient()
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "NewClient failed: %v\n", err)
				return err
			}

			// create a list of all registered devices
			device, err := resources.NewEdgeDevice(client, "")
			if err != nil {
				fmt.Printf("NewEdgeDeviceSet failed: %v\n", err)
			}
			devicesList, err := device.List()
			if err != nil {
				fmt.Printf("List() device failed: %v\n", err)
			}

			// loop over registered devices
			for _, dvc := range devicesList.Items {
				device, err := resources.NewEdgeDevice(client, dvc.Name)
				if err != nil {
					fmt.Fprintf(cmd.OutOrStderr(), "NewEdgeDevice failed: %v\n", err)
					return err
				}

				// get workloads by device
				registeredDevice, err := device.Get()
				if err != nil {
					fmt.Fprintf(cmd.OutOrStderr(), "Get device failed: %v\n", err)
					return err
				}
				workloads := registeredDevice.Status.Workloads
				for _, workload := range workloads {
					createdTime, err := getWorkloadCreationTime(workload.Name)
					if err != nil {
						fmt.Fprintf(cmd.OutOrStderr(), "getWorkloadCreationTime failed: %v\n", err)
						return err
					}
					formattedTime := units.HumanDuration(time.Now().UTC().Sub(createdTime)) + " ago"
					fmt.Fprintf(writer, "%s\t%v\t%s\t\n", workload.Name, workload.Phase, formattedTime)
				}
			}
			return nil
		},
	}

	return workloadCmd
}

func init() {
	// subcommand of list
	listCmd.AddCommand(NewWorkloadCmd())
}

func getWorkloadCreationTime(name string) (time.Time, error) {
	client, err := resources.NewClient()
	if err != nil {
		return time.Time{}, err
	}

	w, err := resources.NewEdgeWorkload(client)
	if err != nil {
		return time.Time{}, err
	}

	workload, err := w.Get(name)
	if err != nil {
		return time.Time{}, err
	}

	return workload.CreationTimestamp.Time, nil
}
