---
title: "Defining application configuration"
linkTitle: "Defining app configuration"
weight: 2
description: >
  This page describes how to configure your application's deployment using plugins and deploy targets.
---

Each PipeCD application requires a configuration file (typically named `app.pipecd.yaml`) that defines how the application should be deployed. In PipeCD v1, applications are configured using a plugin-based architecture where plugins define the deployment behavior for different platforms.

This page explains how to configure your application to work with plugins and deploy targets.

## Overview

In PipeCD v1, application configuration is platform-agnostic. Instead of specifying a platform type (like Kubernetes or Terraform), you configure which plugin and deploy target to use. The plugin handles all platform-specific deployment logic.

The application configuration file contains:

- **Metadata**: Application name, labels, and other identifying information
- **Plugin configuration**: Which plugin to use and its input parameters
- **Pipeline definition**: Optional stages that customize the deployment process
- **Deploy target**: Which deploy target (defined in `piped` configuration) to use

## Basic application configuration

A minimal application configuration file looks like this:

```yaml
apiVersion: pipecd.dev/v1beta1
kind: Application
spec:
  name: my-application
  labels:
    env: production
    team: backend
  plugins:
    kubernetes:
      deployTarget: kubernetes-prod
      input:
        kubectlVersion: 1.32.2
        manifests:
          - deployment.yaml
          - service.yaml
```

This configuration:

- Names the application `my-application`
- Applies labels for organization and filtering
- Uses the `kubernetes` plugin
- References the `kubernetes-prod` deploy target (defined in your `piped` configuration)
- Specifies the Kubernetes manifests to deploy

## Plugin configuration

The `plugins` section defines which plugin to use and provides plugin-specific input parameters. Each plugin has its own input schema.

### Kubernetes plugin

For Kubernetes applications, configure the `kubernetes` plugin:

```yaml
apiVersion: pipecd.dev/v1beta1
kind: Application
spec:
  name: web-app
  plugins:
    kubernetes:
      deployTarget: k8s-cluster-1
      input:
        kubectlVersion: 1.32.2
        manifests:
          - deployment.yaml
          - service.yaml
          - configmap.yaml
```

**Key fields:**

- `deployTarget`: The name of the deploy target defined in your `piped` configuration. This determines which Kubernetes cluster to deploy to.
- `input.kubectlVersion`: The version of `kubectl` to use for deployments.
- `input.manifests`: List of manifest files relative to the application directory.

**Using Helm charts:**

You can also deploy using Helm charts:

```yaml
apiVersion: pipecd.dev/v1beta1
kind: Application
spec:
  name: helm-app
  plugins:
    kubernetes:
      deployTarget: k8s-cluster-1
      input:
        kubectlVersion: 1.32.2
        helmChart:
          repository: https://charts.example.com
          name: my-chart
          version: 1.2.3
```

**Using Kustomize:**

For Kustomize-based deployments:

```yaml
apiVersion: pipecd.dev/v1beta1
kind: Application
spec:
  name: kustomize-app
  plugins:
    kubernetes:
      deployTarget: k8s-cluster-1
      input:
        kubectlVersion: 1.32.2
        kustomize:
          path: ./kustomize/overlays/production
```

### Terraform plugin

For infrastructure-as-code deployments using Terraform:

```yaml
apiVersion: pipecd.dev/v1beta1
kind: Application
spec:
  name: infrastructure
  plugins:
    terraform:
      deployTarget: aws-production
      input:
        terraformVersion: 1.5.0
        workspace: production
        vars:
          - key: instance_type
            value: t3.large
          - key: region
            value: us-west-2
```

**Key fields:**

- `deployTarget`: The Terraform deploy target defined in your `piped` configuration.
- `input.terraformVersion`: The version of Terraform to use.
- `input.workspace`: The Terraform workspace name.
- `input.vars`: List of Terraform variables to pass during deployment.

## Deploy targets

A deploy target is a named configuration that specifies where and how to deploy your application. Deploy targets are defined in your `piped` configuration file, not in the application configuration.

When you specify a `deployTarget` in your application configuration, you're referencing a deploy target that was configured in your `piped` instance. For example:

**In `piped` configuration:**
```yaml
apiVersion: pipecd.dev/v1beta1
kind: Piped
spec:
  plugins:
    - name: kubernetes
      deployTargets:
        - name: k8s-cluster-1
          config:
            kubeConfigPath: /path/to/kubeconfig
            context: production-cluster
```

**In application configuration:**
```yaml
apiVersion: pipecd.dev/v1beta1
kind: Application
spec:
  plugins:
    kubernetes:
      deployTarget: k8s-cluster-1  # References the deploy target above
      input:
        manifests:
          - deployment.yaml
```

This separation allows you to:

- Reuse the same application configuration across different environments by changing the deploy target
- Keep sensitive connection details (like kubeconfig paths) in the `piped` configuration, not in Git
- Manage multiple clusters or environments from a single `piped` instance

## Pipeline configuration

You can optionally define a deployment pipeline to control how your application is deployed. Pipelines allow you to implement progressive deployment strategies like canary or blue-green deployments.

```yaml
apiVersion: pipecd.dev/v1beta1
kind: Application
spec:
  name: web-app
  plugins:
    kubernetes:
      deployTarget: k8s-cluster-1
      input:
        manifests:
          - deployment.yaml
  pipeline:
    stages:
      - name: K8S_CANARY_ROLLOUT
        with:
          replicas: 10%
      - name: WAIT
        with:
          duration: 5m
      - name: K8S_PRIMARY_ROLLOUT
      - name: K8S_CANARY_CLEAN
```

If you don't specify a pipeline, PipeCD uses Quick Sync, which applies all changes immediately.

For more information about customizing deployment pipelines, see [Customizing an application's deployment pipeline](../customizing-deployment/).

## Quick sync vs pipeline sync

PipeCD supports two deployment strategies:

**Quick Sync** (default when no pipeline is specified):
- Applies all changes immediately
- Fast and straightforward
- Suitable for non-critical changes or development environments

**Pipeline Sync** (when a pipeline is specified):
- Executes deployment through defined stages
- Supports progressive rollout strategies
- Allows manual approvals and automated analysis
- Suitable for production deployments

PipeCD automatically selects Quick Sync when:
- No pipeline is defined in the application configuration
- The changes don't affect workload resources (e.g., only changing replica counts)

## Next steps

- Learn how to [customize your deployment pipeline](../customizing-deployment/)
- See the [configuration reference](../../configuration-reference/) for all available options
- Check out [examples](../../../../examples/) for complete application configuration samples

