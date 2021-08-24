# Kubedd

## Motivation

https://github.com/kubernetes/kubernetes/issues/58131#issuecomment-403829566

Currently there is no easy way to upgrade kubernetes objects in case of kubernetes upgrade. There are some tools
which are available for this purpose, but we found then inadequate for migration requirements.

`kubedd` is a tool to check issues in migration of kubernetes yaml objects from one kubernetes version to another. 

It uses openapi spec provided by the kubernetes with releases, for eg. in case of target kubernetes version 1.22 openapi spec for [1.22](https://raw.githubusercontent.com/kubernetes/kubernetes/release-1.22/api/openapi-spec/swagger.json),
 to validate the kubernetes objects for depreciation or non-conformity with openapi spec.

Supported input formats
1. Directory containing files to be validated
2. Read kubernetes objects directly from cluster.Uses `kubectl.kubernetes.io/last-applied-configuration` to get
   last applied configuration and in its absence uses the manifest itself. 
 
It provides details of issues with the kubernetes object in case they are migrated to cluster with newer kubernetes
version.

## Install

Download kubedd, and it is ready for use.

## Usage

```
./kubedd 

Results for cluster at version 1.12 to 1.22
-------------------------------------------
>>>> Removed API Version's <<<<
 Namespace   Name                          Kind         API Version (Current Available)   Replace With API Version (Latest Available)   Migration Status                               
 prod        demmoo-prod-ingress           Ingress      extensions/v1beta1                                                              Alert! cannot migrate kubernetes version  
 prod        devtron-static-prod-ingress   Ingress      extensions/v1beta1                                                              Alert! cannot migrate kubernetes version  
 prod        ghost-blog-dt-prod            Ingress      extensions/v1beta1                                                              Alert! cannot migrate kubernetes version  
 prod        ghost-blog-dt-prod-auth       Ingress      extensions/v1beta1                                                              Alert! cannot migrate kubernetes version  
 prod        ghost-devtron-blog-prod       Ingress      extensions/v1beta1                                                              Alert! cannot migrate kubernetes version  
 prod        oauth2-proxy                  Ingress      extensions/v1beta1                                                              Alert! cannot migrate kubernetes version  
 prod        telemetry-prod-ingress        Ingress      extensions/v1beta1                                                              Alert! cannot migrate kubernetes version  
 prod        ghost-blog-dt-prod            Deployment   extensions/v1beta1                apps/v1                                       can be migrated with just apiVersion change    
 prod        ghost-devtron-blog-prod       Deployment   extensions/v1beta1                apps/v1                                       can be migrated with just apiVersion change  
```


## Arguments

```
./kubedd --help
Validates migration of Kubernestes YAML file against specific kubernetes version, It provides details of issues with the kubernetes object in case they are migrated to cluster with newer kubernetes version

Usage:
  kubedd <file> [file...] [flags]

Flags:
  -d, --directories strings                   A comma-separated list of directories to recursively search for YAML documents
      --force-color                           Force colored output even if stdout is not a TTY
  -h, --help                                  help for kubedd
      --ignore-keys-for-deprecation strings   A comma-separated list of keys to be ignored for depreciation check (default [metadata*,status*])
      --ignore-keys-for-validation strings    A comma-separated list of keys to be ignored for validation check (default [status*,metadata*])
      --ignore-kinds strings                  A comma-separated list of kinds to be skipped (default [event,CustomResourceDefinition])
      --ignore-namespaces strings             A comma-separated list of namespaces to be skipped (default [kube-system])
      --ignore-null-errors                    Ignore null value errors (default true)
      --ignored-filename-patterns strings     An alias for ignored-path-patterns
  -i, --ignored-path-patterns strings         A comma-separated list of regular expressions specifying paths to ignore
      --insecure-skip-tls-verify              If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string                     Path of kubeconfig file of cluster to be scanned
      --kubecontext string                    Kubecontext to be selected
      --select-kinds strings                  A comma-separated list of kinds to be selected, if left empty all kinds are selected
      --select-namespaces strings             A comma-separated list of namespaces to be selected, if left empty all namespaces are selected
      --source-kubernetes-version string      Version of Kubernetes of the cluster on which kubernetes objects are deployed currently, ignored in case cluster is provided. In case of directory defaults to same as target-kubernetes-version.
      --source-schema-location string         SourceSchemaLocation is the file path of kubernetes versions of the cluster on which manifests are deployed. Use this in air-gapped environment where it internet access is unavailable.
      --target-kubernetes-version string      Version of Kubernetes to migrate to eg 1.22, 1.21, 1.12 (default "1.22")
      --target-schema-location string         TargetSchemaLocation is the file path of kubernetes version of the target cluster for these manifests. Use this in air-gapped environment where it internet access is unavailable.
      --version                               version for kubedd


```

## Output

It categorises kubernetes objects based on change in ApiVersion. Categories are
1. Removed ApiVersion
2. Deprecated ApiVersion
3. Newer ApiVersion
4. Unchanged ApiVersion

Within each category it identifies migration path to newer ApiVersion, possible paths are
1. It cannot be migrated as there are no common ApiVersions between source and target kubernetes version
2. It can be migrated but has some issues which needs to be resolved
3. It can be migrated with just ApiVersion change

This activity is performed for both current and new ApiVersion.

## Other Similar Tools

1. [kubeval](https://github.com/instrumenta/kubeval) - most popular, only validates against the given kubernetes version, doesn't provide migration path
2. [kube-no-trouble](https://github.com/doitintl/kube-no-trouble) - provides information about removed and deprecated api but doesnt validate schema
3. [kubepug](https://github.com/rikatz/kubepug) - provides information based on deprecation comments in the schema, doesn't provide information


