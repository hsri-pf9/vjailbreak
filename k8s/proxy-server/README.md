# Kubernetes API Server Extension Layer Proxy

This proxy server acts as an extension layer for the Kubernetes API server. It provides a way to intercept, modify, and forward requests to the Kubernetes API server.

## Features

- Transparent proxying of Kubernetes API requests
- Automatic handling of authentication via kubeconfig or in-cluster config
- Support for both in-cluster and external usage

## Usage

Build and run the proxy server:

```bash
go build -o proxy-server
./proxy-server --kubeconfig=/path/to/kubeconfig --port=8443
```

### Command Line Flags

- `--port`: Port to listen on (default: 8443)
- `--kubeconfig`: Path to kubeconfig file (optional, will use in-cluster config if not provided)
- `--master`: URL of the Kubernetes API server (optional)

## Architecture

The proxy server uses a reverse proxy to forward requests to the Kubernetes API server while maintaining authentication. It can be extended to add custom request/response handling logic.
