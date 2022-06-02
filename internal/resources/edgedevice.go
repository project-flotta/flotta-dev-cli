package resources

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/project-flotta/flotta-operator/api/v1alpha1"
	"github.com/project-flotta/flotta-operator/generated/clientset/versioned"
	managementv1alpha1 "github.com/project-flotta/flotta-operator/generated/clientset/versioned/typed/v1alpha1"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

var certsPath = "/etc/pki/consumer"
var CAcertsPath = filepath.Join(certsPath, "ca.pem")
var ClientCertPath = filepath.Join(certsPath, "cert.pem")
var ClientKeyPath = filepath.Join(certsPath, "key.pem")
var certificates = []string{CAcertsPath, ClientKeyPath, ClientCertPath}
var localCertificates = []string{
	"/tmp/ca.pem",
	"/tmp/cert.pem",
	"/tmp/key.pem",
}

const (
	EdgeDeviceImage string = "quay.io/project-flotta/edgedevice:latest"
	Namespace       string = "default"
	waitTimeout     int    = 120
	sleepInterval   int    = 2
)

type EdgeDevice interface {
	GetName() string
	Register(cmds ...string) error
	Unregister() error
	Get() (*v1alpha1.EdgeDevice, error)
	Remove() error
	Stop() error
}

type edgeDevice struct {
	device    managementv1alpha1.ManagementV1alpha1Interface
	client       *client.Client
	name      string
}

func NewEdgeDevice(fclient managementv1alpha1.ManagementV1alpha1Interface, deviceName string) (EdgeDevice, error) {
	client, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &edgeDevice{device: fclient, client: client, name: deviceName}, nil
}

func (e *edgeDevice) GetName() string {
	return e.name
}

func (e *edgeDevice) Register(cmds ...string) error {
	image := EdgeDeviceImage
	if name, exists := os.LookupEnv("DEVICE_IMAGE"); exists {
		image = name
	}
	ctx := context.Background()
	resp, err := e.client.ContainerCreate(ctx, &container.Config{Image: image}, &container.HostConfig{Privileged: true, ExtraHosts: []string{"project-flotta.io:172.17.0.1"}}, nil, nil, e.name)
	if err != nil {
		return err
	}

	if err := e.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	for _, cmd := range cmds {
		if _, err = e.Exec(cmd); err != nil {
			return fmt.Errorf("cannot execute register command '%s': %v", cmd, err)
		}
	}

	if _, err = e.Exec(fmt.Sprintf("echo 'client-id = \"%v\"' >> /etc/yggdrasil/config.toml", e.name)); err != nil {
		return err
	}

	if err := e.CopyCerts(); err != nil {
		return fmt.Errorf("cannot copy certificates to device: %v", err)
	}

	if _, err = e.Exec("systemctl start podman"); err != nil {
		return err
	}

	if _, err = e.Exec("systemctl start yggdrasild.service"); err != nil {
		return err
	}

	return e.waitForDevice(func() bool {
		device, err := e.Get()

		if err != nil || device == nil {
			return false
		}

		if _, ok := device.ObjectMeta.Labels["edgedeviceSignedRequest"]; ok {
			// Is not yet fully registered
			return false
		}

		if device.Status.Hardware == nil {
			return false
		}

		return true
	})
}

func (e *edgeDevice) Unregister() error {
	err := e.device.EdgeDevices(Namespace).Delete(context.TODO(), e.name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return e.waitForDevice(func() bool {
		if eCr, err := e.Get(); eCr == nil && err != nil {
			return true
		}
		return false
	})
}

func (e *edgeDevice) Get() (*v1alpha1.EdgeDevice, error) {
	device, err := e.device.EdgeDevices(Namespace).Get(context.TODO(), e.name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return device, nil
}

func (e *edgeDevice) Stop() error {
	timeout := time.Duration(waitTimeout)
	return e.client.ContainerStop(context.TODO(), e.name, &timeout)
}

func (e *edgeDevice) Remove() error {
	return e.client.ContainerRemove(context.TODO(), e.name, types.ContainerRemoveOptions{Force: true})
}

func (e *edgeDevice) CopyCerts() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	for _, certificatePath := range localCertificates {
		fp, err := archive.Tar(certificatePath, archive.Gzip)
		if err != nil {
			return err
		}
		err = e.client.CopyToContainer(ctx, e.name, certsPath, fp, types.CopyToContainerOptions{AllowOverwriteDirWithFile: true})
		if err != nil {
			return err
		}
	}

	for _, certificatePath := range certificates {
		if _, err := e.Exec(fmt.Sprintf("chmod 660 %s", certificatePath)); err != nil {
			return err
		}
	}

	if _, err := e.Exec(fmt.Sprintf("echo 'ca-root = [\"%v\"]' >> /etc/yggdrasil/config.toml", CAcertsPath)); err != nil {
		return err
	}

	return nil
}

func (e *edgeDevice) Exec(command string) (string, error) {
	resp, err := e.client.ContainerExecCreate(context.TODO(), e.name, types.ExecConfig{AttachStdout: true, AttachStderr: true, Cmd: []string{"/bin/bash", "-c", command}})
	if err != nil {
		return "", err
	}
	response, err := e.client.ContainerExecAttach(context.Background(), resp.ID, types.ExecStartCheck{})
	if err != nil {
		return "", err
	}
	defer response.Close()

	data, err := ioutil.ReadAll(response.Reader)
	if err != nil {
		return "", err
	}

	return strings.TrimFunc(string(data), func(r rune) bool {
		return !unicode.IsGraphic(r)
	}), nil
}

func (e *edgeDevice) waitForDevice(cond func() bool) error {
	for i := 0; i <= waitTimeout; i += sleepInterval {
		if cond() {
			return nil
		} else {
			time.Sleep(time.Duration(sleepInterval) * time.Second)
		}
	}

	return fmt.Errorf("error waiting for edgedevice %v[%v]", e.name, e.name)
}


func NewClient() (managementv1alpha1.ManagementV1alpha1Interface, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	config, err := clientcmd.BuildConfigFromFlags("", path.Join(homedir, ".kube/config"))
	if err != nil {
		return nil, err
	}
	clientset, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset.ManagementV1alpha1(), nil
}

func newClientset() (*kubernetes.Clientset, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	config, err := clientcmd.BuildConfigFromFlags("", path.Join(homedir, ".kube/config"))
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}
