<!-- Image at Center -->
<p align="center">
  <img src="./assets/usage.png">
</p>

<!-- Description & Menu at Center -->
<div align="center">
  <h1 align="center">Silver Surfer - Kubedd</h1>
  <p align="center">
    ApiVersion Compatibility Checker & Provides Migration Path for K8s Objects
    <br />
    <a href="#bulb-motivation"><strong>Motivation</strong></a>
    |
    <a href="#rocket-getting-started"><strong>Getting Started</strong></a>
    |
    <a href="#gear-usage"><strong>Usage</strong></a>
    |
    <a href="#file_folder-output"><strong>Output</strong></a>
    <br />
    <a href="https://github.com/devtron-labs/silver-surfer/issues/new">Report Bug</a>
    |
    <a href="https://github.com/devtron-labs/silver-surfer/issues/new">Request Feature</a>
    |
    <a href="#handshake-contribute">Support</a>

  <a href="https://discord.gg/jsRG5qx2gp"><img src="https://img.shields.io/badge/Join%20us%20on-Discord-e01563.svg" alt="Join Discord"></a>
  <a href="https://goreportcard.com/badge/github.com/devtron-labs/devtron"><img src="https://goreportcard.com/badge/github.com/devtron-labs/devtron" alt="Go Report Card"></a>
  <a href="./LICENSE"><img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="License"></a>
  <a href="https://bestpractices.coreinfrastructure.org/projects/4411"><img src="https://bestpractices.coreinfrastructure.org/projects/4411/badge" alt="CII Best Practices"></a>
  <a href="http://golang.org"><img src="https://img.shields.io/badge/Made%20with-Go-1f425f.svg" alt="made-with-Go"></a>
  <a href="http://devtron.ai/"><img src="https://img.shields.io/website-up-down-green-red/http/shields.io.svg" alt="Website devtron.ai"></a>
  <a href="https://twitter.com/intent/tweet?text=Silver Surfer%20helps%20in%20checking%20ApiVersion compatibility%20and%20gives%20migration-path%20for%20Kubernetes%20objects%20check%20it%20out!!%20&hashtags=OpenSource,Kubernetes,DevOps,golang&url=https://github.com/devtron-labs/silver-surfer%0a"><img src="https://img.shields.io/twitter/url/http/shields.io.svg?style=social" alt="Tweet"></a>
  </p>
</div>

## :bulb: Motivation

Currently there is no easy way to upgrade Kubernetes objects in case of Kubernetes newer versions. There are some tools
which are available for this purpose, but we found them inadequate for migration requirements.

`kubedd` is a tool to check issues in migration of Kubernetes yaml objects from one Kubernetes version to another.

It uses openapi spec provided by the Kubernetes with releases, for eg. in case of target kubernetes version 1.27 openapi spec for [1.27](https://raw.githubusercontent.com/kubernetes/kubernetes/release-1.27/api/openapi-spec/swagger.json),
 to validate the kubernetes objects for depreciation or non-conformity with openapi spec.

Supported input formats

1. Directory containing files to be validated
2. Read kubernetes objects directly from cluster. Uses `kubectl.kubernetes.io/last-applied-configuration` to get
   last applied configuration and in its absence uses the manifest itself.

It provides details of issues with the Kubernetes object in case they are migrated to cluster with newer Kubernetes
version.

## :rocket: Getting Started

### Quick Installation

Just with few commands, it's ready to serve your cluster.

```bash
git clone https://github.com/devtron-labs/silver-surfer.git
cd silver-surfer
go mod vendor
go mod download
make 
```

It's done. A `bin` directory must have created with the binary ready to use `./kubedd` command.

### Running Within Container

You can also use the Dockerfile present to run command within a container and analyze the cluster running in your host machine. Switch to the project directory containing the dockerfile and execute the following commands.

1. Build the container image with name `silver-surfer`

```bash
docker build . -t silver-surfer --build-arg RELEASE=goreleaser --build-arg AUTH_TOKEN=YOUR_GITHUB_TOKEN
```

2. Mount the host directory containing kubeconfig with container and run the container with name `kubedd`

```bash
docker run -v /host/path-to/.kube-dir/:/opt/.kube --privileged --net=host --name kubedd silver-surfer --kubeconfig /opt/.kube/config
```

### Using Binaries

You can download the binaries for Windows, Linux and MacOS from the [release page](https://github.com/devtron-labs/silver-surfer/releases) on this repository.

## :gear: Usage

Use the binary `./kubedd` to execute the commands and get insights of your current Kubernetes objects.

```yaml
./kubedd --help
Validates migration of Kubernetes YAML file against specific Kubernetes versions. It provides details of issues with the Kubernetes objects in case they are migrated to cluster with newer Kubernetes version

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
      --no-color                              Display results without color
      --select-kinds strings                  A comma-separated list of kinds to be selected, if left empty all kinds are selected
      --select-namespaces strings             A comma-separated list of namespaces to be selected, if left empty all namespaces are selected
      --source-kubernetes-version string      Version of Kubernetes of the cluster on which kubernetes objects are deployed currently, ignored in case cluster is provided. In case of directory defaults to same as target-kubernetes-version.
      --source-schema-location string         SourceSchemaLocation is the file path of kubernetes versions of the cluster on which manifests are deployed. Use this in air-gapped environment where internet access is unavailable.
      --target-kubernetes-version string      Version of Kubernetes to migrate to eg 1.22, 1.21, 1.12 (default "1.22")
      --target-schema-location string         TargetSchemaLocation is the file path of kubernetes version of the target cluster for these manifests. Use this in air-gapped environment where internet access is unavailable.
      --version                               version for kubedd
```

## :file_folder: Output

It categorises Kubernetes objects based on change in ApiVersion. Categories are -

1. Removed ApiVersion
2. Deprecated ApiVersion
3. Newer ApiVersion
4. Unchanged ApiVersion

Within each category it identifies migration path to newer ApiVersion, possible paths are -

1. It cannot be migrated as there are no common ApiVersions between source and target kubernetes version
2. It can be migrated but has some issues which need to be resolved
3. It can be migrated with just ApiVersion change

This activity is performed for both current and new ApiVersion.

## :handshake: Contribute

Collaborations and contributions are the beauty of open source communities. It creates an environment where we learn, inspire and create amazing tools with the help of community to solve the real-life use cases. Here are couple of ways you can contribute to silver-surfer -

1. Create content around silver-surfer (blogs, videos, podcast, etc)
2. Pick any of the [open issues](https://github.com/devtron-labs/silver-surfer/issues) and [raise a PR](https://dev.to/abhinavd26/start-your-open-source-journey-with-git-20o3) for it.
3. Give it a [star ⭐️](https://github.com/devtron-labs/silver-surfer) if you like the project.

Check out our [contributing guidelines](CONTRIBUTING.md) for more details. We deeply appreciate your contributions.

## :link: Other Similar Tools

1. [kubeval](https://github.com/instrumenta/kubeval) - The most popular, only validates against the given Kubernetes version, doesn't provide migration path and is no longer maintained.
2. [kube-no-trouble](https://github.com/doitintl/kube-no-trouble) - Provides information about removed and deprecated APIs but doesn't validate schema.
3. [kubepug](https://github.com/rikatz/kubepug) - Provides information based on deprecation comments in the schema, doesn't provide information
