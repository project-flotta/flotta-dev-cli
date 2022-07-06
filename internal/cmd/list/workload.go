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
	"context"
	"fmt"
	"github.com/project-flotta/flotta-dev-cli/internal/resources"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/go-units"
	"github.com/spf13/cobra"
	"os"
	"sort"
	"text/tabwriter"
	"time"
)

// workloadCmd represents the workload command
var workloadCmd = &cobra.Command{
	Use:   "workload",
	Short: "list workloads",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			panic(err)
		}

		// list of containers that contain the label flotta
		filter := filters.Arg("label", "flotta")
		opts := types.ContainerListOptions{All: true, Filters: filters.NewArgs(filter)}

		containers, err := cli.ContainerList(ctx, opts)
		if err != nil {
			panic(err)
		}

		// sort containers by container name
		sort.Slice(containers, func(i, j int) bool {
			return containers[i].Names[0] < containers[j].Names[0]
		})

		writer := tabwriter.NewWriter(os.Stdout, 0, 8, 2, '\t', tabwriter.AlignRight)

		defer writer.Flush()

		fmt.Fprintf(writer, "%s\t%s\t%s\t\n", "NAME", "STATUS", "CREATED")

		client, err := resources.NewClient()
		if err != nil {
			fmt.Printf("NewClient failed: %v\n", err)
			return
		}

		// loop over registered devices
		for _, container := range containers {
			containerName := container.Names[0][1:]
			device, err := resources.NewEdgeDevice(client, containerName)
			if err != nil {
				fmt.Printf("NewEdgeDevice failed: %v\n", err)
				return
			}

			// get workloads by device
			registeredDevice, err := device.Get()
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
