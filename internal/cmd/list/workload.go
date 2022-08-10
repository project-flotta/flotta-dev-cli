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

// workloadCmd represents the workload command
var workloadCmd = &cobra.Command{
	Use:     "workload",
	Aliases: []string{"workloads"},
	Short:   "List workloads",
	Run: func(cmd *cobra.Command, args []string) {

		writer := tabwriter.NewWriter(os.Stdout, 0, 8, 2, '\t', tabwriter.AlignRight)
		defer writer.Flush()
		fmt.Fprintf(writer, "%s\t%s\t%s\t\n", "NAME", "STATUS", "CREATED")

		client, err := resources.NewClient()
		if err != nil {
			fmt.Printf("NewClient failed: %v\n", err)
			return
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
				fmt.Printf("NewEdgeDevice failed: %v\n", err)
				return
			}

			// get workloads by device
			registeredDevice, err := device.Get()
			if err != nil {
				fmt.Printf("Get device failed: %v\n", err)
				return
			}
			workloads := registeredDevice.Status.Workloads
			for _, workload := range workloads {
				createdTime := getWorkloadCreationTime(workload.Name)
				formattedTime := units.HumanDuration(time.Now().UTC().Sub(createdTime)) + " ago"
				fmt.Fprintf(writer, "%s\t%v\t%s\t\n", workload.Name, workload.Phase, formattedTime)
			}
		}
	},
}

func init() {
	// subcommand of list
	listCmd.AddCommand(workloadCmd)
}

func getWorkloadCreationTime(name string) time.Time {
	client, err := resources.NewClient()
	if err != nil {
		fmt.Printf("NewClient failed: %v\n", err)
		return time.Time{}
	}

	w, err := resources.NewEdgeWorkload(client)
	if err != nil {
		fmt.Printf("NewEdgeWorkload failed: %v\n", err)
		return time.Time{}
	}

	workload, err := w.Get(name)
	if err != nil {
		fmt.Printf("Get workload failed: %v\n", err)
		return time.Time{}
	}

	return workload.CreationTimestamp.Time
}
