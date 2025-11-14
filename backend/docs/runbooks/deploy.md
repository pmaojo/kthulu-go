# Rolling Update Deployment Runbook

This guide outlines the procedure for performing rolling updates of the Kthulu backend on Kubernetes.

## Prerequisites
- Access to the Kubernetes cluster with appropriate permissions.
- The `kthulu` Deployment is installed.
- kubectl configured for the target cluster.

## Steps
1. **Prepare the new container image** and push it to the registry.
2. **Deploy the update** using the Helm chart or kustomize overlay:
   ```bash
   helm upgrade --install kthulu deploy/helm/kthulu \
     --set image.repository=<registry>/kthulu \
     --set image.tag=<tag>
   # or
   make deploy-k8s IMAGE_REPOSITORY=<registry>/kthulu IMAGE_TAG=<tag>
   ```
3. **Monitor the rollout**:
   ```bash
   kubectl rollout status deployment/kthulu
   ```
4. **Verify health probes** using the service endpoint to ensure new pods are ready before old ones terminate.
5. If needed, **roll back** the deployment:
   ```bash
   kubectl rollout undo deployment/kthulu
   ```

This rolling strategy ensures zero-downtime updates by gradually replacing pods while leveraging readiness and liveness probes.
