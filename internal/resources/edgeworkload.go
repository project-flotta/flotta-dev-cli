package resources

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/project-flotta/flotta-operator/api/v1alpha1"
	mgmtv1alpha1 "github.com/project-flotta/flotta-operator/generated/clientset/versioned/typed/v1alpha1"
)

type EdgeWorkload interface {
	Create(*v1alpha1.EdgeWorkload) (*v1alpha1.EdgeWorkload, error)
	Get(string) (*v1alpha1.EdgeWorkload, error)
	Remove(string) error
	RemoveAll() error
}

type edgeWorkload struct {
	workload mgmtv1alpha1.ManagementV1alpha1Interface
}

func NewEdgeWorkload(client mgmtv1alpha1.ManagementV1alpha1Interface) (*edgeWorkload, error) {
	return &edgeWorkload{workload: client}, nil
}

func (e *edgeWorkload) Get(name string) (*v1alpha1.EdgeWorkload, error) {
	return e.workload.EdgeWorkloads(Namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

func (e *edgeWorkload) Create(ew *v1alpha1.EdgeWorkload) (*v1alpha1.EdgeWorkload, error) {
	return e.workload.EdgeWorkloads(Namespace).Create(context.TODO(), ew, metav1.CreateOptions{})
}

func (e *edgeWorkload) RemoveAll() error {
	return e.workload.EdgeWorkloads(Namespace).DeleteCollection(context.TODO(), metav1.DeleteOptions{}, metav1.ListOptions{})
}

func (e *edgeWorkload) Remove(name string) error {
	err := e.workload.EdgeWorkloads(Namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return e.waitForWorkload(func() bool {
		if _, err := e.Get(name); err != nil {
			return true
		}
		return false
	})
}

func (e *edgeWorkload) waitForWorkload(cond func() bool) error {
	for i := 0; i <= waitTimeout; i += sleepInterval {
		if cond() {
			return nil
		} else {
			time.Sleep(time.Duration(sleepInterval) * time.Second)
		}
	}

	return fmt.Errorf("error waiting for edgeworkload")
}

func edgeworkloadDeviceIdCtrName(name string, ctrName string, device string, image string) *v1alpha1.EdgeWorkload {
	workload := edgeworkload(name, ctrName, nil, nil, image)
	workload.Spec.Device = device
	return workload
}

func EdgeworkloadDeviceId(name string, device string, image string) *v1alpha1.EdgeWorkload {
	return edgeworkloadDeviceIdCtrName(name, name, device, image)
}

func edgeworkload(name string, ctrName string, secretRef *string, configRef *string, image string) *v1alpha1.EdgeWorkload {
	return edgeworkloadContainers(name, ctrName, secretRef, configRef, 1, image)
}

func edgeworkloadContainers(name string, ctrName string, secretRef *string, configRef *string, ctrCount int, image string) *v1alpha1.EdgeWorkload {
	envFrom := make([]corev1.EnvFromSource, 0)
	if secretRef != nil {
		envFrom = append(envFrom, corev1.EnvFromSource{SecretRef: &corev1.SecretEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{Name: *secretRef},
		}})
	}
	if configRef != nil {
		envFrom = append(envFrom, corev1.EnvFromSource{ConfigMapRef: &corev1.ConfigMapEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{Name: *configRef},
		}})
	}

	var containers = make([]corev1.Container, 0)
	for i := 0; i < ctrCount; i++ {
		if i > 0 {
			ctrName = fmt.Sprintf("%s_%d", name, i)
		}
		containers = append(containers, corev1.Container{
			Name:    ctrName,
			Image:   image,
			EnvFrom: envFrom,
		})
	}

	workload := &v1alpha1.EdgeWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.EdgeWorkloadSpec{
			Type: "pod",
			Pod: v1alpha1.Pod{
				Spec: corev1.PodSpec{
					Containers: containers,
				},
			},
		},
	}

	return workload
}
