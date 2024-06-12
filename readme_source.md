<!--- Insert the Tool Name in the main heading! --->
# EJBCA UpstreamAuthority plugin for SPIRE Server

<!--EJBCA Community logo -->
<a href="https://ejbca.org">
    <img src=".github/images/community-ejbca.png?raw=true)" alt="EJBCA logo" title="EJBCA" height="70" />
</a>
<!--EJBCA Enterprise logo -->
<a href="https://www.keyfactor.com/products/ejbca-enterprise/">
    <img src=".github/images/keyfactor-ejbca-enterprise.png?raw=true)" alt="EJBCA logo" title="EJBCA" height="70" />
</a>

[![Go Report Card](https://goreportcard.com/badge/github.com/Keyfactor/ejbca-spire-upstreamauthority-plugin)](https://goreportcard.com/report/github.com/Keyfactor/ejbca-spire-upstreamauthority-plugin)

<!--- Short intro here! --->
<!--- Include a description of the project/repository, the purpose of it, what problems it solves, when to use it (and not use it), etc. --->

The `ejbca` UpstreamAuthority plugin uses a connected [EJBCA](https://www.ejbca.org/) to issue intermediate signing certificates for the SPIRE server. The plugin can authenticate to EJBCA using mTLS (client certificate) or using the OAuth 2.0 "client credentials" token flow (sometimes called two-legged OAuth 2.0).

## System Requirements

<!--- Insert any requirements in this section. --->
### To build

* [Git](https://git-scm.com/)
* [Make](https://www.gnu.org/software/make/)
* [Go](https://golang.org/) >= v1.23.3

### To use

* EJBCA [Community](https://www.ejbca.org/) or EJBCA [Enterprise](https://www.keyfactor.com/products/ejbca-enterprise/)
  * The "REST Certificate Management" protocol must be enabled under System Configuration > Protocol Configuration.

## Getting Started

* [Installation](docs/install.md)
* [Usage](docs/usage.md)

## Community Support

In the [Keyfactor Community](https://www.keyfactor.com/community/), we welcome contributions.

The Community software is open-source and community-supported, meaning that **no SLA** is applicable.

* To report a problem or suggest a new feature, go to [Issues](../../issues).
* If you want to contribute actual bug fixes or proposed enhancements, see the [Contributing Guidelines](CONTRIBUTING.md) and go to [Pull requests](../../pulls).

## Commercial Support

Commercial support is available for [EJBCA Enterprise](https://www.keyfactor.com/products/ejbca-enterprise/).

<!--- For SignServer, update to the following text and link:
Commercial support is available for [SignServer Enterprise](https://www.keyfactor.com/products/signserver-enterprise/).
--->

## License

<!--- No updates needed --->
For License information, see [LICENSE](LICENSE).

## Related Projects

See all [Keyfactor EJBCA GitHub projects](https://github.com/orgs/Keyfactor/repositories?q=ejbca).
