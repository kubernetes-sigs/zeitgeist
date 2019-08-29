**Zeitgeist** ([/ˈzaɪtɡaɪst/](https://en.wikipedia.org/wiki/Help:IPA/English)) is a language-agnostic dependency checker that keeps track of external dependencies across your project and ensure they're up-to-date.

⚠️ Under active development ⚠️

Rationale
=========

More and more projects nowadays have external dependencies, and the best way to ensure stability and reproducibility is to pin these dependencies to a specific version.

However, this leads to a new problem: the world changes around us, and new versions of these dependencies are released _all the time_.

For a simple project with a couple of dependencies, a team can usually keep up to speed by following mailing lists or Slack channels, but for larger projects this becomes a daunting task.

This problem is pretty much solved by package managers in specific programming languages (see _When is Zeitgeist _not_ suggested_, but it remains a big issue when your project relies on packages outside your programming language of choice, or even more so when declaring infrastructure-as-code, where the "build step" is usually bespoke and dependencies are managed manually.

What is Zeitgeist
=================

Zeitgeist is a tool that takes a configuration file with a list of dependencies, and ensures that:

- These dependencies versions are consistent within your project
- These dependencies are up-to-date

A Zeitgeist configuration file (usually `dependencies.yaml`) is a list of _dependencies_, referenced in files, which may or may not have an _upstream_:

```yaml
dependencies:
- name: terraform
  version: 0.12.3
  upstream:
    flavour: github
    url: hashicorp/terraform
  refPaths:
  - path: helper-image/Dockerfile
    match: TERRAFORM_VERSION
- name: helm
  version: 2.12.2
  upstream:
    flavour: github
    url: helm/helm
    constraints: <3.0.0
  refPaths:
  - path: bootstrap/tiller.yaml
    match: gcr.io/kubernetes-helm/tiller
  - path: helper-image/Dockerfile
    match: HELM_LATEST_VERSION
- name: fluentd-chart
  version: 2.1.1
  upstream:
    flavour: helm
    repo: stable
    name: fluentd
  refPaths:
  - path: helm/fluentbit/requirements.yaml
    match: version
- name: aws-eks-ami
  version: ami-09bbefc07310f7914
  scheme: random
  upstream:
    flavour: ami
    owner: amazon
    name: "amazon-eks-node-1.13-*"
  refPaths:
  - path: clusters.yaml
    match: workers_ami
```

Use `zeitgeist local` to verify that the dependency version is correct in all files referenced in _`refPaths`_.

Use `zeitgeist validate` to also check with defined `upstreams` whether a new version is available for the given dependencies.

When is Zeitgeist _not_ suggested
=================================

While Zeitgeist aims to be a great cross-language solution for tracking external dependencies, it won't be as well integrated as native package managers.

If your project is mainly written in one single language with a well-known and supported package manager (e.g. [`npm`](https://www.npmjs.com/), [`maven`](https://maven.apache.org/), [`rubygems`](https://rubygems.org/), [`pip`](https://pypi.org/project/pip/), [`cargo`](https://crates.io/)...), you definitely should use your package manager rather than Zeitgeist.

Naming
======

[Zeitgeist](https://en.wikipedia.org/wiki/Zeitgeist), a German compound word, can be translated as "spirit of the times" and refers to _a schema of fashions or fads which prescribes what is considered to be acceptable or tasteful for an era_.

Credit
======

Zeitgeist is inspired by [Kubernetes' script to manage external dependencies](https://groups.google.com/forum/?pli=1#!topic/kubernetes-dev/cTaYyb1a18I) and extended to include checking with upstream sources to ensure dependencies are up-to-date.

To do
=====

- [x] Find a good name for the project
- [x] Support `helm` upstream
- [ ] Support `eks` upstream
- [x] Support `ami` upstream
- [x] Cleanly separate various upstreams to make it easy to add new upstreams
- [x] Implement non-semver support (e.g. for AMI, but also for classic releases)
- [x] Write good docs :)
- [x] Write good tests!
- [x] Externalise the project into its own repo
- [ ] Generate releases
- [ ] Test self
