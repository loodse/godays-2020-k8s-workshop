# PodHealth Controller using kubebuilder

## Created with

```sh
kubebuilder init --domain 'loodse.io'
kubebuilder create api --group training --version v1alpha1 --kind PodHealth

# after implementing types
make manifests
make generate
```

And now implementing the type and controller.

## To Try out

```sh
# connect to a Kubernetes cluster

# install CRDs
make install

# run Operator
make run

# create sample
kubectl apply -f config/samples/training_v1alpha1_podhealth.yaml -n kube-system
```
