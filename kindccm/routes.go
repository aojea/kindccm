package kindccm

import (
	"context"
	"fmt"

	"k8s.io/client-go/kubernetes"

	cloudprovider "k8s.io/cloud-provider"
)

type KindccmRoute struct {
	kubeClient *kubernetes.Clientset
	RouteMap   map[string]string
}

var _ cloudprovider.Routes = &KindccmRoute{}

func NewKindccmRoute(kubeClient *kubernetes.Clientset) cloudprovider.Routes {
	return &KindccmRoute{kubeClient}
}

// CreateRoute creates the described managed route
// route.Name will be ignored, although the cloud-provider may use nameHint
// to create a more user-meaningful name.
func (k *KindccmRoute) CreateRoute(ctx context.Context, clusterName string, nameHint string, route *cloudprovider.Route) error {
	k.Lock.Lock()
	defer k.Lock.Unlock()
	k.addCall("create-route")
	name := clusterName + "-" + string(route.TargetNode) + "-" + route.DestinationCIDR
	if _, exists := k.RouteMap[name]; exists {
		k.Err = fmt.Errorf("route %q already exists", name)
		return k.Err
	}
	fakeRoute := Route{}
	fakeRoute.Route = *route
	fakeRoute.Route.Name = name
	fakeRoute.ClusterName = clusterName
	k.RouteMap[name] = &fakeRoute
	return nil
}

// DeleteRoute deletes the specified managed route
// Route should be as returned by ListRoutes
func (k *KindccmRoute) DeleteRoute(ctx context.Context, clusterName string, route *cloudprovider.Route) error {
	k.Lock.Lock()
	defer k.Lock.Unlock()
	k.addCall("delete-route")
	name := ""
	for key, saved := range k.RouteMap {
		if route.DestinationCIDR == saved.Route.DestinationCIDR &&
			route.TargetNode == saved.Route.TargetNode &&
			clusterName == saved.ClusterName {
			name = key
			break
		}
	}

	if len(name) == 0 {
		f.Err = fmt.Errorf("no route found for node:%v with DestinationCIDR== %v", route.TargetNode, route.DestinationCIDR)
		return f.Err
	}

	delete(k.RouteMap, name)
	return nil
}
