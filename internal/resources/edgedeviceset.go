package resources

import (
	"context"
	"fmt"
	"github.com/project-flotta/flotta-operator/api/v1alpha1"
	mgmtv1alpha1 "github.com/project-flotta/flotta-operator/generated/clientset/versioned/typed/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type EdgeDeviceSet interface {
	GetName() string
	Create(set *v1alpha1.EdgeDeviceSet) (*v1alpha1.EdgeDeviceSet, error)
	Get(string) (*v1alpha1.EdgeDeviceSet, error)
	Remove(string) error
	RemoveAll() error
	List() (*v1alpha1.EdgeDeviceSetList, error)
}

type edgeDeviceSet struct {
	deviceset mgmtv1alpha1.ManagementV1alpha1Interface
	name      string
}

func NewEdgeDeviceSet(client mgmtv1alpha1.ManagementV1alpha1Interface, deviceSetName string) (*edgeDeviceSet, error) {
	return &edgeDeviceSet{deviceset: client, name: deviceSetName}, nil
}

func (e *edgeDeviceSet) GetName() string {
	return e.name
}

func (e *edgeDeviceSet) Get(name string) (*v1alpha1.EdgeDeviceSet, error) {
	return e.deviceset.EdgeDeviceSets(Namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

func (e *edgeDeviceSet) List() (*v1alpha1.EdgeDeviceSetList, error) {
	return e.deviceset.EdgeDeviceSets(Namespace).List(context.TODO(), metav1.ListOptions{})
}

func (e *edgeDeviceSet) Create(eds *v1alpha1.EdgeDeviceSet) (*v1alpha1.EdgeDeviceSet, error) {
	return e.deviceset.EdgeDeviceSets(Namespace).Create(context.TODO(), eds, metav1.CreateOptions{})
}

func (e *edgeDeviceSet) RemoveAll() error {
	return e.deviceset.EdgeDeviceSets(Namespace).DeleteCollection(context.TODO(), metav1.DeleteOptions{}, metav1.ListOptions{})
}

func (e *edgeDeviceSet) Remove(name string) error {
	err := e.deviceset.EdgeDeviceSets(Namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return e.waitForDeviceSet(func() bool {
		if _, err := e.Get(name); err != nil {
			return true
		}
		return false
	})
}

func (e *edgeDeviceSet) waitForDeviceSet(cond func() bool) error {
	for i := 0; i <= waitTimeout; i += sleepInterval {
		if cond() {
			return nil
		} else {
			time.Sleep(time.Duration(sleepInterval) * time.Second)
		}
	}

	return fmt.Errorf("error waiting for edge device set")
}

func EdgeDeviceSetConfig(name string) *v1alpha1.EdgeDeviceSet {
	deviceset := &v1alpha1.EdgeDeviceSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.EdgeDeviceSetSpec{
			Heartbeat: &v1alpha1.HeartbeatConfiguration{
				PeriodSeconds: 5,
			},
			Metrics: &v1alpha1.MetricsConfiguration{
				SystemMetrics: &v1alpha1.SystemMetricsConfiguration{
					Interval: 600,
				},
			},
		},
	}

	return deviceset
}
