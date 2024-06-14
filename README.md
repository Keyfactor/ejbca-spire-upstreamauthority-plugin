
# EJBCA UpstreamAuthority plugin for SPIRE Server

EJBCA UpstreamAuthority plugin for SPIFFE SPIRE using the SPIRE Plugin SDK

#### Integration status: Production - Ready for use in production environments.

## About the Keyfactor API Client

This API client allows for programmatic management of Keyfactor resources.

## Support for EJBCA UpstreamAuthority plugin for SPIRE Server

EJBCA UpstreamAuthority plugin for SPIRE Server is open source and supported on best effort level for this tool/library/client.  This means customers can report Bugs, Feature Requests, Documentation amendment or questions as well as requests for customer information required for setup that needs Keyfactor access to obtain. Such requests do not follow normal SLA commitments for response or resolution. If you have a support issue, please open a support ticket via the Keyfactor Support Portal at https://support.keyfactor.com/

###### To report a problem or suggest a new feature, use the **[Issues](../../issues)** tab. If you want to contribute actual bug fixes or proposed enhancements, use the **[Pull requests](../../pulls)** tab.

---


---



<h1 align="center" style="border-bottom: none">
    EJBCA UpstreamAuthority plugin for SPIRE Server
</h1>

<p align="center">
  <!-- Badges -->
<img src="https://img.shields.io/badge/integration_status-production-3D1973?style=flat-square" alt="Integration Status: production" />
<a href="https://github.com/Keyfactor/ejbca-spiffe-spire-upstreamauthority-plugin/releases"><img src="https://img.shields.io/github/v/release/Keyfactor/ejbca-spiffe-spire-upstreamauthority-plugin?style=flat-square" alt="Release" /></a>
<img src="https://img.shields.io/github/issues/Keyfactor/ejbca-spiffe-spire-upstreamauthority-plugin?style=flat-square" alt="Issues" />
<img src="https://img.shields.io/github/downloads/Keyfactor/ejbca-spiffe-spire-upstreamauthority-plugin/total?style=flat-square&label=downloads&color=28B905" alt="GitHub Downloads (all assets, all releases)" />
</p>

<p align="center">
  <!-- TOC -->
  <a href="#support">
    <b>Support</b>
  </a> 
  ·
  <a href="#license">
    <b>License</b>
  </a>
  ·
  <a href="https://github.com/topics/keyfactor-integration">
    <b>Related Integrations</b>
  </a>
</p>

## Support
The EJBCA UpstreamAuthority plugin for SPIRE Server is open source and there is **no SLA**. Keyfactor will address issues as resources become available. Keyfactor customers may request escalation by opening up a support ticket through their Keyfactor representative. 

> To report a problem or suggest a new feature, use the **[Issues](../../issues)** tab. If you want to contribute actual bug fixes or proposed enhancements, use the **[Pull requests](../../pulls)** tab.


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



## License

Apache License 2.0, see [LICENSE](LICENSE).

## Related Integrations

See all [Keyfactor integrations](https://github.com/topics/keyfactor-integration).


