# Mock Backend

A simple HTTP server that implements the Kubernetes custom metrics API. It exposes a `/get` endpoint similar to httpbin and provides request count metrics that can be used by the Kubernetes HPA (Horizontal Pod Autoscaler).

## Features

- `/get` endpoint that returns request information (similar to httpbin)
- Custom metrics API implementation for Kubernetes HPA
- HTTPS support with auto-generated self-signed certificates
- Request counting metrics

## Getting Started

### Prerequisites

- Go 1.24 or later
- Kubernetes cluster with metrics API enabled

### Building

```bash
go build -o mock-backend cmd/main.go
```

### Running Locally

```bash
./mock-backend
```

The server will start on port 8080 with HTTPS enabled.

### Deploying to Kubernetes

1. Build and push the Docker image:
```bash
docker build -t your-registry/mock-backend:latest .
docker push your-registry/mock-backend:latest
```

2. Update the image in `deploy/k8s/deployment.yaml` and apply:
```bash
kubectl apply -f deploy/k8s/deployment.yaml
```

## API Endpoints

### GET /get

Returns information about the incoming request, including:
- Headers
- Query parameters
- Origin IP
- URL

### GET /apis/custom.metrics.k8s.io/v1beta1

Returns request count metrics in the Kubernetes custom metrics API format.

## License

This project is licensed under the MIT License - see the LICENSE file for details. 