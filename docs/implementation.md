# Implementation Details

- Golang, so we can use native vault client
- Generate crypto config yaml from input (template)
- `cryptogen` must be provided as binary (see `Dockerfile`)
- Exec cryptogen, generate assets
- Read generated assets from disk, store into vault

## Vault 

- Secrets are stored in vault
- App is authenticating to vault using Kubernetes auth (via init container)
- In a kube deployment they could be retrieved by other pods on the same cluster

## Persistence

- Used to persist job status and vault paths to assets
- Currently only Redis supported
- Designed to be swappable, could be CouchDB, DynamoDB, etc.
