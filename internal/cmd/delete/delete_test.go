package delete_test

import (
	"bytes"
	"context"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/project-flotta/flotta-dev-cli/internal/resources"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spf13/cobra"

	. "github.com/project-flotta/flotta-dev-cli/internal/cmd"
	. "github.com/project-flotta/flotta-dev-cli/internal/cmd/delete"
)

var (
	device    resources.EdgeDevice
	workload  resources.EdgeWorkload
	deviceSet resources.EdgeDeviceSet
)

const (
	deviceName   = "device1"
	workloadName = "workload1"
	setName      = "set1"
	defaultImage = "quay.io/project-flotta/nginx:1.21.6"
)

var _ = Describe("Delete", Ordered, func() {
	var (
		actualOut *bytes.Buffer
		actualErr *bytes.Buffer
		rootCmd   *cobra.Command
		deleteCmd *cobra.Command
	)

	BeforeEach(func() {
		// create a new buffer to capture stdout and stderr
		actualOut = new(bytes.Buffer)
		actualErr = new(bytes.Buffer)

		// create a new root command with the buffers as stdout and stderr
		rootCmd = NewRootCmd()
		rootCmd.SetOut(actualOut)
		rootCmd.SetErr(actualErr)

		deleteCmd = NewDeleteCmd()
		rootCmd.AddCommand(deleteCmd)
		deleteCmd.AddCommand(NewDeviceCmd(), NewWorkloadCmd(), NewDeviceSetCmd())
	})

	Context("Sanity", Ordered, func() {
		BeforeAll(func() {
			initializeResources()
		})

		AfterEach(func() {
			actualOut.Reset()
			actualErr.Reset()
		})

		It("should delete an edge workload", func() {
			// given
			rootCmd.SetArgs([]string{"delete", "workload", "--name", workloadName})

			// when
			err := rootCmd.Execute()
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(actualOut.String()).To(Equal(fmt.Sprintf("workload '%s' was deleted \n", workloadName)))
			Expect(actualErr.String()).To(BeEmpty())

		})

		It("should delete an edge device", func() {
			// given
			rootCmd.SetArgs([]string{"delete", "device", "--name", deviceName})

			// when
			err := rootCmd.Execute()
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(actualOut.String()).To(Equal(fmt.Sprintf("device '%s' was deleted \n", deviceName)))
			Expect(actualErr.String()).To(BeEmpty())
		})

		It("should delete an edge device set", func() {
			// given
			rootCmd.SetArgs([]string{"delete", "deviceset", "--name", setName})

			// when
			err := rootCmd.Execute()
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(actualOut.String()).To(Equal(fmt.Sprintf("deviceset '%s' was deleted \n", setName)))
			Expect(actualErr.String()).To(BeEmpty())
		})
	})

	Context("Required flags", func() {
		AfterEach(func() {
			actualOut.Reset()
			actualErr.Reset()
		})

		It("should fail to delete a device without a name", func() {
			// given
			rootCmd.SetArgs([]string{"delete", "device"})

			// when
			err := rootCmd.Execute()
			Expect(err).To(HaveOccurred())

			// then
			Expect(actualErr.String()).To(Equal("Error: required flag(s) \"name\" not set\n"))
		})

		It("should fail to delete a workload without a name", func() {
			// given
			rootCmd.SetArgs([]string{"delete", "workload"})

			// when
			err := rootCmd.Execute()
			Expect(err).To(HaveOccurred())

			// then
			Expect(actualErr.String()).To(Equal("Error: required flag(s) \"name\" not set\n"))
		})

		It("should fail to delete a device-set without a name", func() {
			// given
			rootCmd.SetArgs([]string{"delete", "deviceset"})

			// when
			err := rootCmd.Execute()
			Expect(err).To(HaveOccurred())

			// then
			Expect(actualErr.String()).To(Equal("Error: required flag(s) \"name\" not set\n"))
		})
	})

	Context("delete all devices of a device set", func() {
		It("should delete all devices of a device set", func() {
			// given
			initializeSetWithDevices()
			rootCmd.SetArgs([]string{"delete", "deviceset", "--name", setName, "--all"})

			// when
			err := rootCmd.Execute()
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(actualOut.String()).To(Equal(fmt.Sprintf("deviceset '%s' was deleted \ndevice '%s' was deleted successfully\ndevice '%s' was deleted successfully\n", setName, "dvc1", "dvc2")))
			Expect(actualErr.String()).To(BeEmpty())
		})
	})

	Context("delete non-existing resources", func() {
		It("should fail to delete a non-existing device", func() {
			// given
			rootCmd.SetArgs([]string{"delete", "device", "--name", deviceName})

			// when
			err := rootCmd.Execute()
			Expect(err).To(HaveOccurred())

			// then
			Expect(actualErr.String()).To(Equal(fmt.Sprintf("Error: edgedevices.management.project-flotta.io \"%s\" not found\n", deviceName)))
		})

		It("should fail to delete a non-existing workload", func() {
			// given
			rootCmd.SetArgs([]string{"delete", "workload", "--name", workloadName})

			// when
			err := rootCmd.Execute()
			Expect(err).To(HaveOccurred())

			// then
			Expect(actualErr.String()).To(Equal(fmt.Sprintf("Error: edgeworkloads.management.project-flotta.io \"%s\" not found\n", workloadName)))
		})

		It("should fail to delete a non-existing device-set", func() {
			// given
			rootCmd.SetArgs([]string{"delete", "deviceset", "--name", setName})

			// when
			err := rootCmd.Execute()
			Expect(err).To(HaveOccurred())

			// then
			Expect(actualErr.String()).To(Equal(fmt.Sprintf("Error: edgedevicesets.management.project-flotta.io \"%s\" not found\n", setName)))
		})
	})
})

