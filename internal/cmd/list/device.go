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
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/go-units"
	"github.com/spf13/cobra"
	"sort"
	"text/tabwriter"
	"time"
)

// NewDeviceCmd returns the device command
func NewDeviceCmd() *cobra.Command {
	deviceCmd := &cobra.Command{
		Use:     "device",
		Aliases: []string{"devices"},
		Short:   "List devices",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "NewClientWithOpts failed: %v\n", err)
				return err
			}

			// list of containers that contain the label flotta
			filter := filters.Arg("label", "flotta")
			opts := types.ContainerListOptions{All: true, Filters: filters.NewArgs(filter)}

			containers, err := cli.ContainerList(ctx, opts)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "ContainerList failed: %v\n", err)
				return err
			}

			// sort containers by container name
			sort.Slice(containers, func(i, j int) bool {
				return containers[i].Names[0] < containers[j].Names[0]
			})

			writer := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 8, 2, '\t', tabwriter.AlignRight)
			defer writer.Flush()

			// if there are no devices, print a message and return
			if len(containers) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No resources were found.\n")
				return nil
			} else {
				fmt.Fprintf(writer, "%s\t%s\t%s\t\n", "NAME", "STATUS", "CREATED")
			}

			for _, container := range containers {
				containerName := container.Names[0][1:]
				createdAt := time.Unix(container.Created, 0)
				runningFor := units.HumanDuration(time.Now().UTC().Sub(createdAt)) + " ago"

				fmt.Fprintf(writer, "%s\t%v\t%s\t\n", containerName, container.State, runningFor)
			}

			return nil
		},
	}

	return deviceCmd
}

func init() {
	// subcommand of list
	listCmd.AddCommand(NewDeviceCmd())
}
