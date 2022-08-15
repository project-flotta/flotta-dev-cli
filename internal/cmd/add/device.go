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
	"os"

	"github.com/project-flotta/flotta-dev-cli/internal/resources"
	"github.com/spf13/cobra"
)

var deviceName string

// deviceCmd represents the device command
var deviceCmd = &cobra.Command{
	Use:     "device",
	Aliases: []string{"devices"},
	Short:   "Add a new device",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := resources.NewClient()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "NewClient failed: %v\n", err)
			return err
		}

		device, err := resources.NewEdgeDevice(client, deviceName)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "NewEdgeDevice failed: %v\n", err)
			return err
		}

		err = device.Register()
		if err != nil {
			// if device.Register() failed, remove the container
			err2 := device.Remove()
			if err2 != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Remove device that failed to register failed: %v\n", err2)
				return err
			}

			fmt.Fprintf(cmd.OutOrStderr(), "Register device failed: %v\n", err)
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "device '%s' was added\n", device.GetName())
		return nil
	},
}

func init() {
	// subcommand of add
	addCmd.AddCommand(deviceCmd)

	// define command flags
	deviceCmd.Flags().StringVarP(&deviceName, "name", "n", "", "name of the device to add")
	err := deviceCmd.MarkFlagRequired("name")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to set flag `name` as required: %v\n", err)
		os.Exit(1)
	}
}
