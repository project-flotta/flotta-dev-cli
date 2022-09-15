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
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"sort"

	"github.com/docker/docker/client"
	"github.com/project-flotta/flotta-dev-cli/internal/resources"
	"github.com/spf13/cobra"
)

var deviceSetName string
var deleteDevices bool

// NewDeviceSetCmd return the deviceset command
func NewDeviceSetCmd() *cobra.Command {
	deviceSetCmd := &cobra.Command{
		Use:     "deviceset",
		Aliases: []string{"devicesets"},
		Short:   "Delete a deviceset from flotta",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := resources.NewClient()
			if err != nil {
				return err
			}

			deviceset, err := resources.NewEdgeDeviceSet(client, deviceSetName)
			if err != nil {
				return err
			}

			err = deviceset.Remove(deviceSetName)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "deviceset '%v' was deleted \n", deviceSetName)

			devices, err := getDevicesNamesList()
			if err != nil {
				return err
			}

			for _, deviceName := range devices {
				err := updateDeviceAfterSetDeletion(deviceName, cmd)
				if err != nil {
					fmt.Fprintf(cmd.OutOrStderr(), "updateDeviceAfterSetDeletion failed: %v\n", err)
				}
			}
			return nil
		},
	}

	// define command flags
	deviceSetCmd.Flags().StringVarP(&deviceSetName, "name", "n", "", "name of the device-set to delete")
	deviceSetCmd.Flags().BoolVarP(&deleteDevices, "all", "a", false, "mark as true will remove all resources related to the device-set")

	// mark name flag as required
	err := deviceSetCmd.MarkFlagRequired("name")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to set flag `name` as required: %v\n", err)
		os.Exit(1)
	}

	// if "all" flag was provided, set "deleteDevices" flag to true
	if deviceSetCmd.Flags().Lookup("all").Changed {
		deleteDevices = true
	}

	return deviceSetCmd
}

func init() {
	// subcommand of delete
	deleteCmd.AddCommand(NewDeviceSetCmd())
}

func getDevicesNamesList() ([]string, error) {
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

	var names []string
	for _, container := range containers {
		names = append(names, container.Names[0][1:])
	}

	return names, nil
}

func updateDeviceAfterSetDeletion(deviceName string, cmd *cobra.Command) error {
	client, err := resources.NewClient()
	if err != nil {
		return err
	}

	device, err := resources.NewEdgeDevice(client, deviceName)
	if err != nil {
		return err
	}

	dvc, err := device.Get()
	if err != nil {
		return err
	}

	// check if the device has 'flotta/member-of:<deviceSetName>' label
	if labelValue, ok := dvc.Labels["flotta/member-of"]; ok && labelValue == deviceSetName {
		if deleteDevices {
			// delete the device
			err = device.Unregister()
			if err != nil {
				return err
			}
			err = device.Remove()
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "device '%v' was deleted successfully\n", deviceName)
		} else {
			// remove the label and update the device
			delete(dvc.Labels, "flotta/member-of")
			_, err = client.EdgeDevices("default").Update(context.TODO(), dvc, v1.UpdateOptions{})
			if err != nil {
				return err
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "the label 'flotta/member-of:%s' was removed successfully from device '%s'\n", deviceSetName, deviceName)
			}
		}
	}
	return nil
}
