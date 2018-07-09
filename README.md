# Kubernetes Cloud Control Manager for 1and1

oneandone-cloud-controller-manager is the Kubernetes cloud controller manager implementation for 1and1. Read more about cloud controller managers [here](https://kubernetes.io/docs/tasks/administer-cluster/running-cloud-controller/). Running oneandone-cloud-controller-manager allows you to leverage the cloud provider features offered by 1and1 on your kubernetes clusters.

**WARNING**: this project is a work in progress.  Still TODO:

- better test coverage
- investigate startup warnings

## Setup and Installation

### Version

These are the recommended cloud controller manager versions based on your Kubernetes version:

| Kubernetes version | CCM version |
| ------------------ | ----------- |
| 1.9                | = 0.1.0     |
| 1.10               | = 0.1.0     |
| 1.11               | = 0.1.0     |

### Preparing Your Cluster

Node names or public IP addresses must match servers in your cloud panel.  If node names do not match, use one of these options:

- Add the label `stackpoint.io/instance_id` to **all nodes** in your cluster.  The label's value must be the server name.
- Set the `--hostname-override` flag on the `kubelet` on **all nodes** in your cluster.

Your cluster should be configured to use an external cloud-provider to take full advantage of this manager. Without this only the loadbalancing provisioning will work.

Enable the external cloud provider by setting the `--cloud-provider=external` flag on the `kubelet` on **all nodes** in your cluster.

**WARNING**: setting `--cloud-provider=external` will taint all nodes in a cluster with `node.cloudprovider.kubernetes.io/uninitialized`.  It is the responsibility of a cloud controller manager to untaint those nodes once it has finished initializing them. This means that most pods will be left unschedulable until the cloud controller manager is running.

**Depending on how kube-proxy is run you _may_ need the following:**

- Ensure that `kube-proxy` tolerates the uninitialised cloud taint. The
  following should appear in the `kube-proxy` pod yaml:

```yaml
- effect: NoSchedule
  key: node.cloudprovider.kubernetes.io/uninitialized
  value: "true"
```

If your cluster was created using `kubeadm` >= v1.7.2 this toleration *may*
already be applied. See [kubernetes/kubernetes#49017][5] for details.

**If you are running flannel, ensure that kube-flannel tolerates the uninitialised cloud taint:**

- The following should appear in the `kube-flannel` daemonset:

```yaml
- effect: NoSchedule
  key: node.cloudprovider.kubernetes.io/uninitialized
  value: "true"
```

Remember to restart any components that you have reconfigured before continuing.

### Authentication and Configuration

The 1&1 Cloud Controller Manager requires a cloud panel API token and the datacenter code stored in the following environment variables:

- ONEANDONE_API_KEY
- ONEANDONE_INSTANCE_REGION

The default manifest is configured to set these environment variables from a secret named `oneandone`:

kubectl -n kube-system create secret generic oneandone --from-literal=token=`<TOKEN>`
--from-literal=credentials-datacenter=GB

### Installation - with RBAC

`kubectl apply -f manifests/oneandone-ccm-rbac.yaml`

`kubectl apply -f manifests/oneandone-ccm.yaml`

### Installation - without RBAC

`kubectl apply -f manifests/oneandone-ccm.yaml`

## Testing
### End-To-End Tests

*NOTE*: the end-to-end tests create cloud panel resources which you may be billed for.  You will need an API key for cloud panel.  The resources created are:

 - 1 x L virtual server
 - 2 x M virtual servers
 - 1 x firewall
 - 1 x private network

If you have ansible, terraform and kubectl installed you can run the tests directly:

```
export ONEANDONE_API_KEY=xxx
go test -c ./test/e2e
./e2e.test -test.v -test.timeout 30m -kubever v1.10.5
```

If you need to clean up any left over cloud panel resources, you can use terraform directly:
```
cd ./test/e2e/terraform
terraform destroy -var provider_token=xxx
```

There is also a Dockerfile which can be used to build an image capable of running the tests:

```
docker build -t ccme2e -f Dockerfile-e2e .
docker run -e ONEANDONE_API_KEY=xxx -e K8S_VERSION=v1.10.5 --rm ccme2e
```

#### TODO

The end-to-end tests need to be expanded to include:

 - Joining a node after the cluster is up and running: the new node should be initialised
 - Deleting a node in cloud panel: the node should be removed from the cluster
 - Joining a node: the new node should be added to all loadbalancers
 - Updating a service: the loadbalancer should be updated
 - Deleting a service: the loadbalancer should be deleted
 - Parameterise CCM version so different CCMs can be e2e tested

