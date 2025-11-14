# Deployment

Kthulu can be deployed for production either on a single host using Docker
Compose or to a Kubernetes cluster using Kustomize overlays and the provided
Helm chart.

## Docker Compose

For a minimal standalone deployment, use the production compose file:

```sh
docker compose -f docker-compose.prod.yml up -d
```

## Kubernetes with Kustomize

Base manifests live under `kustomize/base` and environment-specific overlays are
in `kustomize/overlays`. To deploy the development overlay to a cluster:

```sh
kustomize build kustomize/overlays/dev | kubectl apply -f -
```

## Kubernetes with Helm

A Helm chart is available under `deploy/helm/kthulu`. Install or upgrade it in a
namespace:

```sh
helm upgrade --install kthulu deploy/helm/kthulu -n kthulu --create-namespace
```

