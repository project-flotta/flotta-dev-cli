package list_test

import (
	"bytes"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/project-flotta/flotta-dev-cli/internal/resources"
	"github.com/spf13/cobra"

	. "github.com/project-flotta/flotta-dev-cli/internal/cmd"
	. "github.com/project-flotta/flotta-dev-cli/internal/cmd/list"
)

var (
	device    resources.EdgeDevice
	workload  resources.EdgeWorkload
	deviceSet resources.EdgeDeviceSet
)

const (
	deviceName   = "device1"
	workloadName = "workload1"
	setName      = "deviceset1"
)

var _ = Describe("List", func() {
	var (
		actualOut *bytes.Buffer
		actualErr *bytes.Buffer
		rootCmd   *cobra.Command
		listCmd   *cobra.Command
	)

	Context("Sanity", Ordered, func() {
		BeforeAll(func() {
			// create a new buffer to capture stdout and stderr
			actualOut = new(bytes.Buffer)
			actualErr = new(bytes.Buffer)

			// create a new root command with the buffers as stdout and stderr
			rootCmd = NewRootCmd()
			rootCmd.SetOut(actualOut)
			rootCmd.SetErr(actualErr)

			listCmd = NewListCmd()
			rootCmd.AddCommand(listCmd)
			listCmd.AddCommand(NewDeviceCmd(), NewWorkloadCmd(), NewDeviceSetCmd())

			initializeResources()
		})

		AfterEach(func() {
			actualOut.Reset()
			actualErr.Reset()
		})

		AfterAll(func() {
			removeResources()
		})

		It("should list devices", func() {
			// given
			rootCmd.SetArgs([]string{"list", "device"})

			// when
			err := rootCmd.Execute()
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(actualOut.String()).To(ContainSubstring("NAME\t\tSTATUS\t\tCREATED\t\t"))
			Expect(actualOut.String()).To(ContainSubstring(fmt.Sprintf("%s\t\trunning\t\t", deviceName)))
			Expect(actualErr.String()).To(BeEmpty())
		})

		It("should list workloads", func() {
			// given
			rootCmd.SetArgs([]string{"list", "workload"})

			// when
			err := rootCmd.Execute()
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(actualOut.String()).To(ContainSubstring("NAME\t\tSTATUS\t\tCREATED\t\t"))
			Expect(actualOut.String()).To(ContainSubstring(fmt.Sprintf("%s\tRunning\t", workloadName)))
			Expect(actualErr.String()).To(BeEmpty())
		})

		It("should list device sets", func() {
			// given
			rootCmd.SetArgs([]string{"list", "deviceset"})

			// when
			err := rootCmd.Execute()
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(actualOut.String()).To(ContainSubstring("NAME\t\tDEVICES\t\tCREATED"))
			Expect(actualOut.String()).To(ContainSubstring(setName))
			Expect(actualErr.String()).To(BeEmpty())
		})
	})
})

func initializeResources() {
	client, err := resources.NewClient()
	Expect(err).NotTo(HaveOccurred())

	// initialize device
	device, err = resources.NewEdgeDevice(client, deviceName)
	Expect(err).NotTo(HaveOccurred())
	err = device.Register()
	Expect(err).NotTo(HaveOccurred())

	// initialize workload
	workload, err = resources.NewEdgeWorkload(client)
	Expect(err).NotTo(HaveOccurred())
	_, err = workload.Create(resources.EdgeworkloadDeviceId(workloadName, deviceName, "quay.io/project-flotta/nginx:1.21.6"))
	Expect(err).NotTo(HaveOccurred())
	err = device.WaitForWorkloadState(workloadName, "Running")
	Expect(err).NotTo(HaveOccurred())

	// initialize device set
	deviceSet, err = resources.NewEdgeDeviceSet(client, setName)
	Expect(err).NotTo(HaveOccurred())
	_, err = deviceSet.Create(resources.EdgeDeviceSetConfig(setName))
	Expect(err).NotTo(HaveOccurred())
}

func removeResources() {
	client, err := resources.NewClient()
	Expect(err).NotTo(HaveOccurred())

	// remove test device
	device, err = resources.NewEdgeDevice(client, deviceName)
	Expect(err).NotTo(HaveOccurred())
	err = device.Remove()
	Expect(err).NotTo(HaveOccurred())

	// remove test workload
	workload, err = resources.NewEdgeWorkload(client)
	Expect(err).NotTo(HaveOccurred())
	err = workload.Remove(workloadName)
	Expect(err).NotTo(HaveOccurred())

	// remove test device set
	deviceSet, err = resources.NewEdgeDeviceSet(client, setName)
	Expect(err).NotTo(HaveOccurred())
	err = deviceSet.Remove(setName)
	Expect(err).NotTo(HaveOccurred())
}
