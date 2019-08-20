Zeitgeist
=========

You have dependencies. You need to track them.

This (heavily IN PROGRESS) folder contains code forked from [Kubernetes' script to manage dependencies](https://groups.google.com/forum/?pli=1#!topic/kubernetes-dev/cTaYyb1a18I) and extended to include checking with upstream sources to ensure dependencies are up-to-date.

To do
=====

[x] Find a good name for the project
[ ] Support `helm` upstream
[ ] Support `eks` upstream
[ ] Support `ami` upstream
[x] Cleanly separate various upstreams to make it easy to add new upstreams
[ ] Implement non-semver support (e.g. for AMI, but also for classic releases)
[ ] Write good docs :)
[ ] Write good tests!
[ ] Externalise the project into its own repo & generate releases. Test self.
