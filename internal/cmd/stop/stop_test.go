package stop_test

import (
	"bytes"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/project-flotta/flotta-dev-cli/internal/resources"
	"github.com/spf13/cobra"

	. "github.com/project-flotta/flotta-dev-cli/internal/cmd"
	. "github.com/project-flotta/flotta-dev-cli/internal/cmd/stop"
)

var device resources.EdgeDevice

const deviceName = "device1"

var _ = Describe("Stop", func() {
	var (
		actualOut *bytes.Buffer
		actualErr *bytes.Buffer
		rootCmd   *cobra.Command
		stopCmd   *cobra.Command
	)

	BeforeEach(func() {
		// create a new buffer to capture stdout and stderr
		actualOut = new(bytes.Buffer)
		actualErr = new(bytes.Buffer)

		// create a new root command with the buffers as stdout and stderr
		rootCmd = NewRootCmd()
		rootCmd.SetOut(actualOut)
		rootCmd.SetErr(actualErr)

		stopCmd = NewStopCmd()
		rootCmd.AddCommand(stopCmd)
		stopCmd.AddCommand(NewDeviceCmd())
	})

	Context("Sanity", Ordered, func() {
		BeforeAll(func() {
			client, err := resources.NewClient()
			Expect(err).NotTo(HaveOccurred())

			device, err = resources.NewEdgeDevice(client, deviceName)
			Expect(err).NotTo(HaveOccurred())

			err = device.Register()
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			actualOut.Reset()
			actualErr.Reset()
		})

		AfterAll(func() {
			err := device.Remove()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should stop a device", func() {
			// given
			rootCmd.SetArgs([]string{"stop", "device", "-n", deviceName})

			// when
			err := rootCmd.Execute()
			Expect(err).NotTo(HaveOccurred())

			// then
			Expect(actualOut.String()).To(Equal(fmt.Sprintf("device '%s' was stopped \n", deviceName)))
		})
	})

	Context("stop non-existing resources", func() {
		It("should fail to stop a non-existing device", func() {
			// given
			rootCmd.SetArgs([]string{"stop", "device", "-n", deviceName})

			// when
			err := rootCmd.Execute()
			Expect(err).To(HaveOccurred())

			// then
			Expect(actualErr.String()).To(ContainSubstring(fmt.Sprintf("No such container: %s", deviceName)))
		})
	})

	Context("Required flags", func() {
		It("should fail to stop a device without a name", func() {
			// given
			rootCmd.SetArgs([]string{"stop", "device"})

			// when
			err := rootCmd.Execute()
			Expect(err).NotTo(BeNil())

			// then
			Expect(actualErr.String()).To(Equal("Error: required flag(s) \"name\" not set\n"))
		})
	})
})
