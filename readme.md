# Seldon Core Assessment

The assignment is the following:

 * Create a standalone program in Go which takes in a Seldon Core Custom Resource and creates it over the Kubernetes API 
 * Watch the created resource to wait for it to become available. 
 * When it is available delete the Custom Resource. 

## Overview

The overall program logic is the following:

1) load Seldon custom resource (CR) file e.g deploy.json and decode it into a SeldonDeployment object.
2) create Seldon CR using Kubernete client from controller runtime pkg.
3) periodically(1s) checking whether CR state == "Available"
4) Once CR state == "Available", delete the CR using the client.

## Setup

Make sure you have following tools on you machine:

* minikube = v1.5.2
* go = v1.13.0
* helm = v3.0.0

## Setup Seldon-operator

Install Seldon-operator into the default namespace:

```sh
$ helm install seldon-core seldon-core-operator --repo https://storage.googleapis.com/seldon-charts --set usageMetrics.enabled=true
NAME: seldon-core
LAST DEPLOYED: Thu Nov 21 00:08:18 2019
NAMESPACE: default
STATUS: deployed
REVISION: 1
TEST SUITE: None
```

Verify Seldon-operator is installed and Seldon-core CRD is configured:

```sh
$ kubectl get pod
NAME                                         READY   STATUS    RESTARTS   AGE
seldon-controller-manager-8655fb75f9-pqlzw   1/1     Running   0          15s

$ kubectl get crd
NAME                                          CREATED AT
seldondeployments.machinelearning.seldon.io   2019-11-20T23:08:20Z
```

## Build main.go

```sh
go build -o seldon-assessment
```

## Run seldon assessment

```sh
./seldon-assessment -cr_file=deploy.json
2019/11/21 00:12:38 Creating custom resource (seldon-model) ...
2019/11/21 00:12:38 Custom resource (seldon-model) created
2019/11/21 00:12:38 Checking if Seldon deployment is already ...
2019/11/21 00:12:38 Deployment state ()
2019/11/21 00:12:39 Deployment state (Creating)
...
2019/11/21 00:13:07 Deployment state (Available)
2019/11/21 00:13:07 Seldon deployment is already
2019/11/21 00:13:07 Deleting Seldon custom resource (seldon-model)
2019/11/21 00:13:07 Seldon custom resource (seldon-model) deleted

```

on a second terminal run following to watch seldon-core creation/deletion
```sh
$ kubectl get pod -w
NAME                                               READY   STATUS    RESTARTS   AGE
seldon-controller-manager-8655fb75f9-pqlzw         1/1     Running   0          4m30s
test-deployment-example-7cd068f-694cf755cc-qkx25   0/2     Running   0          12s
test-deployment-example-7cd068f-694cf755cc-qkx25   1/2     Running   0          27s
test-deployment-example-7cd068f-694cf755cc-qkx25   2/2     Running   0          28s
test-deployment-example-7cd068f-694cf755cc-qkx25   2/2     Terminating   0          30s
test-deployment-example-7cd068f-694cf755cc-qkx25   1/2     Terminating   0          43s
test-deployment-example-7cd068f-694cf755cc-qkx25   0/2     Terminating   0          52s
test-deployment-example-7cd068f-694cf755cc-qkx25   0/2     Terminating   0          53s
test-deployment-example-7cd068f-694cf755cc-qkx25   0/2     Terminating   0          53s
```