func initializeResources() {
	client, err := resources.NewClient()
	Expect(err).NotTo(HaveOccurred())

	// initialize device
	device, err = resources.NewEdgeDevice(client, deviceName)
	Expect(err).NotTo(HaveOccurred())
	err = device.Register("")
	Expect(err).NotTo(HaveOccurred())

	// initialize workload
	workload, err = resources.NewEdgeWorkload(client)
	Expect(err).NotTo(HaveOccurred())
	_, err = workload.Create(resources.EdgeworkloadDeviceId(workloadName, deviceName, defaultImage))
	Expect(err).NotTo(HaveOccurred())
	err = device.WaitForWorkloadState(workloadName, "Running")
	Expect(err).NotTo(HaveOccurred())

	// initialize device set
	deviceSet, err = resources.NewEdgeDeviceSet(client, setName)
	Expect(err).NotTo(HaveOccurred())
	_, err = deviceSet.Create(resources.EdgeDeviceSetConfig(setName))
	Expect(err).NotTo(HaveOccurred())
}

func initializeSetWithDevices() {
	client, err := resources.NewClient()
	Expect(err).NotTo(HaveOccurred())

	// initialize device set
	deviceSet, err = resources.NewEdgeDeviceSet(client, setName)
	Expect(err).NotTo(HaveOccurred())
	_, err = deviceSet.Create(resources.EdgeDeviceSetConfig(setName))
	Expect(err).NotTo(HaveOccurred())

	// add the devices of the device set
	devices := []string{"dvc1", "dvc2"}
	for _, deviceName := range devices {
		device, err = resources.NewEdgeDevice(client, deviceName)
		Expect(err).NotTo(HaveOccurred())
		err = device.Register("")
		Expect(err).NotTo(HaveOccurred())

		dvc, err := device.Get()
		Expect(err).NotTo(HaveOccurred())
		dvc.Labels["flotta/member-of"] = setName
		_, err = client.EdgeDevices("default").Update(context.TODO(), dvc, v1.UpdateOptions{})
		Expect(err).NotTo(HaveOccurred())
	}
}
