# Cryptogen

## Setup Draft

Draft is used to simplify development workflow against k8s. It conveniently generates a helm chart which could later be used for deployment.

Via Homebrew:

```sh
brew install draft
```

Or download:

```sh
# or curl
https://azuredraft.blob.core.windows.net/draft/draft-$(draft_rel)-$(draft_arch).tar.gz
# ...untar and copy to path
```

Init draft locally (not sure it's needed on an already bootstrapped project):

```sh
draft init
```

## Running 

When using minikube, make sure we're using minikubes registry:

```sh
eval $(minikube docker-env)
```

Release to test cluster:

```sh
draft up
```

## Testing

With draft and the same kube context still active, use `test.sh` for a demo

## Implementation & Design

See [design document](./docs/cryptogen.md)

## Misc

### Accessing Vault

```sh
VAULT_TOKEN=`kubectl get secret bank-vaults -o jsonpath='{.data.vault-root}' | base64 -D`
export VAULT_TOKEN
export VAULT_ADDR=https://localhost:8200
export VAULT_SKIP_VERIFY="true
```

### Container Logs

```sh
kubectl logs -f `kubectl get pod -l draft=cryptogen -o name | awk -F'/' '{ print $2 }'`
```