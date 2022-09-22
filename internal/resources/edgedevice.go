package resources

import (
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/pkg/archive"
	"io"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/project-flotta/flotta-operator/api/v1alpha1"
	"github.com/project-flotta/flotta-operator/generated/clientset/versioned"
	mgmtv1alpha1 "github.com/project-flotta/flotta-operator/generated/clientset/versioned/typed/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	certsPath      = "/etc/pki/consumer"
	CACertsPath    = filepath.Join(certsPath, "ca.pem")
	clientCertPath = filepath.Join(certsPath, "cert.pem")
	clientKeyPath  = filepath.Join(certsPath, "key.pem")
	certificates   = []string{CACertsPath, clientKeyPath, clientCertPath}
)

const (
	EdgeDeviceImage string = "quay.io/project-flotta/edgedevice:cli"
	Namespace       string = "default"
	waitTimeout     int    = 120
	sleepInterval   int    = 2
)

type EdgeDevice interface {
	GetName() string
	Register(image string, cmds ...string) error
	Unregister() error
	Get() (*v1alpha1.EdgeDevice, error)
	List() (*v1alpha1.EdgeDeviceList, error)
	Remove() error
	Stop() error
	Start() error
	WaitForWorkloadState(string, v1alpha1.EdgeWorkloadPhase) error
}

type edgeDevice struct {
	device mgmtv1alpha1.ManagementV1alpha1Interface
	client *client.Client
	name   string
}

func NewEdgeDevice(fclient mgmtv1alpha1.ManagementV1alpha1Interface, deviceName string) (EdgeDevice, error) {
	client, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &edgeDevice{device: fclient, client: client, name: deviceName}, nil
}

func (e *edgeDevice) GetName() string {
	return e.name
}

func (e *edgeDevice) Register(image string, cmds ...string) error {
	if image == "" {
		image = EdgeDeviceImage
	}
	ctx := context.Background()
	out, err := e.client.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image '%s': %v", image, err)
	}
	defer out.Close()
	if _, err := io.ReadAll(out); err != nil {
		return err
	}

	resp, err := e.client.ContainerCreate(ctx, &container.Config{Image: image, Labels: map[string]string{"flotta": "true"}}, &container.HostConfig{Privileged: true, ExtraHosts: []string{"project-flotta.io:172.17.0.1"}}, nil, nil, e.name)
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

func (e *edgeDevice) List() (*v1alpha1.EdgeDeviceList, error) {
	return e.device.EdgeDevices(Namespace).List(context.TODO(), metav1.ListOptions{})
}

func (e *edgeDevice) Stop() error {
	timeout := time.Duration(waitTimeout)
	return e.client.ContainerStop(context.TODO(), e.name, &timeout)
}

func (e *edgeDevice) Start() error {
	return e.client.ContainerStart(context.TODO(), e.name, types.ContainerStartOptions{})
}

func (e *edgeDevice) Remove() error {
	return e.client.ContainerRemove(context.TODO(), e.name, types.ContainerRemoveOptions{Force: true})
}

func (e *edgeDevice) CopyCerts() error {
	dir, err := os.MkdirTemp("", "certs")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir) // clean up

	certsMap := make(map[string][]byte)
	err = CopyCertsToTempDir(dir, certsMap)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	for certificatePath := range certsMap {
		if _, err := os.Stat(certificatePath); errors.Is(err, os.ErrNotExist) {
			return err
		}
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

	if _, err := e.Exec(fmt.Sprintf("echo 'ca-root = [\"%v\"]' >> /etc/yggdrasil/config.toml", CACertsPath)); err != nil {
		return err
	}

	return nil
}

func CopyCertsToTempDir(dir string, certsMap map[string][]byte) error {
	// get secrets
	config, err := GetKubeConfig()
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	CASecret, err := clientset.CoreV1().Secrets("flotta").Get(context.TODO(), "flotta-ca", metav1.GetOptions{})
	if err != nil {
		return err
	}
	secrets, err := clientset.CoreV1().Secrets("flotta").List(context.TODO(), metav1.ListOptions{LabelSelector: "reg-client-ca=true"})
	if err != nil {
		return err
	}

	// sort secrets by creation time
	data := secrets.Items
	sort.Slice(data, func(i, j int) bool {
		return data[i].CreationTimestamp.Before(&data[j].CreationTimestamp)
	})
	regSecret := data[len(data)-1]

	// save secrets to files
	certsMap[filepath.Join(dir, "ca.pem")] = CASecret.Data["ca.crt"]
	certsMap[filepath.Join(dir, "cert.pem")] = regSecret.Data["client.crt"]
	certsMap[filepath.Join(dir, "key.pem")] = regSecret.Data["client.key"]

	for path, data := range certsMap {
		if err := os.WriteFile(path, data, 0666); err != nil {
			return err
		}
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

	data, err := io.ReadAll(response.Reader)
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

func (e *edgeDevice) WaitForWorkloadState(workloadName string, workloadPhase v1alpha1.EdgeWorkloadPhase) error {
	return e.waitForDevice(func() bool {
		device, err := e.Get()
		if device == nil || err != nil {
			return false
		}

		if len(device.Status.Workloads) == 0 {
			return false
		}
		workloads := device.Status.Workloads
		for _, workload := range workloads {
			if workload.Name == workloadName && workload.Phase == workloadPhase {
				return true
			}
		}
		return false
	})
}

func NewClient() (mgmtv1alpha1.ManagementV1alpha1Interface, error) {
	config, err := GetKubeConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset.ManagementV1alpha1(), nil
}

func GetKubeConfig() (*rest.Config, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	config, err := clientcmd.BuildConfigFromFlags("", path.Join(homedir, ".kube/config"))
	if err != nil {
		return nil, fmt.Errorf("cannot get kube config from path '%s': %v", path.Join(homedir, ".kube/config"), err)
	}
	return config, nil
}
