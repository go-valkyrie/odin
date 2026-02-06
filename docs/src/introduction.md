# Introduction

Welcome to the Odin Guide! Odin is a tool for generating Kubernetes manifests using the CUE language. It provides a simple, type-safe alternative to tools like Helm, Timoni, and KCL.

## What is Odin?

Odin is **not** a package manager and does not apply resources to Kubernetes clusters. It has one job: **generate manifests**. This focused approach makes Odin:

- **Simple** - Easy enough for junior developers to pick up in an hour
- **Type-safe** - Catch configuration errors at build time, not apply time
- **Flexible** - Deploy arbitrary Kubernetes resources with simple abstractions for common cases
- **Integration-friendly** - Works seamlessly with ArgoCD and other GitOps tools

## Why Odin?

Existing Kubernetes templating tools often fall into common traps:

- **Brittle text templates** (Helm) - String interpolation without type safety
- **Poor tool integration** (Timoni, KubeVela) - Don't play nicely with external CD tools
- **Complex abstractions** (KCL, Pkl, ytt) - Steep learning curves for simple tasks

Odin was designed to avoid these pitfalls while providing a powerful foundation for Kubernetes configuration management.

## Part of the Valkyrie Ecosystem

While Odin was designed to work alongside [Valkyrie](https://github.com/go-valkyrie/valkyrie) and [Freyr](https://github.com/go-valkyrie/freyr), it's completely usable standalone. The Freyr project provides reusable component templates that work with any Odin bundle, but you're free to create your own or use Odin without any external templates.

## What You'll Learn

This guide will teach you:

1. **Getting Started** - Install Odin and create your first bundle
2. **Core Concepts** - Understand bundles, components, resources, and values
3. **CLI Reference** - Master all of Odin's commands
4. **Guides** - Learn to write components, use templates, and configure registries
5. **Integration** - Deploy with ArgoCD and build GitOps workflows

Let's get started!
