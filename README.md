# Kubernetes Cluster API Provider Aliyun

## What is the Cluster API Provider Aliyun?

The [Cluster API](https://github.com/kubernetes-sigs/cluster-api) is a Kubernetes project to bring declarative, Kubernetes-style APIs to cluster creation, configuration, and management. It provides optional, additive functionality on top of core Kubernetes.

Provider Aliyun Implemented v1alpha2 of Cluster Api, make it easy to create and scale the native kubernetes cluster running on ECS .

## Features
- Native Kubernetes manifests and API.
- Manages the bootstrapping of VPCs, gateways, security groups and instances.
- Installs only the minimal components to bootstrap a control plane and workers.

## Launching a Kubernetes cluster on ECS

### Prerequisites
- Install and setup [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) in your local environment
- Install and/or configure a [management cluster](https://cluster-api.sigs.k8s.io/reference/glossary.html#management-cluster)

### Setup Management Cluster
Cluster API requires an existing kubernetes cluster accessible via kubectl. 

#### 1. Existing Management Cluster
```bash
export KUBECONFIG=...
```

#### 2. Kind
**Warning**: [kind](https://github.com/kubernetes-sigs/kind) is not designed for production use; it is intended for development and testing environments.

```
kind create cluster --name=clusterapi
export KUBECONFIG="$(kind get kubeconfig-path --name="clusterapi")"
```

### Quick Start
#### 1. Install CRDs
```
kubectl apply -f third_party/crd
```

#### 2. Install Providers
```
# 创建AccessKey https://help.aliyun.com/document_detail/53045.html

export ACCESS_KEY_ID=...
export ACCESS_SECRET=...

cat third_party/component/all-in-one.yaml | envsubst | kubectl apply -f -
```

#### 3. Create kubernetes cluster
```
kubectl apply -f example/cluster.yaml
kubectl apply -f example/controlplane.yaml
```

#### 4. Wait for controlplane running
```
kubectl get machine

NAME                               PROVIDERID                        PHASE
testcluster-controlplane-0         aliyun://i-8vbbykhgonzz6da7abas   running
testcluster-controlplane-1         aliyun://i-8vb28ugr0yzluc0h6vsa   running
testcluster-controlplane-2         aliyun://i-8vb2gdx8nsmeihx59aub   running
```

#### 5. Get kubeconfig of target kubernetes

```
# kubeconfig saved to $HOME/.kube/capal-config-testcluster
kubectl get secret testcluster-kubeconfig -o go-template="{{.data.value}}" | base64 -d > $HOME/.kube/capal-testcluster-config
```

#### 6. Install network addon & create nodes on target kubernetes
```
kubectl apply -f example/addon/flannel.yaml -n kube-system --kubeconfig=$HOME/.kube/capal-testcluster-config
kubectl apply -f example/machineset.yaml
```

#### 7. Check target cluster status
```
kubectl get nodes --kubeconfig=$HOME/.kube/capal-testcluster-config

NAME                      STATUS   ROLES    AGE     VERSION
iz8vb28ugr0yzluc0h6vsaz   Ready    master   7m     v1.14.4
iz8vb28ugr0yzm6i1c746nz   Ready    <none>   1m     v1.14.4
iz8vb2gdx8nsmeihx59aubz   Ready    master   4m     v1.14.4
iz8vb4efgq8tzq5v3ijediz   Ready    <none>   1m     v1.14.4
iz8vbbykhgonzz6da7abasz   Ready    master   4m     v1.14.4
iz8vbbykhgonzzila3elf1z   Ready    <none>   1m     v1.14.4
```


## Uninstall
```bash
kubectl delete -f example/machineset.yaml
kubectl delete -f example/controlplane.yaml
kubectl delete -f example/cluster.yaml
```


## Build
```bash
# Generate code
make generate

# Generate manifests e.g. CRD, RBAC etc.
make manifests

# Install CRDs into a cluster
make install

# Run against the configured Kubernetes cluster in ~/.kube/config
make run

# Build manager binary
make manager

# Build the docker image
make docker-build

# Push the docker image
make docker-push
```