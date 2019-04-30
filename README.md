# vrddt

Project skeleton and general architecture from: [spy16/droplets](https://github.com/spy16/droplets)

## TODO

    - Research and implement context correctly
    - CMD
        - ADMIN
        - API
            - Authorization / OAuth
            - Rate limiting
        - CLI
            - Get Metadata
        - WEB
            - Authorization / OAuth
            - Rate limiting
        - WATCHER
        - WORKER
    - INTERNALS
        - API Address needs to be sorted out where the Address can be anything local or remote
        - Makefile/Dockerfile/docker-compose.yml refactor for DRY
        - Add S3 storage support
        - Implement other video types for video processor
            - Breakout Upload feature and vrddt video association so it can be
            used again by other types.

## Kubernetes setup

### Traefik

```shell
helm init
kubectl create serviceaccount --namespace kube-system tiller
kubectl create clusterrolebinding tiller-cluster-rule --clusterrole=cluster-admin --serviceaccount=kube-system:tiller
kubectl patch deploy --namespace kube-system tiller-deploy -p '{"spec":{"template":{"spec":{"serviceAccount":"tiller"}}}}'

# Loop until tiller is "Running"
kubectl get pods -n kube-system

helm install stable/traefik --name traefik --set dashboard.enabled=true,serviceType=NodePort,dashboard.domain=dashboard.traefik,rbac.enabled=true,ssl.enabled=true,ssl.enforced=true --namespace kube-system

kubectl describe svc traefik --namespace kube-system
```
