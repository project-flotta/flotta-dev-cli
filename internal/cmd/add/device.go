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

package add

import (
	"fmt"
	"github.com/project-flotta/flotta-dev-cli/internal/resources"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var deviceName string

// NewDeviceCmd returns the device command
func NewDeviceCmd() *cobra.Command {
	deviceCmd := &cobra.Command{
		Use:     "device",
		Aliases: []string{"devices"},
		Short:   "Add a new device",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := resources.NewClient()
			if err != nil {
				return err
			}

			device, err := resources.NewEdgeDevice(client, deviceName)
			if err != nil {
				return err
			}

			dvc, err := device.Get()
			if err == nil && dvc != nil {
				return fmt.Errorf("edgedevices.management.project-flotta.io \"%s\" already exists", deviceName)
			} else {
				errSubstring := fmt.Sprintf("edgedevices.management.project-flotta.io \"%s\" not found", deviceName)
				if !strings.Contains(err.Error(), errSubstring) {
					return err
				}
			}

			err = device.Register(deviceImage)
			if err != nil {
				// if device.Register() failed, remove the container
				err2 := device.Remove()
				if err2 != nil {
					return err2
				}

				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "device '%s' was added\n", device.GetName())
			return nil
		},
	}

	// define command flags
	deviceCmd.Flags().StringVarP(&deviceName, "name", "n", "", "name of the device to add")
	deviceCmd.Flags().StringVarP(&deviceImage, "image", "i", "", "image of the device to add")

	// mark name flag as required
	err := deviceCmd.MarkFlagRequired("name")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to set flag `name` as required: %v\n", err)
		os.Exit(1)
	}
	return deviceCmd
}

func init() {
	// subcommand of add
	addCmd.AddCommand(NewDeviceCmd())
}
