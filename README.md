# CRI Test Compatibility Matrix

This branch contains auto-generated compatibility tables showing which CRI conformance
tests pass, skip, or fail across different runtime configurations.

## Tables

- [containerd compatibility matrix](compatibility-matrix-containerd.md) — results across
  containerd versions, operating systems, shims, and OCI runtimes
- [CRI-O compatibility matrix](compatibility-matrix-crio.md) — results across
  CRI-O OCI runtime and monitor configurations

## How it works

These files are automatically updated on every push to `master` by the CI workflows
in `.github/workflows/containerd.yml` and `.github/workflows/crio.yml`. Each file
embeds the commit SHA from which the results were generated, so staleness is visible
if a CI run fails.
