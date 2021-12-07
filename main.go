package main

import (
	"context"
	"encoding/json"
	"flag"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

func main() {
	kubeConfig := flag.String("kubeconfig", "", "path to kubeconfig")
	podName := flag.String("pod", "", "pod to add ephemeral libguestfs container")
	ns := "default"
	flag.Parse()
	config, _ := clientcmd.BuildConfigFromFlags("", *kubeConfig)
	// creates the clientset
	clientset, _ := kubernetes.NewForConfig(config)
	// access the API to list pods
	pod, err := clientset.CoreV1().Pods("default").Get(context.TODO(), *podName, v1.GetOptions{})
	if err != nil {
		panic(err)
	}
	podJS, err := json.Marshal(pod)
	if err != nil {
		panic(err)
	}
	debugContainer := corev1.EphemeralContainer{
		EphemeralContainerCommon: corev1.EphemeralContainerCommon{
			Name:    "libguestfs",
			Image:   "registry:5000/kubevirt/libguestfs-tools:devel",
			Command: []string{"/entrypoint.sh"},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "libvirt-runtime",
					ReadOnly:  false,
					MountPath: "/var/run/libvirt",
				},
			},
			Stdin: true,
			TTY:   true,
		},
	}
	debugPod := pod.DeepCopy()
	debugPod.Spec.EphemeralContainers = append(debugPod.Spec.EphemeralContainers, debugContainer)
	debugJS, err := json.Marshal(debugPod)
	if err != nil {
		panic(err)
	}
	patch, err := strategicpatch.CreateTwoWayMergePatch(podJS, debugJS, pod)
	if err != nil {
		panic(err)
	}
	_, err = clientset.CoreV1().Pods(ns).Patch(context.TODO(), pod.Name, types.StrategicMergePatchType, patch, metav1.PatchOptions{}, "ephemeralcontainers")
	if err != nil {
		panic(err)
	}

	klog.Infof("Succesfully added libguestfs container")
}