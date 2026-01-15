---
title: "Customizing an application's deployment pipeline"
linkTitle: "Customizing deployment"
weight: 3
description: >
  This page describes how to customize an application's deployment pipeline with PipeCD stages.
---

A deployment pipeline defines the sequence of stages that PipeCD executes when deploying your application. By default, PipeCD uses Quick Sync, which applies all changes immediately. However, you can customize the deployment process by defining a pipeline with stages that implement progressive deployment strategies, manual approvals, automated analysis, and more.

This page explains how to build custom deployment pipelines using PipeCD stages.

## Overview

A pipeline is defined in your application configuration file using the `pipeline` field. Each pipeline consists of one or more stages that execute in sequence:

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

![Example deployment with a WAIT stage](/images/deployment-wait-stage.png)
<p style="text-align: center;">
Example deployment with a WAIT stage
</p>

## Plugin-specific stages

Each plugin provides stages specific to its platform. These stages handle the core deployment operations for that platform.

### Kubernetes plugin stages

The Kubernetes plugin provides the following stages:

- **`K8S_PRIMARY_ROLLOUT`**: Updates the primary (stable) resources to the state defined in the target commit.
- **`K8S_CANARY_ROLLOUT`**: Creates canary resources based on the primary resource definition in the target commit.
- **`K8S_CANARY_CLEAN`**: Removes all canary resources after a successful deployment.
- **`K8S_BASELINE_ROLLOUT`**: Creates baseline resources for comparison during analysis.
- **`K8S_BASELINE_CLEAN`**: Removes baseline resources after analysis completes.
- **`K8S_TRAFFIC_ROUTING`**: Splits traffic between variants (canary, baseline, primary).

**Example: Canary deployment**

```yaml
apiVersion: pipecd.dev/v1beta1
kind: Application
spec:
  name: canary-app
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
          replicas: 20%
      - name: WAIT
        with:
          duration: 10m
      - name: K8S_TRAFFIC_ROUTING
        with:
          canary: 20%
          primary: 80%
      - name: WAIT
        with:
          duration: 30m
      - name: K8S_PRIMARY_ROLLOUT
      - name: K8S_CANARY_CLEAN
```

This pipeline:
1. Deploys 20% of replicas as canary
2. Waits 10 minutes
3. Routes 20% of traffic to canary
4. Waits 30 minutes for observation
5. Rolls out the new version to primary
6. Cleans up canary resources

### Terraform plugin stages

The Terraform plugin provides:

- **`TERRAFORM_PLAN`**: Generates an execution plan showing what changes will be made.
- **`TERRAFORM_APPLY`**: Applies the Terraform configuration.

**Example: Terraform deployment with plan review**

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
  pipeline:
    stages:
      - name: TERRAFORM_PLAN
      - name: WAIT_APPROVAL
      - name: TERRAFORM_APPLY
```

## Common stages

These stages are available across all plugins and can be used to enhance any deployment pipeline.

### WAIT stage

The `WAIT` stage pauses the deployment for a specified duration. This is useful for:
- Allowing time for new deployments to stabilize
- Observing metrics before proceeding
- Creating delays between rollout phases

```yaml
pipeline:
  stages:
    - name: K8S_CANARY_ROLLOUT
      with:
        replicas: 50%
    - name: WAIT
      with:
        duration: 15m
    - name: K8S_PRIMARY_ROLLOUT
```

**Configuration:**

- `duration`: The amount of time to wait. Supports formats like `5m`, `30s`, `1h`.

### WAIT_APPROVAL stage

The `WAIT_APPROVAL` stage pauses the deployment and waits for manual approval before continuing. This is useful for:
- Reviewing deployment plans before applying changes
- Requiring stakeholder sign-off for production deployments
- Implementing change control processes

```yaml
pipeline:
  stages:
    - name: TERRAFORM_PLAN
    - name: WAIT_APPROVAL
    - name: TERRAFORM_APPLY
