![Golang](https://img.shields.io/badge/Go-1.25-informational)
[![Go Report Card](https://goreportcard.com/badge/github.com/SAP/terraform-provider-scc)](https://goreportcard.com/report/github.com/SAP/terraform-provider-scc)
[![CodeQL](https://github.com/SAP/terraform-provider-scc/actions/workflows/codeql.yml/badge.svg)](https://github.com/SAP/terraform-provider-scc/actions/workflows/codeql.yml)
[![REUSE status](https://api.reuse.software/badge/github.com/SAP/terraform-provider-scc)](https://api.reuse.software/info/github.com/SAP/terraform-provider-scc)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/10659/badge)](https://www.bestpractices.dev/projects/10659)

## About This Project

The Terraform provider for SAP Cloud Connector allows the management of resources on the [SAP Cloud Connector](https://help.sap.com/docs/connectivity/sap-btp-connectivity-cf/cloud-connector) via [Terraform](https://terraform.io/).

You will find the detailed information about the [provider](https://registry.terraform.io/browse/providers) in the official [documentation](https://registry.terraform.io/providers/SAP/scc/latest/docs) in the [Terraform registry](https://registry.terraform.io/).


## Usage of the Provider

Refer to the [Quick Start Guide](./guides/QUICKSTART.md) for instructions to efficiently begin utilizing the Terraform Provider for SAP Cloud Identity Services. For the best experience using the Terraform Provider for SAP Cloud Identity Services, we recommend applying the common best practices for Terraform adoption as described in the [Hashicorp documentation](https://developer.hashicorp.com/well-architected-framework/operational-excellence/operational-excellence-terraform-maturity).

## Developing & Contributing to the Provider

The [developer documentation](DEVELOPER.md) file is a basic outline on how to build and develop the provider.

## Support, Feedback, Contributing

❓ - If you have a *question* you can ask it here in [GitHub Discussions](https://github.com/SAP/terraform-provider-scc/discussions/) or in the [SAP Community](https://answers.sap.com/questions/ask.html).

🐞 - If you find a bug, feel free to create a [bug report](https://github.com/SAP/terraform-provider-scc/issues/new?assignees=&labels=bug%2Cneeds-triage&projects=&template=bug_report.yml&title=%5BBUG%5D).

💡 - If you have an idea for improvement or a feature request, please open a [feature request](https://github.com/SAP/terraform-provider-scc/issues/new?assignees=&labels=enhancement%2Cneeds-triage&projects=&template=feature_request.yml&title=%5BFEATURE%5D).

For more information about how to contribute, the project structure, and additional contribution information, see our [Contribution Guidelines](CONTRIBUTING.md).

> **Note**: We take Terraform's security and our users' trust seriously. If you believe you have found a security issue in the Terraform provider for SAP Cloud Identity Services, please responsibly disclose it. You find more details on the process in [our security policy](https://github.com/SAP/terraform-provider-scc/security/policy).

## Code of Conduct

Members, contributors, and leaders pledge to make participation in our community a harassment-free experience. By participating in this project, you agree to always abide by its [Code of Conduct](https://github.com/SAP/.github/blob/main/CODE_OF_CONDUCT.md).

## Licensing

Copyright 2026 SAP SE or an SAP affiliate company and `terraform-provider-btp` contributors. See our [LICENSE](LICENSE) for copyright and license information. Detailed information, including third-party components and their licensing/copyright information, is available [via the REUSE tool](https://api.reuse.software/info/github.com/SAP/terraform-provider-scc).

## OpenTofu Compatibility

The Terraform Provider for SAP BTP supports [OpenTofu](https://opentofu.org/) under the following conditions:
1. **Drop-In Replacement**: The provider can be used with [OpenTofu CLI](https://opentofu.org/docs/cli/) as a direct replacement for [HashiCorp Terraform CLI](https://developer.hashicorp.com/terraform/cli) without modifications.
2. **Feature Limitations**: The provider does not support OpenTofu specific features or functions outside the standard Terraform functionality.
3. **Issue Reporting**: Any issues reported for the Terraform Provider for SAP BTP will only be addressed if they are reproducible using the Terraform CLI.


## Additional Information and Guides

Through the course of the development of the Terraform provider for SAP BTP and during the constant exchange with customers, several points and questions crossed our path have gathered additional information and guides that might be useful for you. You can find them in the [guides folder](./guides/).
