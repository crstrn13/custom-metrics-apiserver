# Custom Metrics APIServer

A minimal Kubernetes custom metrics adapter that allows you to write and expose any custom metric through the Kubernetes custom metrics API. This makes your metrics available to HPA (Horizontal Pod Autoscaler) for scaling decisions.

## Overview

This adapter:
- Implements the Kubernetes custom metrics API discovery endpoints
- Provides a `/write-metrics` endpoint to write any custom metric
- Exposes metrics through the custom metrics API (`/apis/custom.metrics.k8s.io/v1beta1/...`)
- Makes metrics available to Kubernetes HPA for scaling decisions

## Getting Started

### Prerequisites

- Go 1.20 or later
- Kubernetes cluster
- kubectl access to your cluster

### Installation

1. Apply the RBAC rules, API service registration, and deployment:
```bash
kubectl apply -f deploy/k8s/deployment.yaml
```

2. Verify the custom metrics API is registered:
```bash
kubectl get apiservice v1beta1.custom.metrics.k8s.io
```

## Usage

### Writing Metrics

Write any custom metric value:
```bash
# Get the service URL
SERVICE_URL=$(kubectl -n custom-metrics get service custom-metrics-apiserver -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')

# Write a custom metric
curl -X POST \
  http://$SERVICE_URL/write-metrics/namespaces/default/pods/my-pod/my_custom_metric \
  -d '"42"'

# Write another metric
curl -X POST \
  http://$SERVICE_URL/write-metrics/namespaces/default/pods/my-pod/requests_per_second \
  -d '"100"'
```

### API Discovery

The adapter implements the standard Kubernetes API discovery endpoints:
```bash
# List all available metrics
kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1"

# List metrics for a specific pod
kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1/namespaces/default/pods/my-pod/"
```

### Viewing Metrics

Check the current value of your custom metrics:
```bash
# Get a specific metric
kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1/namespaces/default/pods/*/my_custom_metric"

# Get another metric
kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1/namespaces/default/pods/*/requests_per_second"
```

### HPA Configuration

Example HPA that scales based on a custom metric:
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
        name: requests_per_second  # Your custom metric name
      describedObject:
        apiVersion: v1
        kind: Pod
        name: my-app-pod
      target:
        type: Value
        value: 100
```

This will scale the deployment based on your custom metric value.

## How it Works

1. The adapter registers itself with the Kubernetes API server via the APIService resource
2. It implements the custom metrics API discovery endpoints
3. You write metrics using the `/write-metrics` endpoint with any metric name you choose
4. These metrics are stored in memory and exposed through the custom metrics API
5. HPA queries these metrics through the Kubernetes API server
6. The adapter responds with the stored metric values
7. HPA uses these values to make scaling decisions

## API Endpoints

### Write Metrics
- `POST /write-metrics/namespaces/{namespace}/{resourceType}/{name}/{metric}`
  - For pod metrics: `/write-metrics/namespaces/default/pods/my-pod/my_custom_metric`
  - Body should be a JSON-encoded Kubernetes quantity string (e.g. `"42"`)
  - You can use any metric name in place of `my_custom_metric`

### Custom Metrics API
- `GET /apis/custom.metrics.k8s.io/v1beta1/` - API discovery
- `GET /apis/custom.metrics.k8s.io/v1beta1/namespaces/{namespace}/pods/*/my_custom_metric` - Get pod metrics

## License

MIT License - see the LICENSE file for details. 