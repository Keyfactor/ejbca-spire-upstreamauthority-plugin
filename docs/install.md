# Installing the EJBCA UpstreamAuthority plugin for SPIRE Server

The EJBCA UpstreamAuthority plugin for SPIRE Server contained within this repository is a third party plugin built using the `[spire-plugin-sdk](https://github.com/spiffe/spire-plugin-sdk)`. Third party plugins for SPIRE must be mounted to a path on the filesystem accessible by the SPIRE server binary. This guide proposes two installation proposes two installation patterns.

## Requirements


## Dockerfile Installation

The EJBCA UpstreamAuthority plugin for SPIRE Server can be installed by creating a custom Dockerfile that uses the SPIRE Server container image as a base and copies the plugin binary to a known path. 

1. Clone the EJBCA UpsreamAuthority plugin repository
    ```shell
    git clone https://github.com/Keyfactor/ejbca-spire-upstreamauthority-plugin.git
    ```

2. Build the plugin binary
    ```shell
    make build
    ```

3. Create a Dockerfile that copies the plugin binary to the SPIRE Server container image
    ```Dockerfile
    FROM ghcr.io/spiffe/spire-server:<tag>

    COPY ejbca-spire-upstreamauthority-plugin /opt/spire/plugins/bin/
    ```

    > You can find a list of available image tags from the [SPIRE Server container registry](https://github.com/spiffe/spire/pkgs/container/spire-server).

## Kubernetes Volume Mount via the SPIRE Helm Chart


