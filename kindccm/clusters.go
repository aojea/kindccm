package kindccm

import (
	"k8s.io/client-go/kubernetes"
	cloudprovider "k8s.io/cloud-provider"
)

type KindccmCluster struct {
	kubeClient *kubernetes.Clientset
}

var _ cloudprovider.Clusters = &KindccmCluster{}

func NewKindccmCluster(kubeClient *kubernetes.Clientset) cloudprovider.Clusters {
	return &KindccmCluster{kubeClient}
}
