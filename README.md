<!-- markdownlint-disable MD041 -->

[![FINOS - Incubating](https://cdn.jsdelivr.net/gh/finos/contrib-toolbox@master/images/badge-incubating.svg)](https://finosfoundation.atlassian.net/wiki/display/FINOS/Incubating)

<!-- markdownlint-enable MD041 -->

<a href="https://ccc.finos.org"><img height="100px" src="https://github.com/finos/branding/blob/master/project-logos/active-project-logos/FINOS%20Common%20Cloud%20Controls%20Logo/Horizontal/2023_FinosCCC_Horizontal.svg?raw=true" alt="CCC Logo"/></a>

# FINOS Common Cloud Controls : Compliant Financial Infrastructure

This repository contains Terraform modules and configuration examples for creating secure, compliant cloud storage solutions that align with the [FINOS Common Cloud Controls (CCC)](https://ccc.finos.org) standard.

## What Is It?

- **Secure by Default**: Terraform modules that implement CCC security controls out of the box
- **Multi-Cloud Support**: Configurations for AWS S3, Azure Storage, and Google Cloud Storage
- **Production Ready**: Battle-tested configurations suitable for financial services environments
- **Compliance Focused**: Each configuration maps to specific CCC controls and requirements

## How To Use It

### 1. Configuration Examples

Browse the `/config` directory for ready-to-use configuration example for aws, azure or gcp.

### 2. CCC Controls Implementation

For the complete list of controls and their implementation details, see the [CCC Standard](https://ccc.finos.org).

You can review the results of testing the above configurations on the [CCC Website](ccc.finos.org/cfi)

### 3. Compliance Testing

The testing system:

- **Discovers resources** automatically using cloud provider APIs
- **Runs Gherkin tests** filtered by catalog type (CCC.ObjStor, CCC.Core, etc.)
- **Generates reports** in HTML and OCSF JSON formats

See the [Testing README](testing/README.md) for full documentation on architecture, adding new services, and writing tests.

## How To Contribute

### 1. Improving or Contributing CFI Code

- Check [the issues](https://github.com/finos-labs/ccc-cfi-compiance/issues) to see if there's anything you'd like to work on
- [Raise a GitHub Issue](https://github.com/finos-labs/ccc-cfi-compliance/issues/new/choose) to ask questions or make suggestions
- Pull Requests are always welcome - the main branch is considered an iterative development branch

### 2. Join FINOS CCC Project Meetings

This project is part of the broader CCC initiative. Join the **Compliant Financial Infrastructure** working group:

- **When**: 10AM UK Thursday / 5PM UK on 4th Thursday each month
- **See**: calendar.finos.org
- **Chair**: @eddie-knight
- **Mailing List**: [cfi+subscribe@lists.finos.org](mailto:cfi+subscribe@lists.finos.org)

Find meetings on the [FINOS Community Calendar](https://finos.org/calendar) and browse [Past Meeting Minutes](https://github.com/finos/common-cloud-controls/labels/meeting).

### 3. DCO Required

#### Using DCO to sign your commits

All commits must be signed with a DCO signature to avoid being flagged by the DCO Bot. This means that your commit log message must contain a line that looks like the following one, with your actual name and email address:

```
Signed-off-by: John Doe <john.doe@example.com>
```

Adding the `-s` flag to your `git commit` will add that line automatically. You can also add it manually as part of your commit log message or add it afterwards with `git commit --amend -s`.

#### Helpful DCO Resources

- [Git Tools - Signing Your Work](https://git-scm.com/book/en/v2/Git-Tools-Signing-Your-Work)
- [Signing commits
  ](https://docs.github.com/en/github/authenticating-to-github/signing-commits)

## License

Copyright 2025 FINOS

Distributed under the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0).

SPDX-License-Identifier: [Apache-2.0](https://spdx.org/licenses/Apache-2.0)

## Security

Please see our [Security Policy](SECURITY.md) for reporting vulnerabilities.

## Code of Conduct

Participants should follow the FINOS Code of Conduct: <https://community.finos.org/docs/governance/code-of-conduct>
