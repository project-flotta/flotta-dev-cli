package add

import (
	"context"
	"fmt"
	"github.com/project-flotta/flotta-dev-cli/internal/resources"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
)

var (
	deviceSetName    string
	deviceSetSize    int
	deviceNamePrefix string
)

// NewDeviceSetCmd returns the device set command
func NewDeviceSetCmd() *cobra.Command {
	deviceSetCmd := &cobra.Command{
		Use:     "deviceset",
		Aliases: []string{"devicesets"},
		Short:   "Add a new device set with registered devices",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if deviceSetSize < 0 {
				err := fmt.Errorf("deviceSetSize is invalid: %d. Only positive values are allowed", deviceSetSize)
				fmt.Fprintf(cmd.OutOrStderr(), err.Error()+"\n")
				return err
			}

			// if devices prefix has not been specified, use deviceSetName as prefix
			if deviceNamePrefix == "" {
				deviceNamePrefix = deviceSetName
			}

			client, err := resources.NewClient()
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "NewClient failed %v\n", err)
				return err
			}

			deviceset, err := resources.NewEdgeDeviceSet(client, deviceSetName)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "NewEdgeDeviceSet failed: %v\n", err)
				return err
			}

			_, err = deviceset.Create(resources.EdgeDeviceSetConfig(deviceSetName))
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Create device-set failed: %v\n", err)
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "device-set '%s' was added\n", deviceSetName)

			// add devices to the deviceset
			devicesCreated := 0
			for i := 1; i <= deviceSetSize; i++ {
				deviceName := fmt.Sprintf("%s%d", deviceNamePrefix, i)
				err := NewDeviceToSet(deviceSetName, deviceName)
				if err != nil {
					fmt.Fprintf(cmd.OutOrStderr(), "NewDeviceToSet failed: %v. Device: %s\n", err, deviceName)
				} else {
					devicesCreated += 1
					fmt.Fprintf(cmd.OutOrStdout(), "device '%s' was added successfully to device-set '%s' (%d/%d)\n", deviceName, deviceSetName, devicesCreated, deviceSetSize)
				}
			}
			return nil
		},
	}

	// define command flags
	deviceSetCmd.Flags().StringVarP(&deviceSetName, "name", "n", "", "name of the deviceset to add")
	deviceSetCmd.Flags().IntVarP(&deviceSetSize, "size", "s", 0, "the amount of edge devices to be created and added to the device set")
	deviceSetCmd.Flags().StringVarP(&deviceNamePrefix, "prefix", "p", "", "the name prefix of the devices to add to the deviceset")

	// mark name flag as required
	err := deviceSetCmd.MarkFlagRequired("name")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to set flag `name` as required: %v\n", err)
		os.Exit(1)
	}

	return deviceSetCmd
}

func init() {
	// subcommand of add
	addCmd.AddCommand(NewDeviceSetCmd())
}

func NewDeviceToSet(deviceSetName, deviceName string) error {
	client, err := resources.NewClient()
	if err != nil {
		return err
	}

	device, err := resources.NewEdgeDevice(client, deviceName)
	if err != nil {
		return err
	}

	err = device.Register()
	if err != nil {
		// if device.Register() failed, remove the container
		err2 := device.Remove()
		if err2 != nil {
			return err2
		}

		return err
	}

	// get the new device in order to add 'flotta/member-of' label
	dvc, err := device.Get()
	if err != nil {
		return err
	}

	// update the device
	dvc.Labels["flotta/member-of"] = deviceSetName
	_, err = client.EdgeDevices("default").Update(context.TODO(), dvc, v1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}
