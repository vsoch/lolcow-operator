# LolCow Operator

> Making my first Kubernetes operator!

I've always wanted to develop for Kubernetes, and I have my first opportunity and
wanted to give it a try! Per instruction of my friend and colleague [Eduardo](https://github.com/ArangoGutierrez/)
I am going to create a simple operator that takes a string argument for a controller,
and then you can update the string and the operator will print it. I'm
going to be marginally following [this guide](https://developers.redhat.com/articles/2021/09/07/build-kubernetes-operator-six-steps)
and also note [this is a nice example](https://github.com/kubernetes-sigs/kueue/blob/main/pkg/controller/core/queue_controller.go)
for a controller. This means that:

 - I have a recent version of Go installed (1.18.1)
 - I also have minikube installed
 
## Making the operator

### 1. Installation

I first [installed the operator-sdk](https://sdk.operatorframework.io/docs/installation/)

```bash
export ARCH=$(case $(uname -m) in x86_64) echo -n amd64 ;; aarch64) echo -n arm64 ;; *) echo -n $(uname -m) ;; esac)
export OS=$(uname | awk '{print tolower($0)}')
```
```bash
export OPERATOR_SDK_DL_URL=https://github.com/operator-framework/operator-sdk/releases/download/v1.22.2
curl -LO ${OPERATOR_SDK_DL_URL}/operator-sdk_${OS}_${ARCH}
```
```bash
gpg --keyserver keyserver.ubuntu.com --recv-keys 052996E2A20B5C7E
```
```bash
curl -LO ${OPERATOR_SDK_DL_URL}/checksums.txt
curl -LO ${OPERATOR_SDK_DL_URL}/checksums.txt.asc
gpg -u "Operator SDK (release) <cncf-operator-sdk@cncf.io>" --verify checksums.txt.asc
```
```bash
grep operator-sdk_${OS}_${ARCH} checksums.txt | sha256sum -c -
```
```bash
$ which operator-sdk
/usr/local/bin/operator-sdk
```

### 2. Start Minikube

I just did:

```bash
$ minikube start
```
Although in the instructions I've seen:

```bash
$ minikube start init
```

### 3. Local Workspace

At this point, I made sure I was in this present working directory, and I created
a new (v2) module and then "init" the operator:

```bash
$ go mod init vsoch/lolcow-operator
$ operator-sdk init
```

### 4. Create Controller

Now let's create a controller, and call it Lolcow

```bash
$ operator-sdk create api --version=v1alpha1 --kind=Lolcow
```

Make sure to install all dependencies (I think this might not be necessary - I saw it happen when I ran the previos command).

```bash
$ go mod tidy
$ go mod vendor
```

### 5. Make Manifests

At this point, we want to edit [controllers/lolcow_controller.go](controllers/lolcow_controller.go).
There is a good example to get started [here](https://github.com/deepak1725/hello-operator2/blob/main/controllers/traveller_controller.go).

Some design decisions I started to make:

1. If we expect to have more than one named API, it makes sense to have another directory under [api](api) (e.g., for lolcow).
2. If we want special struct/functions for a particular named api, this should be a custom package under [pkg](pkg) (so I created this directory, e.g., `pkg/lolcow`). The reason (I think) is because different versions of an API might want to use shared code.
3. kueue puts a lot of the controller logic under [pkg](https://github.com/kubernetes-sigs/kueue/tree/e571d42e390f96a95efa799d720777e92e4f69a4/pkg) but I'm not convinced I want that yet.
4. The examples use `mydomainv1alpha1` to reference the API package. This probably makes sense if you are importing different versions (why?) but my preference (only importing one) is to name it something simple like `api`.
5. I realized that if we want more than one controller, we should have subdirectories in controllers too. I mirrored the kueue design and made one called "core."
6. Since I don't know the ultimate design wanted (e.g., queue doesn't directly make a deployment or service but does via a queue manager) I mimicked the hello world example and made a deployment / service. I'd like to try making my own web UI to deploy for lolcow.

For all points, given that you are changing a path, make sure to grep for the old one so you don't miss updating one ;)

```bash
$ grep -R "vsoch/lolcow-operator/api"
PROJECT:  path: vsoch/lolcow-operator/api/v1alpha1
controllers/suite_test.go:	mydomainv1alpha1 "vsoch/lolcow-operator/api/v1alpha1"
controllers/lolcow_controller.go:	lolcow "vsoch/lolcow-operator/api/lolcow/v1alpha1"
main.go:	mydomainv1alpha1 "vsoch/lolcow-operator/api/v1alpha1"
```

When you finish developing (or as you develop!) you can do:

```bash
# quicker way to get errors to debug
$ go build main.go
```

And then officially build.
```bash
$ make
```

To make your manifests:

```bash
$ make manifests
```

And install?

```
$ make install
```

### 6. Deploy

At this point, edit the config yamls [here](config/samples/_v1alpha1_lolcow.yaml). We need to add a greeting, e.g,


```yaml
apiVersion: my.domain/v1alpha1
kind: Lolcow
metadata:
  name: lolcow-sample
spec:
  greeting: HELLO
```

And then apply (kustomize is in the bin).

```bash
$ bin/kustomize build config/samples | kubectl apply -f -
lolcow.my.domain/lolcow-sample created
```

And finally, run it.

```bash
$ make run
```

And you should be able to open the web-ui:

```bash
$ minikube service backend-service
```

This isn't perfect yet, but it's a start!

![img/hello-kubernetes.png](img/hello-kubernetes.png)

And then when it's running (in a separate terminal) change the greeting and do:

```bash
$ bin/kustomize build config/samples | kubectl apply -f -
lolcow.my.domain/lolcow-sample configured
```

I'd like to better understand what's going on under the hood here, but for now I'm happy to have something that sort of works!
