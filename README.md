# KINDCCM: KIND Cloud Controller Manager

kindccm is an out-of-tree Kubernetes cloud provider implementation.

## Getting started

To use the cloud provider, we'll need to do a few things:

Install kindccm

Set --cloud-provider=external on our kube-controller-manager master component

Deploy keepalived-cloud-provider

Create a service with type: LoadBalancer