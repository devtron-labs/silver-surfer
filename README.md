# Kubedd

As kubernetes has come of its age, it has started deprecating and removing older kubernetes objects apiVersions.

`kubedd` is a tool to help specifically with the migration of kubernetes version, it detects
1. Identify and provide recommendations for apiVersions which have been removed or deprecated
2. Validates schema against current kubernetes version as well as target kubernetes version
3. Warns in case it is not possible to migrate from current version to target kubernetes version

It does so using Kubernetes OpenAPI specification, and can validate schemas for multiple versions of Kubernetes.



```
$ ./bin/kubedd --target-kubernetes-version 1.22  


```


For full usage and installation instructions see [devtron.ai](https://docs.devtron.ai/).
