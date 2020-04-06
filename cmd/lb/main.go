/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Note: the example only works with the code within the same release/branch.
package main

import (
	"context"
	"flag"
	"fmt"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	sharedInformers := informers.NewSharedInformerFactory(clientset, time.Minute)
	serviceInformer := NewServiceConfig(sharedInformers.Core().V1().Services(), time.Minute)

	serviceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			fmt.Println("AddFunc", obj)
		},
		UpdateFunc: func(old, cur interface{}) {
			fmt.Println("Update", old, cur)
		},
		DeleteFunc: func(obj interface{}) {
			fmt.Println("DeleteFunc", obj)
		},
	})
	go sharedInformers.Start(stopCh)
	go serviceConfig.Run(stopCh)

	for {
	}
}

func loopServices(clientset clientset.Interface) {
	for {
		svc, err := clientset.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Fatalf("Error getting services from kubernetes cluster: %s", err)
		}
		for _, svc := range svc.Items {
			if svc.Spec.Type == "LoadBalancer" && len(svc.Status.LoadBalancer.Ingress) == 0 {
				fmt.Println("Load Balancer", svc)
			} else {
				fmt.Println("No Load Balancer", svc)
			}
		}
		time.Sleep(10 * time.Second)
	}
}
func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
