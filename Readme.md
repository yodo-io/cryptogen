# Cryptogen

## Setup Draft

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

## Release 

When using minikube, make sure we're using minikubes registry:

```sh
eval $(minikube docker-env)
```

Release:

```sh
draft up
```

## Accessing Vault

```sh
VAULT_TOKEN=`kubectl get secret bank-vaults -o jsonpath='{.data.vault-root}' | base64 -D`
export VAULT_TOKEN
export VAULT_ADDR=https://localhost:8200
export VAULT_SKIP_VERIFY="true"
```
