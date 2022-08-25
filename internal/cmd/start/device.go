package start

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

var deviceName string

// NewDeviceCmd returns the device command
func NewDeviceCmd() *cobra.Command {
	deviceCmd := &cobra.Command{
		Use:     "device",
		Aliases: []string{"devices"},
		Short:   "Start a device",
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

			err = device.Start()
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Start failed: %v\n", err)
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "device '%v' was started \n", device.GetName())
			return nil
		},
	}

	// define command flags
	deviceCmd.Flags().StringVarP(&deviceName, "name", "n", "", "name of the device to start")
	err := deviceCmd.MarkFlagRequired("name")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to set flag `name` as required: %v\n", err)
		os.Exit(1)
	}

	return deviceCmd
}

func init() {
	// subcommand of start
	startCmd.AddCommand(NewDeviceCmd())
}
