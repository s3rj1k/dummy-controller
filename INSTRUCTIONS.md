# How to test dummy-controller

## Setting up local KinD cluster.

Before creating local KinD cluster, please ensure that Docker config is updated to handle DNS (see example below), on some Linux distributions Docker DNS redirection does not work out of the box, restart Docker daemon to apply new config.

#### Docker daemon config (working DNS in KinD cluster).
```
cat <<EOF > /etc/docker/daemon.json
{
 "default-address-pools":
 [
   {"base": "10.42.0.0/16", "size": 24}
 ],
 "dns": ["1.1.1.1", "8.8.8.8"]
}
EOF
```

#### Create cluster.
```
kind create cluster --wait 60s --name=sandbox
kubectl config use-context kind-sandbox
kubectl cluster-info --context kind-sandbox
```

## Apply charts to Kubernetes cluster.

There is no specific need to do this in KinD, any cluster would work, as long as `kubectl` works from terminal, meaning that there is a cluster config present in home directory.

#### Deploy to cluster.
Run the below command in CWD of clonned git repository, it will apply all needed charts (that will pull needed images) into Kubernetes cluster including `Namespace` creation, everything related to the controller will be in `dummy-controller-system` namespace, this is done specificaly to allow Operator SDK generated assets to work as expected and to allow proper cleanup.
```
make deploy
```

The controller image itself is avaliable at: `docker pull rg.nl-ams.scw.cloud/s3rj1k-hub/dummy-controller:latest`.

#### Apply CR to cluster.

Note that the apiVersion is `homework.interview.me/v1alpha1`. 

```
cat <<EOF > /tmp/dummy.yaml
apiVersion: homework.interview.me/v1alpha1
kind: Dummy
metadata:
  name: dummy1
spec:
  message: "I'm just a dummy"
EOF
```

```
kubectl -n dummy-controller-system apply -f /tmp/dummy.yaml
```

#### Check related resources.

```
kubectl -n dummy-controller-system get dummy dummy1 -o yaml -w
```

```
kubectl -n dummy-controller-system get pods
```

#### Delete resource and verify that bound Pod is also gets deleted.

```
kubectl -n dummy-controller-system delete dummy dummy1
```

```
kubectl -n dummy-controller-system get pods -w
```

## Remove deployed charts from cluster.
```
make undeploy
```

## Remove local KinD cluster.

```
kind delete clusters sandbox
```

## Cheetsheet. 

### Init controller.
```
docker run --rm -it -v $(pwd):/usr/src/$(basename $(pwd)) -w /usr/src/$(basename $(pwd)) rg.nl-ams.scw.cloud/s3rj1k-hub/operator-sdk:latest init --domain interview.com --repo github.com/s3rj1k/dummy-controller
```

### Create API.
```
docker run --rm -it -v $(pwd):/usr/src/$(basename $(pwd)) -w /usr/src/$(basename $(pwd)) rg.nl-ams.scw.cloud/s3rj1k-hub/operator-sdk:latest create api --group homework --version v1alpha1 --kind Dummy --resource --controller
```

### Run Golang release that is supported by Operator SDK.
```
docker run --rm -it -v $(pwd):/usr/src/$(basename $(pwd)) -w /usr/src/$(basename $(pwd)) golang:1.21 bash
```

### Build and Push image.
```
make docker-build docker-push
```
