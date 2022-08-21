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
 - my lolcow operator container is prebuilt at [ghcr.io/vsoch/lolcow-operator](https://github.com/vsoch/lolcow-operator/pkgs/container/lolcow-operator)
  
The sections below will describe:

 - [Using the Operator](#using-the-operator): as it is provided here
 - [Making the Operator](#making-the-operator): steps that I went through (and what I learned)
 - [Building the Lolcat Container](#building-operator-container): yes, I made this little custom UI for the example, if you want to play with it separately!
 - [Wisdom](#wisdom): I picked up from the kubebuidler slack - shout out to mogsie for being so helpful!


## Using the Operator

If you aren't starting from scratch, then you can use the code here to see how things work!

### 1. Start Minikube

First, start minikube.

```bash
$ minikube start
```

If you haven't ever installed it, you can see [install instructions here](https://minikube.sigs.k8s.io/docs/start/).

### 2. Build

And then officially build.

```bash
$ make
```

To make your manifests:

```bash
$ make manifests
```

And install. Note that this places an executable [bin/kustomize](bin/kustomize) that you'll need to delete first if you make install again.

```bash
$ make install
```

### 3. Deploy

Note that you will be using the config yamls [here](config/samples/_v1alpha1_lolcow.yaml) to start, which include a greeting and port.
We will look at these later for demonstrating how the operator watches for changes. Apply your configs (kustomize is in the bin).

```bash
$ bin/kustomize build config/samples | kubectl apply -f -
lolcow.my.domain/lolcow-pod created
```

And finally, run it.

```bash
$ make run
```

And you should be able to open the web-ui:

```bash
$ minikube service lolcow-pod
```

Note that if you get a 404 page, do `kubectl get svc` and wait until the service goes from "pending" to "ready." You should 
see the initial message from the lolcow:

![img/hello-lolcow.png](img/hello-lolcow.png)

If you were to Control+C and restart the controller, you'd see the greeting hasn't changed:

```bash
1.6611115156486864e+09	INFO	👋️ No Change to Greeting! 👋️: 	{"controller": "lolcow", "controllerGroup": "my.domain", "controllerKind": "Lolcow", "lolcow": {"name":"lolcow-pod","namespace":"default"}, "namespace": "default", "name": "lolcow-pod", "reconcileID": "73ace2ec-c882-45d2-bdd9-860dd5a65f22", "Lolcow": "default/lolcow-pod", "Hello, this is a message from the lolcow!": "Hello, this is a message from the lolcow!"}
```

### 4. Change the Greeting 

Now let's try changing the greeting. This will test our controllers ability to watch the config and update the deployment accordingly. At this point, edit the config yamls [here](config/samples/_v1alpha1_lolcow.yaml). Change just the greeting for now:

```yaml
apiVersion: my.domain/v1alpha1
kind: Lolcow
metadata:
  name: lolcow-pod
spec:
  port: 30685
```
```diff
-  greeting: Hello, this is a message from the lolcow!
+ greeting: What, you've never seen a poptart cat before?
```

You can make this change while it's running (in a separate terminal) and then change the greeting in the original config and do:

```bash
$ bin/kustomize build config/samples | kubectl apply -f -
lolcow.my.domain/lolcow-pod configured
```

The change might be quick, but if you scroll up you should see:

```
1.6611116913288918e+09	INFO	👋️ New Greeting! 👋️: 	{"controller": "lolcow", "controllerGroup": "my.domain", "controllerKind": "Lolcow", "lolcow": {"name":"lolcow-pod","namespace":"default"}, "namespace": "default", "name": "lolcow-pod", "reconcileID": "a3ee80ae-7cd3-4f90-8dac-81c2cfb1708c", "Lolcow": "default/lolcow-pod", "What, you've never seen a poptart cat before?": "Hello, this is a message from the lolcow!"}
```

and the interface should change too!

![img/poptart-cat.png](img/poptart-cat.png)
   
### 5. Change the Port

Since we haven't changed the port, in the logs you should see:

```bash
1.6611135948097324e+09	INFO	🔁 No Change to Port! 🔁:
```

So now let's try changing the port, maybe to one number higher:

```yaml
apiVersion: my.domain/v1alpha1
kind: Lolcow
metadata:
  name: lolcow-pod
spec:
```
```diff
-  port: 30685
+  port: 30686
```

And then apply the config. Refreshing the current browser should 404, and you should be able to tweak the port number in your browser and see the user interface again!
Yay, it works!

### 7. Cleanup

When cleaning up, you can control+c to kill the operator from running, and then:

```bash
$ kubectl delete pod --all
$ kubectl delete svc --all
$ minikube stop
```

And that's it! You can also delete your minikube cluster if you like.

### 7. Caveats

This is my first time doing any kind of development for Kubernetes, and this is a very basic intro
that doesn't necessarily reflect best practices. When I'm newly learning something, my main goal
is to get it to work (period!) and then to slowly learn better practices over time (and use them
as a standard). I hope this has been useful to you!

## Making the operator

This section will walk through some of the steps that @vsoch took to create the controller using the operator-sdk, and challenges she faced.

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

Note that you don't need to do this, obviously, if you are using the existing operator here!

### 4. Create Controller

Now let's create a controller, and call it Lolcow (again, no need to do this if you are using the one here).

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

And then see the instructions above for [using the operator](#using-the-operator).


### 6. Bugs

#### Service / Deployment Detection

For the longest time, the original service and deployment would start (because they were not found) but they would *continue* to be not found
and sort of spiral into a chain of error messages. This took me many evenings to figure out, but it comes down to these (sort of hidden) lines
at the top of the controllers file:

```
//+kubebuilder:rbac:groups=my.domain,resources=lolcows,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=my.domain,resources=lolcows/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=my.domain,resources=lolcows/finalizers,verbs=update
//+kubebuilder:rbac:groups=my.domain,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=my.domain,resources=pods,verbs=get;list;watch;create;
//+kubebuilder:rbac:groups=my.domain,resources=services,verbs=get;list;watch;create;update;patch;delete
```

The template only had the first three (for lolcows) and I needed to add the last three, giving my contoller permission (RBAC refers
to a set of rules that represent a set of permissions) to interact with services and deployments. I think what was happening
before is that my controller couldn't see them, period, so of course the Get always failed. I found [this page](https://cluster-api.sigs.k8s.io/developer/providers/implementers-guide/controllers_and_reconciliation.html) and [this page](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#role-and-clusterrole) useful for learning about this.
Another important note (that I didn't do here) is that you can namespace these, which I suspect is best practice but I didn't do for this little demo. The other bit that
seemed important was to say that my controller owned services and deployments:

```go
func (r *LolcowReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.Lolcow{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		// Defaults to 1, putting here so we know it exists!
		WithOptions(controller.Options{MaxConcurrentReconciles: 1}).
		Complete(r)
}
```
But I'm not entirely sure if that was necessary given the RBAC - something to test for sure.

#### Pod Names

For some reason, at one point I switched my name from 'lolcow-sample' to 'lolcow-pod', and although I thought I cleaned everything up,
whenever I'd create the cluster again it would show me *two* pods made, one lolcow-pod and one lolcow-sample. I had to try resetting and
"starting fresh" multiple times (e.g., deleting stuff in bin and reinstalling everything) until `kubectl get pod` didn't show the older
and new name. If you run into errors about not finding a service, it could be that somewhere the older name is still being created or referenced,
so it's a good sanity check to do.

## Building Operator Container

**note** you shouldn't need to do this as it will pull from GitHub packages, but if you want to locally test, this
is how you do it!

### 1. Build the Container

```bash
$ docker build -f docker/Dockerfile -t ghcr.io/vsoch/lolcow-operator .
```

And then you can run it without a statement (and we will use the fortune command to get one) or with a custom statement.

```bash
$ docker run -p 8080:8080 -it ghcr.io/vsoch/lolcow-operator  "Oh my gosh, I am a cow in a container!"
$ docker run -p 8080:8080 -it ghcr.io/vsoch/lolcow-operator  
```

This will be the container we deploy to our operator, with entrypoint modified with our greeting.
 
After the greeting you'll see that a web server is started, and you can open up to [http://localhost:8080](http://localhost:8080) to see it.

```bash
$ docker run -it -p 8080:8080 ghcr.io/vsoch/lolcow-operator
 ____________________________________
< Be cautious in your daily affairs. >
 ------------------------------------
        \   ^__^
         \  (oo)\_______
            (__)\       )\/\
                ||----w |
                ||     ||
 * Serving Flask app 'app'
 * Debug mode: off
WARNING: This is a development server. Do not use it in a production deployment. Use a production WSGI server instead.
 * Running on all addresses (0.0.0.0)
 * Running on http://127.0.0.1:8080
 * Running on http://172.17.0.2:8080
Press CTRL+C to quit
```

And then when you open to [http://localhost:8080](http://localhost:8080) you will see a *much improved* lol cat... has turned
into Nyan Cat!

![img/nyan-cat.png](img/nyan-cat.png)


## Troubleshooting

If you need to clean things up (ensuring you only have this one pod and service running first) I've found it easier to do:

```bash
$ kubectl delete pod --all
$ kubectl delete svc --all
```

If you see:

```bash
1.6605195805812113e+09	ERROR	controller-runtime.source	if kind is a CRD, it should be installed before calling Start	{"kind": "Lolcow.my.domain", "error": "no matches for kind \"Lolcow\" in version \"my.domain/v1alpha1\""}
```

You need to remove the previous kustomize and install the CRD again:

```bash
$ rm bin/kustomize
$ make install
```

## Wisdom

**from the kubebuilder slack**

### Learned Knowledge

- Reconciling should only take into account the spec of your object, and the real world.  Don't use status to hold knowledge for future reconcile loops.  Use a workspace object instead.
- Status should only hold observations of the reconcile loop.  Conditions, perhaps a "Phase", IDs of stuff you've found, etc.
- Use k8s ownership model to help with cleaning up things that should automatically be reclaimed when your object is deleted.
- Use finalizers to do manual clean-up-tasks
- Send events, but be very limited in how often you send events.  We've opted now to send events, essentially only when a Condition is modified (e.g. a Condition changes state or reason).
- Try not to do too many things in a single reconcile.  One thing is fine.  e.g. see one thing out of order?  Fix that and ask to be reconciled.  The next time you'll see that it's in order and you can check the next thing.  The resulting code is very robust and can handle almost any failure you throw at it.
- Add "kubebuilder:printcolums" markers to help kubectl-users get a nice summary when they do "kubectl get yourthing".
- Accept and embrace that you will be reconciling an out-of-date object from time to time.  It shouldn't really matter.  If it does, you might want to change things around so that it doesn't matter.  Inconsistency is a fact of k8s life.
- Place extra care in taking errors and elevating them to useful conditions, and/or events.  These are the most visible part of an operator, and the go-to-place for humans when trying to figure out why your code doesn't work.  If you've taken the time to extract the error text from the underlying system into an Event, your users will be able to fix the problem much quicker.

### What is a workspace?

A workspace object is when you need to record some piece of knowledge about a thing you're doing, so that later you can use that when reconciling this object. MyObject "foo" is reconciled; so to record the thing you need to remember, create a MyObjectWorkspace — Owned by the MyObject, and with the same name + namespace.  MyObjectWorkspace doesn't need a reconciler; it's simply a tool for you to remember the thing. Next time you reconcile a MyObject, also read your MyObjectWorkspace so you can remember "what happened last time". E.g. I've made a controller to create an EC2 instance, and we needed to be completely sure that we didn't make the "launch instance" API call twice.  EC2 has a "post once only" technique whereby you specify a nonce to avoid duplicate API calls.  You would write the nonce to the workspace use the nonce to call the EC2 API write any status info of what you observed to the status. Rremove the nonce when you know that you've stored the results (e.g. instance IDs or whatever) When you reconcile, if the nonce is set, you can re-use it because it means that the EC2 call failed somehow.  EC2 uses the nonce the second time to recognise that "heh, this is the same request as before ..." Stuff like this nonce shouldn't go in your status. Put simply, the status should really never be used as input for your reconcile.

Know that the scaffolded k8sClient includes a cache that automatically updates based on watches, and may give you out-of-date data (but this is fine because if it is out-of-date, there should be a reconcile in the queue already). Also know that there is a way to request objets bypassing a cache (look for APIReader).  This gives a read-only, but direct access to the API.  Useful for e.g. those workspace objects.
