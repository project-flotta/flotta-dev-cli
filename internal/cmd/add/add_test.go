package add_test

import (
	"bytes"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/project-flotta/flotta-dev-cli/internal/resources"
	"github.com/spf13/cobra"

	. "github.com/project-flotta/flotta-dev-cli/internal/cmd"
	. "github.com/project-flotta/flotta-dev-cli/internal/cmd/add"
)

var _ = Describe("Add", func() {
	var (
		actualOut *bytes.Buffer
		actualErr *bytes.Buffer
		rootCmd   *cobra.Command
		addCmd    *cobra.Command
	)

	const (
		deviceName   = "device1"
		workloadName = "workload1"
		setName      = "set1"
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

			addCmd = NewAddCmd()
			rootCmd.AddCommand(addCmd)
			addCmd.AddCommand(NewDeviceCmd(), NewWorkloadCmd(), NewDeviceSetCmd())
		})

		AfterEach(func() {
			actualOut.Reset()
			actualErr.Reset()
		})

		AfterAll(func() {
			client, err := resources.NewClient()
			Expect(err).NotTo(HaveOccurred())

			// remove test device and workload
			device, err := resources.NewEdgeDevice(client, deviceName)
			Expect(err).NotTo(HaveOccurred())
			dvc, err := device.Get()
			Expect(err).NotTo(HaveOccurred())

			workloads := dvc.Status.Workloads
			for _, w := range workloads {
				workload, err := resources.NewEdgeWorkload(client)
				Expect(err).NotTo(HaveOccurred())
				err = workload.Remove(w.Name)
				Expect(err).NotTo(HaveOccurred())
			}

			err = device.Remove()
			Expect(err).NotTo(HaveOccurred())

			// remove test device-set
			deviceset, err := resources.NewEdgeDeviceSet(client, setName)
			Expect(err).NotTo(HaveOccurred())
			err = deviceset.Remove(setName)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should add a new device", func() {
			// given
			rootCmd.SetArgs([]string{"add", "device", "-n", deviceName})

			// when
			err := rootCmd.Execute()
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(actualOut.String()).To(Equal(fmt.Sprintf("device '%s' was added\n", deviceName)))
			Expect(actualErr.String()).To(BeEmpty())
		})

		It("should add a new workload", func() {
			// given
			rootCmd.SetArgs([]string{"add", "workload", "-d", deviceName, "-n", workloadName})

			// when
			err := rootCmd.Execute()
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(actualOut.String()).To(Equal(fmt.Sprintf("workload '%s' was added to device '%s'\n", workloadName, deviceName)))
			Expect(actualErr.String()).To(BeEmpty())
		})

		It("should add a new device-set", func() {
			// given
			rootCmd.SetArgs([]string{"add", "deviceset", "-n", setName})

			// when
			err := rootCmd.Execute()
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(actualOut.String()).To(Equal(fmt.Sprintf("device-set '%s' was added\n", setName)))
			Expect(actualErr.String()).To(BeEmpty())
		})
	})

	Context("Required flags", Ordered, func() {
		BeforeAll(func() {
			// create a new buffer to capture stdout and stderr
			actualOut = new(bytes.Buffer)
			actualErr = new(bytes.Buffer)

			// create a new root command with the buffers as stdout and stderr
			rootCmd = NewRootCmd()
			rootCmd.SetOut(actualOut)
			rootCmd.SetErr(actualErr)

			addCmd = NewAddCmd()
			rootCmd.AddCommand(addCmd)
			addCmd.AddCommand(NewDeviceCmd(), NewWorkloadCmd(), NewDeviceSetCmd())
		})

		AfterEach(func() {
			actualOut.Reset()
			actualErr.Reset()
		})

		It("should fail to add a new device without a name", func() {
			// given
			rootCmd.SetArgs([]string{"add", "device"})

			// when
			err := rootCmd.Execute()
			Expect(err).To(HaveOccurred())

			// then
			Expect(actualErr.String()).To(Equal("Error: required flag(s) \"name\" not set\n"))
		})

		It("should fail to add a new workload without a device", func() {
			// given
			rootCmd.SetArgs([]string{"add", "workload"})

			// when
			err := rootCmd.Execute()
			Expect(err).To(HaveOccurred())

			// then
			Expect(actualErr.String()).To(Equal("Error: required flag(s) \"device\" not set\n"))
		})

		It("should fail to add a new device-set without a name", func() {
			// given
			rootCmd.SetArgs([]string{"add", "deviceset"})

			// when
			err := rootCmd.Execute()
			Expect(err).To(HaveOccurred())

			// then
			Expect(actualErr.String()).To(Equal("Error: required flag(s) \"name\" not set\n"))
		})
	})
})
