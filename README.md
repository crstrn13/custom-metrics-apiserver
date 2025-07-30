# Custom Metrics APIServer

A Kubernetes custom metrics API server implementation that tracks HTTP request counts. It exposes a simple `/get` endpoint and provides these request counts as metrics that can be used by the Kubernetes HPA (Horizontal Pod Autoscaler).

## Overview

This custom metrics adapter:
- Tracks HTTP request counts per pod
- Exposes metrics through the Kubernetes custom metrics API
- Provides a simple `/get` endpoint that increments the counter
- Integrates with HPA for autoscaling based on request counts

## Features

- Automatic request counting for each pod
- Custom metrics API implementation for Kubernetes HPA
- Simple `/get` endpoint for testing
- Kubernetes-native metrics reporting

## Getting Started

### Prerequisites

- Go 1.24 or later
- Kubernetes cluster with metrics API enabled
- kubectl access to your cluster

### Building

```bash
go build -o custom-metrics-apiserver cmd/main.go
```

### Running Locally

```bash
./custom-metrics-apiserver
```

The server will start on port 8080.

### Deploying to Kubernetes

1. Apply the deployment and RBAC configuration:
```bash
kubectl apply -f deploy/k8s/deployment.yaml
```

2. Verify the installation:
```bash
kubectl get apiservice v1beta1.custom.metrics.k8s.io
```

## Usage

### Testing Request Counting

Make requests to increment the counter:
```bash
curl http://your-service/get
```

### Checking Metrics

View the metrics for a pod:
```bash
kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1/namespaces/default/pods/*/http_requests_total"
```

### HPA Configuration

Example HPA configuration using the custom metric:
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: my-app
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: my-app
  minReplicas: 1
  maxReplicas: 10
  metrics:
  - type: Object
    object:
      metric:
        name: http_requests_total
      describedObject:
        apiVersion: v1
        kind: Pod
        name: my-app-pod
      target:
        type: Value
        value: 10
```

## License

This project is licensed under the MIT License - see the LICENSE file for details. 