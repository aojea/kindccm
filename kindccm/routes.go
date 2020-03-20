package kindccm

import (
	"context"
	"fmt"

	"k8s.io/kubernetes/pkg/cloudprovider"
)

// CreateRoute creates the described managed route
// route.Name will be ignored, although the cloud-provider may use nameHint
// to create a more user-meaningful name.
func (f *Cloud) CreateRoute(ctx context.Context, clusterName string, nameHint string, route *cloudprovider.Route) error {
	f.Lock.Lock()
	defer f.Lock.Unlock()
	f.addCall("create-route")
	name := clusterName + "-" + string(route.TargetNode) + "-" + route.DestinationCIDR
	if _, exists := f.RouteMap[name]; exists {
		f.Err = fmt.Errorf("route %q already exists", name)
		return f.Err
	}
	fakeRoute := Route{}
	fakeRoute.Route = *route
	fakeRoute.Route.Name = name
	fakeRoute.ClusterName = clusterName
	f.RouteMap[name] = &fakeRoute
	return nil
}

// DeleteRoute deletes the specified managed route
// Route should be as returned by ListRoutes
func (f *Cloud) DeleteRoute(ctx context.Context, clusterName string, route *cloudprovider.Route) error {
	f.Lock.Lock()
	defer f.Lock.Unlock()
	f.addCall("delete-route")
	name := ""
	for key, saved := range f.RouteMap {
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

	delete(f.RouteMap, name)
	return nil
}
