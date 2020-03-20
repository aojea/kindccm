package kindccm

import (
	"fmt"
	"net"

	"github.com/golang/glog"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/kubernetes/pkg/cloudprovider"
)

type KindccmLoadBalancer struct {
	kubeClient      *kubernetes.Clientset
	namespace, name string
	serviceCidr     string
	forwardMethod   string
}

var _ cloudprovider.LoadBalancer = &KindccmLoadBalancer{}

func NewKindccmLoadBalancer(kubeClient *kubernetes.Clientset) cloudprovider.LoadBalancer {
	return &KindccmLoadBalancer{kubeClient}
}

func (k *KindccmLoadBalancer) GetLoadBalancer(clusterName string, service *v1.Service) (status *v1.LoadBalancerStatus, exists bool, err error) {
	cm, err := k.getConfigMap()

	if err != nil {
		return nil, false, err
	}

	cfg, err := configFrom(cm)

	if err != nil {
		return nil, false, err
	}

	for _, svc := range cfg.Services {
		if svc.UID == string(service.UID) {
			return &v1.LoadBalancerStatus{
				Ingress: []v1.LoadBalancerIngress{{IP: svc.IP}},
			}, true, nil
		}
	}

	return nil, false, nil
}

func (k *KindccmLoadBalancer) EnsureLoadBalancer(clusterName string, service *v1.Service, nodes []*v1.Node) (*v1.LoadBalancerStatus, error) {
	return k.syncLoadBalancer(service)
}

func (k *KindccmLoadBalancer) UpdateLoadBalancer(clusterName string, service *v1.Service, nodes []*v1.Node) error {
	_, err := k.syncLoadBalancer(service)
	return err
}

func (k *KindccmLoadBalancer) EnsureLoadBalancerDeleted(clusterName string, service *v1.Service) error {
	return k.deleteLoadBalancer(service)
}

func (k *KindccmLoadBalancer) deleteLoadBalancer(service *v1.Service) error {
	glog.Infof("ensure service '%s' (%s) is deleted", service.Name, service.UID)

	cm, err := k.getConfigMap()

	if err != nil {
		return err
	}

	cfg, err := configFrom(cm)

	if err != nil {
		return err
	}

	for _, svc := range cfg.Services {
		// service already exists in the config so just return the status
		if svc.UID == string(service.UID) {
			glog.Infof("found service '%s' (%s) for deletion (%s)", service.Name, service.UID, svc.IP)
			cfg.deleteService(svc)
			delete(cm.Data, svc.IP)

			cfgBytes, err := cfg.encode()

			if err != nil {
				return fmt.Errorf("error encoding updated config: %s", err.Error())
			}

			cm.Annotations[configMapAnnotationKey] = string(cfgBytes)

			glog.Infof("update configmap config annotation: %s", string(cfgBytes))
			if _, err = k.kubeClient.ConfigMaps(k.namespace).Update(cm); err != nil {
				return fmt.Errorf("error updating kindccm config: %s", err.Error())
			}

			glog.Infof("updated configmap")
			return nil
		}
	}

	return nil
}

func (k *KindccmLoadBalancer) syncLoadBalancer(service *v1.Service) (*v1.LoadBalancerStatus, error) {
	glog.Infof("syncing service '%s' (%s)", service.Name, service.UID)

	cm, err := k.getConfigMap()

	if err != nil {
		return nil, err
	}

	cfg, err := configFrom(cm)

	if err != nil {
		return nil, err
	}

	forwardMethod := k.forwardMethod
	reallocateIP := true
	var svc serviceConfig
	for _, svc = range cfg.Services {
		// service already exists in the config so just return the status
		if svc.UID == string(service.UID) {
			glog.Infof("found existing loadbalancer for service '%s' (%s) with IP: %s", service.Name, service.UID, svc.IP)
			// if there's a mismatch between desired loadBalancerIP and actual,
			// break out of this loop and continue to update
			if service.Spec.LoadBalancerIP != "" && service.Spec.LoadBalancerIP != svc.IP {
				break
			}

			if annotationForwardMethod, ok := service.Annotations[serviceForwardMethodAnnotationKey]; ok {
				forwardMethod = annotationForwardMethod
			}
			if forwardMethod != svc.ForwardMethod {
				reallocateIP = false
				break
			}

			return &v1.LoadBalancerStatus{
				Ingress: []v1.LoadBalancerIngress{{IP: svc.IP}},
			}, nil
		}
	}

	ip := svc.IP
	if lbip := service.Spec.LoadBalancerIP; lbip != "" {
		if i := net.ParseIP(lbip); i == nil {
			return nil, fmt.Errorf("invalid loadBalancerIP specified '%s': %s", lbip, err.Error())
		}
		ip = lbip
	} else if reallocateIP {
		ip, err = cfg.allocateIP(k.serviceCidr)
		if err != nil {
			return nil, err
		}
	}

	sc := serviceConfig{
		UID:              string(service.UID),
		IP:               ip,
		ServiceNamespace: service.Namespace,
		ServiceName:      service.Name,
		ForwardMethod:    forwardMethod,
	}
	cfg.ensureService(sc)
	cfgBytes, err := cfg.encode()

	if err != nil {
		return nil, fmt.Errorf("error encoding updated config: %s", err.Error())
	}

	cm.Data = cfg.toConfigMapData()
	if forwardMethod != "" {
		cm.Data[ip] = service.Namespace + "/" + service.Name + ":" + forwardMethod
	} else {
		cm.Data[ip] = service.Namespace + "/" + service.Name
	}
	cm.Annotations[configMapAnnotationKey] = string(cfgBytes)

	glog.Infof("update configmap config annotation: %s", string(cfgBytes))
	if _, err = k.kubeClient.ConfigMaps(k.namespace).Update(cm); err != nil {
		return nil, fmt.Errorf("error updating kindccm config: %s", err.Error())
	}

	glog.Infof("synced service '%s' (%s): %s", service.Name, service.UID, ip)

	return &v1.LoadBalancerStatus{
		Ingress: []v1.LoadBalancerIngress{{IP: ip}},
	}, nil
}

func (k *KindccmLoadBalancer) getConfigMap() (*apiv1.ConfigMap, error) {
	cm, err := k.kubeClient.ConfigMaps(k.namespace).Get(k.name, metav1.GetOptions{})

	if err != nil {
		return nil, fmt.Errorf("error getting kindccm configmap: %s", err.Error())
	}

	if cm.Data == nil {
		cm.Data = map[string]string{}
	}

	if cm.Annotations == nil {
		cm.Annotations = map[string]string{}
	}

	return cm, err
}
