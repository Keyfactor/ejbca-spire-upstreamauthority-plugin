# Installing the EJBCA UpstreamAuthority plugin for SPIRE Server

The EJBCA UpstreamAuthority plugin for SPIRE Server contained within this repository is a third party plugin built using the `[spire-plugin-sdk](https://github.com/spiffe/spire-plugin-sdk)`. Third party plugins for SPIRE must be mounted to a path on the filesystem accessible by the SPIRE server binary. This guide proposes two installation proposes two installation patterns.

## Requirements
### To build

* [Git](https://git-scm.com/)
* [Make](https://www.gnu.org/software/make/)
* [Go](https://golang.org/) >= v1.23.3

> The Makefile works best on Mac/Linux systems - Windows users may need to run the commands manually.

### To use

* EJBCA [Community](https://www.ejbca.org/) or EJBCA [Enterprise](https://www.keyfactor.com/products/ejbca-enterprise/)
  * The "REST Certificate Management" protocol must be enabled under System Configuration > Protocol Configuration.

## Dockerfile Installation

The EJBCA UpstreamAuthority plugin for SPIRE Server can be installed by creating a custom Dockerfile that uses the SPIRE Server container image as a base and copies the plugin binary to a known path. 

1. Build or Download the plugin binary.
    <details><summary>Build from source</summary>

    1. Clone the EJBCA UpsreamAuthority plugin repository

    ```shell
    git clone https://github.com/Keyfactor/ejbca-spire-upstreamauthority-plugin.git
    cd ejbca-spire-upstreamauthority-plugin
    ```

    2. Build the plugin binary

    ```shell
    make build
    ```

    3. Calculate the SHA256 checksum of the compiled binary - keep this value for later

    ```shell
    sha256sum ejbca-spire-upstreamauthority-plugin | cut -d ' ' -f1
    ```
    </details>

    <details><summary>Download from GitHub</summary>

    1. Download and extract the plugin binary for your platform

    ```shell
    OS=$(go env GOOS)
    ARCH=$(go env GOARCH)
    curl -L https://github.com/Keyfactor/ejbca-vault-pki-engine/releases/latest/download/ejbca-vault-pki-engine-$OS-$ARCH.tar.gz
    tar xzf ejbca-vault-pki-engine-$OS-$ARCH.tar.gz
    ```

    2. Calculate the SHA256 checksum of the compiled binary - keep this value for later

    ```shell
    sha256sum ejbca-spire-upstreamauthority-plugin | cut -d ' ' -f1
    ```

    </details>

  1. 

3. Create a Dockerfile that copies the plugin binary to the SPIRE Server container image
    ```Dockerfile
    FROM ghcr.io/spiffe/spire-server:<tag>

    COPY ejbca-spire-upstreamauthority-plugin /opt/spire/plugins/bin/
    ```

    > You can find a list of available image tags from the [SPIRE Server container registry](https://github.com/spiffe/spire/pkgs/container/spire-server).

## Kubernetes Volume Mount via the SPIRE Helm Chart


