# PodHealth Controller using Operator SDK

## Created with

```sh
operator-sdk new operatorsdk --repo github.com/loodse/operator-workshop/podhealth/operatorsdk
operator-sdk add api --api-version='training.loodse.io/v1alpha1' --kind=PodHealth
operator-sdk add controller --api-version=training.loodse.io/v1alpha1 --kind=PodHealth

# after implementing types
operator-sdk generate k8s
operator-sdk generate openapi
```

And now implementing the type and controller.

## To Try out

```sh
# connect to a Kubernetes cluster

# install CRDs and create sample
kubectl apply -f deploy/crds

# run Operator
operator-sdk up local
```
