# Installing the EJBCA UpstreamAuthority plugin for SPIRE Server

The EJBCA UpstreamAuthority plugin for SPIRE Server contained within this repository is a third party plugin built using the `[spire-plugin-sdk](https://github.com/spiffe/spire-plugin-sdk)`. Third party plugins for SPIRE must be mounted to a path on the filesystem accessible by the SPIRE server binary. This guide details the general steps required to install the EJBCA UpstreamAuthority. The specific steps will vary widely depending on how/where SPIRE server is running.

## Requirements

### To build

* [Git](https://git-scm.com/)
* [Make](https://www.gnu.org/software/make/)
* [Go](https://golang.org/) >= v1.23.3

> The Makefile works best on Mac/Linux systems - Windows users may need to run the commands manually.

### To use

* EJBCA [Community](https://www.ejbca.org/) or EJBCA [Enterprise](https://www.keyfactor.com/products/ejbca-enterprise/)
  * The "REST Certificate Management" protocol must be enabled under System Configuration > Protocol Configuration.

## Local Installation

If the SPIRE server will not be running in a container, the plugin binary can be built and installed on the server's filesystem.

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

    3. Calculate the SHA256 checksum of the compiled binary

    ```shell
    export EJBCA_CHECKSUM=$( sha256sum bin/ejbca-spire-upstreamauthority-plugin | cut -d ' ' -f1 )
    ```

    </details>

    <details><summary>Download from GitHub</summary>

    1. Download and extract the plugin binary for your platform

    ```shell
    OS=$(go env GOOS)
    ARCH=$(go env GOARCH)
    curl -L https://github.com/Keyfactor/ejbca-spire-upstreamauthority-plugin/releases/latest/download/ejbca-spire-upstreamauthority-plugin-$OS-$ARCH.tar.gz
    mkdir -p bin
    tar xzf ejbca-vault-pki-engine-$OS-$ARCH.tar.gz -C bin/
    ```

    > If `go` isn't installed on your system, you can access the [releases](https://github.com/Keyfactor/ejbca-spire-upstreamauthority-plugin/releases) to download the plugin binary for your platform.

    2. Calculate the SHA256 checksum of the compiled binary - keep this value for later

    ```shell
    export EJBCA_CHECKSUM=$( sha256sum bin/ejbca-spire-upstreamauthority-plugin | cut -d ' ' -f1 )
    ```

    </details>

2. Copy the plugin binary to a known path on the SPIRE server filesystem

    ```shell
    sudo cp bin/ejbca-spire-upstreamauthority-plugin /opt/spire/plugins/bin/
    ```

3. Update the SPIRE server configuration to include the plugin binary

    ```shell
    UpstreamAuthority "ejbca" {
        plugin_cmd = "/opt/spire/plugins/bin/ejbca-spire-upstreamauthority-plugin"
        plugin_checksum = "$EJBCA_CHECKSUM"
            plugin_data {
            hostname = "ejbca.example.com"
            ca_cert = <<EOF
    -----BEGIN CERTIFICATE-----
    MIIDE ... mn+GJf
    -----END CERTIFICATE-----
    EOF
            oauth {
                token_url = "https://dev.idp.com/oauth/token"
                client_id = "<client_id>"
                client_secret = "<client_secret>"
            }
            ca_name = "Fake-Sub-CA"
            end_entity_profile_name = "fakeSpireIntermediateCAEEP"
            certificate_profile_name = "fakeSubCACP"
            end_entity_name = ""
            account_binding_id = "abc123"
        }
    }
    ```

    > For a complete list of configuration parameters and their descriptions, please refer to the [usage](usage.md) documentation.

## Using the EJBCA UpstreamAuthority plugin for SPIRE Server

Refer to the [usage](usage.md) documentation for information on how to configure the EJBCA UpstreamAuthority plugin for SPIRE Server.
