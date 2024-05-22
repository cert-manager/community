# cert-manager Project Roadmap


While this is a summary of the direction we want to go we welcome pull requests across all projects, even if they don't
fall under any of the roadmap items listed here.

We unfortunately can't merge every change, and if you're looking to contribute a new feature you might want to
check the [contributing guide](https://cert-manager.io/docs/contributing/) on the cert-manager website and specifically
our [feature policy](https://cert-manager.io/docs/contributing/policy/) for details on what we're likely to accept and
how to maximise your chances of being accepted.

## Changes to this Roadmap

Changes to this roadmap will take the form of pull requests containing the suggested change. All such PRs must be posted to the `#cert-manager-dev` Slack channel in
Kubernetes slack so that they're made visible to all other developers and maintainers and steering committee members.

Significant changes to this document should be discussed in either a [steering committee](./STEERING.md) meeting or a [biweekly meeting](https://cert-manager.io/docs/contributing/#meetings)
before merging, to raise awareness of the change and to provide an opportunity for discussion. A significant change is one which meaningfully alters
one of the roadmap items, adds a new item, or removes an item.

Insignificant changes include updating links to issues, spelling fixes or minor rewordings which don't significantly change meanings. These insignificant changes
don't need to be discussed in a meeting but should still be shared in Slack.

## Goals and Objectives of the cert-manager Project

Our primary goal is to ensure that the following statement holds true:

> cert-manager is the easiest way to automatically manage certificates in Kubernetes clusters.

This applies to cert-manager itself, and to other subprojects which are part of the wider cert-manager project
including:

- trust-manager
- csi-driver
- csi-driver-spiffe
- istio-csr
- approver-policy

## Roadmap Items

### Adoption of Upstream Changes

Continue to support latest versions, new APIs for upstream Kubernetes and related upstream projects.

For example:

- Kubernetes APIs: keep up to date with Kubernetes API changes
- Kubernetes Versions: maintain support for a reasonable spread of Kubernetes versions, and support new versions as quickly as we can
- [ClusterTrustBundle](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#clustertrustbundle-v1alpha1-certificates-k8s-io) support
- Gateway API

#### Goals for cert-manager `v1.15`

- [x] Graduate Gateway API support to Beta ([#6961](https://github.com/cert-manager/cert-manager/pull/6961))
- [x] Support Kubernetes `v1.30` ([#7006](https://github.com/cert-manager/cert-manager/issues/7006))

### Shrinking Core / Aiding Extensibility

Minimise the surface area of cert-manager's core components to reduce attack surface, binary size, container size and complexity.

Nothing in this item should make cert-manager harder to use for any user. Mostly it talks about code architecture changes to improve
security, reduce blast radius and more tightly scope powerful permissions to only those components which need them.

For example:

- Move "core" issuers with dependencies (ACME, Vault, Venafi) into external issuers, which would be still be included in the cert-manager Helm chart but would run separately
- Likewise, change all "core" DNS solvers into external solvers
- Focus on good integration points for external plugins rather than adding things into core

### PKI Lifecycle

Enable best-practice PKI management with cert-manager. Users with existing PKI deployments should be able to move them into cert-manager, while
users getting started with PKI should be able to use cert-manager to achieve their goals.

For example:

- Handle CA certs being renewed: deal with the cases where the CA cert is renewed and allow for all signed certs to be renewed
- Make cert-manager a viable way to create and manage private PKI deployments at scale
- Handle trust in a safe way, wherever it needs to be used

## Notes / See Also

- This file is based on a [CNCF document](https://contribute.cncf.io/maintainers/community/contributor-growth-framework/open-source-roadmaps)
- That document has vendor neutrality at its core, and this roadmap is [vendor neutral](https://contribute.cncf.io/maintainers/community/vendor-neutrality/)
- cert-manager's [Supported Releases](https://cert-manager.io/docs/releases/) page detailing upcoming versions and Kubernetes version support