```

When a deployment reaches a `WAIT_APPROVAL` stage:
1. The deployment pauses and shows as "Waiting Approval" in the UI
2. Users with appropriate permissions can approve or reject the deployment
3. If approved, the pipeline continues
4. If rejected, the deployment is cancelled

**Configuration:**

The `WAIT_APPROVAL` stage doesn't require any configuration. Approval is handled through the PipeCD web UI.

### ANALYSIS stage

The `ANALYSIS` stage performs automated analysis of your deployment by evaluating metrics, logs, and HTTP requests. If the analysis fails, the deployment can be automatically rolled back.

```yaml
pipeline:
  stages:
    - name: K8S_CANARY_ROLLOUT
      with:
        replicas: 50%
    - name: ANALYSIS
      with:
        duration: 10m
        metrics:
          - provider: prometheus
            expected:
              max: 0.05
            query: |
              rate(http_requests_total{status="5xx"}[1m])
    - name: K8S_PRIMARY_ROLLOUT
```

**Configuration:**

- `duration`: How long to run the analysis
- `metrics`: List of metric queries to evaluate
- `logs`: List of log queries to evaluate
- `http`: List of HTTP endpoints to test

For detailed information about configuring analysis, see [Adding an automated deployment analysis stage](./automated-deployment-analysis/).

### SCRIPT_RUN stage

The `SCRIPT_RUN` stage executes custom scripts or commands during deployment. This allows you to:
- Run custom validation scripts
- Perform pre or post-deployment tasks
- Integrate with external tools

```yaml
pipeline:
  stages:
    - name: K8S_CANARY_ROLLOUT
      with:
        replicas: 50%
    - name: SCRIPT_RUN
      with:
        env:
          ENVIRONMENT: production
        run: |
          ./scripts/validate-deployment.sh
          ./scripts/notify-slack.sh
    - name: K8S_PRIMARY_ROLLOUT
```

**Configuration:**

- `env`: Environment variables to set for the script
- `run`: The script or command to execute

> **Note:** The `SCRIPT_RUN` stage is in alpha status. Rollback support is currently limited to Kubernetes applications.

## Building a complete pipeline

Here's an example of a complete production-ready pipeline that combines multiple stages:

```yaml
apiVersion: pipecd.dev/v1beta1
kind: Application
spec:
  name: production-app
  plugins:
    kubernetes:
      deployTarget: k8s-production
      input:
        manifests:
          - deployment.yaml
          - service.yaml
  pipeline:
    stages:
      # Deploy canary with 10% of replicas
      - name: K8S_CANARY_ROLLOUT
        with:
          replicas: 10%
      
      # Wait for canary to be ready
      - name: WAIT
        with:
          duration: 2m
      
      # Route 10% of traffic to canary
      - name: K8S_TRAFFIC_ROUTING
        with:
          canary: 10%
          primary: 90%
      
      # Run automated analysis
      - name: ANALYSIS
        with:
          duration: 15m
          metrics:
            - provider: prometheus
              expected:
                max: 0.01
              query: |
                rate(http_requests_total{status="5xx"}[1m])
      
      # Wait for manual approval before full rollout
      - name: WAIT_APPROVAL
      
      # Rollout to primary
      - name: K8S_PRIMARY_ROLLOUT
      
      # Clean up canary resources
      - name: K8S_CANARY_CLEAN
```

This pipeline implements a safe, progressive deployment with:
- Gradual rollout (10% canary)
- Traffic splitting
- Automated health checks
- Manual approval gate
- Automatic cleanup

## Best practices

When building deployment pipelines:

1. **Start small**: Begin with simple pipelines and add complexity as needed
2. **Use analysis**: Always include analysis stages for production deployments
3. **Set appropriate wait times**: Give deployments time to stabilize between stages
4. **Test pipelines**: Test your pipeline configuration in non-production environments first
5. **Document stages**: Add comments or documentation explaining why each stage is needed

## Next steps

- Learn about [adding a wait stage](./adding-a-wait-stage/)
- Configure [automated deployment analysis](./automated-deployment-analysis/)
- Set up [manual approval stages](./adding-a-manual-approval/)
- See the [configuration reference](../../configuration-reference/) for all stage options

