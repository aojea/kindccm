package kindccm

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	cloudprovider "k8s.io/cloud-provider"
)

type KindccmZone struct {
	kubeClient *kubernetes.Clientset
}

var _ cloudprovider.Zones = &KindccmZone{}

func NewKindccmZone(kubeClient *kubernetes.Clientset) cloudprovider.Zones {
	return &KindccmZone{kubeClient}
}

// GetZone returns the Zone containing the current failure zone and locality region that the program is running in
// In most cases, this method is called from the kubelet querying a local metadata service to acquire its zone.
// For the case of external cloud providers, use GetZoneByProviderID or GetZoneByNodeName since GetZone
// can no longer be called from the kubelets.
func (k *KindccmZone) GetZone(ctx context.Context) (cloudprovider.Zone, error) {
	f.addCall("get-zone")
	return f.Zone, f.Err
}

// GetZoneByProviderID implements Zones.GetZoneByProviderID
// This is particularly useful in external cloud providers where the kubelet
// does not initialize node data.
func (k *KindccmZone) GetZoneByProviderID(ctx context.Context, providerID string) (cloudprovider.Zone, error) {
	f.addCall("get-zone-by-provider-id")
	return f.Zone, f.Err
}

// GetZoneByNodeName implements Zones.GetZoneByNodeName
// This is particularly useful in external cloud providers where the kubelet
// does not initialize node data.
func (k *KindccmZone) GetZoneByNodeName(ctx context.Context, nodeName types.NodeName) (cloudprovider.Zone, error) {
	f.addCall("get-zone-by-node-name")
	return f.Zone, f.Err
}
