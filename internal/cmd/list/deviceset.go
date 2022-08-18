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
	"github.com/docker/go-units"
	"github.com/spf13/cobra"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/project-flotta/flotta-dev-cli/internal/resources"
)

// deviceSetCmd represents the deviceSet command
var deviceSetCmd = &cobra.Command{
	Use:     "deviceset",
	Aliases: []string{"devicesets"},
	Short:   "List device-sets",
	RunE: func(cmd *cobra.Command, args []string) error {

		writer := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 8, 2, '\t', tabwriter.AlignRight)
		defer writer.Flush()
		fmt.Fprintf(writer, "%s\t%s\t%s\t\n", "NAME", "DEVICES", "CREATED")

		client, err := resources.NewClient()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "NewClient failed: %v\n", err)
			return err
		}

		// create a list of all device-sets
		deviceset, err := resources.NewEdgeDeviceSet(client, "")
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "NewEdgeDeviceSet failed: %v\n", err)
			return err
		}
		setsList, err := deviceset.List()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "List() device-set failed: %v\n", err)
			return err
		}

		// create a list of all registered devices
		device, err := resources.NewEdgeDevice(client, "")
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "NewEdgeDeviceSet failed: %v\n", err)
			return err
		}
		devicesList, err := device.List()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "List() device failed: %v\n", err)
			return err
		}

		// create a map of sets names and their devices
		devicesMap := make(map[string][]string)
		for _, dvc := range devicesList.Items {
			if labelValue, ok := dvc.Labels["flotta/member-of"]; ok {
				devicesMap[labelValue] = append(devicesMap[labelValue], dvc.Name)
			}
		}

		// output all device-sets and their devices
		for _, set := range setsList.Items {
			devices := strings.Join(devicesMap[set.Name], ", ")
			runningFor := units.HumanDuration(time.Now().UTC().Sub(set.CreationTimestamp.Time)) + " ago"
			fmt.Fprintf(writer, "%s\t%v\t%s\t\n", set.Name, devices, runningFor)
		}

		return nil
	},
}

func init() {
	// subcommand of list
	listCmd.AddCommand(deviceSetCmd)
}
