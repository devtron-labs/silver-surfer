# Kubedd

## Motivation

Currently there is no easy way to upgrade kubernetes objects in case of kubernetes upgrade. There are some tools
which are available for this purpose, but we found them inadequate for migration requirements.

`kubedd` is a tool to check issues in migration of kubernetes yaml objects from one kubernetes version to another. 

It uses openapi spec provided by the kubernetes with releases, for eg. in case of target kubernetes version 1.22 openapi spec for [1.22](https://raw.githubusercontent.com/kubernetes/kubernetes/release-1.22/api/openapi-spec/swagger.json),
 to validate the kubernetes objects for depreciation or non-conformity with openapi spec.

Supported input formats
1. Directory containing files to be validated
2. Read kubernetes objects directly from cluster.Uses `kubectl.kubernetes.io/last-applied-configuration` to get
   last applied configuration and in its absence uses the manifest itself. 
 
It provides details of issues with the kubernetes object in case they are migrated to cluster with newer kubernetes
version.

## Getting Started

#### Quick Installation
Just with few commands, it's ready to serve your cluster.

```bash
git clone https://github.com/devtron-labs/silver-surfer.git
cd silver-surfer
go mod vendor
go mod download
make 
```

It's done. A `bin` directory might have created with the binary ready to use `./kubedd` command.

#### Running Within Container
You can also use the Dockerfile present to run command within a container and analyze the cluster running in your host machine.

```bash
docker build -t silver-surfer:v1.0 --build-arg RELEASE=goreleaser --build-arg auth=YOUR_GITHUB_TOKEN
docker run -v /host/path-to/.kube-dir/:/opt/.kube --privileged --net=host --name kubedd silver-surfer:v1.0 --kubeconfig /opt/.kube/config
```
#### Using Binaries
You can download the binaries for Windows, Linux and MacOS from the [release page](https://github.com/devtron-labs/silver-surfer/releases) on this repository.

## Usage

<p align="center"><img src="./assets/usage.png"></p>

## Arguments

```
./kubedd --help
Validates migration of Kubernestes YAML file against specific kubernetes version, it provides details of issues with the kubernetes object in case they are migrated to cluster with newer kubernetes version

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
      --source-schema-location string         SourceSchemaLocation is the file path of kubernetes versions of the cluster on which manifests are deployed. Use this in air-gapped environment where internet access is unavailable.
      --target-kubernetes-version string      Version of Kubernetes to migrate to eg 1.22, 1.21, 1.12 (default "1.22")
      --target-schema-location string         TargetSchemaLocation is the file path of kubernetes version of the target cluster for these manifests. Use this in air-gapped environment where internet access is unavailable.
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
2. It can be migrated but has some issues which need to be resolved
3. It can be migrated with just ApiVersion change

This activity is performed for both current and new ApiVersion.

## Contribute

Check out our [contributing guidelines](CONTRIBUTING.md). We deeply appreciate your contributions.

## Other Similar Tools

1. [kubeval](https://github.com/instrumenta/kubeval) - most popular, only validates against the given kubernetes version, doesn't provide migration path
2. [kube-no-trouble](https://github.com/doitintl/kube-no-trouble) - provides information about removed and deprecated api but doesn't validate schema
3. [kubepug](https://github.com/rikatz/kubepug) - provides information based on deprecation comments in the schema, doesn't provide information


