#!/bin/bash

# Kubernetes bootstrapping script for cryptogen service

# Print message and exit with code 1
function fatal() {
    echo $@
    exit 1
}

current_namespace() {
  ns="$(kubectl config view -o=jsonpath="{.contexts[?(@.name==\"${1}\")].context.namespace}")"
  if [ "$ns" == "" ]; then
    echo "default"
  else
    echo "${ns}"
  fi
}

# We need kubectl and helm
which kubectl > /dev/null || fatal "kubectl is required"
which helm > /dev/null || fatal "helm is required"

# Some vars
DIR=`dirname $0`
CTX=`kubectl config current-context`
NS=`current_namespace $CTX`

echo "Context is '$CTX', namespace '$NS'. (Enter to continue, Ctrl-C to abort)"
read

# Ensure tiller is running in the cluster. In a real production scenario we'd not use helm init,
# rather install it from kube manifests along with dedicated service accounts, per namespace, etc.
tiller=`kubectl get pod -n kube-system -l app=helm,name=tiller 2>&1 -o name | grep tiller-deploy`
if [ "$tiller" == "" ] ; then
    helm init
    sleep 5
fi

helm repo add stable http://storage.googleapis.com/kubernetes-charts
helm repo add banzaicloud-stable http://kubernetes-charts.banzaicloud.com/branch/master

# Install Vault using Banzaiclouds awesome operator to simplify configuration, automate unsealing etc.
# See https://github.com/banzaicloud/bank-vaults
# 
# To keep it simple, we're using filesystem storage and a bunch of insecure settings:
# map kube default SA to a very broad policy, store vault root token and unseal keys in kube secrets
# 
# In real life we'd need more restrictive policies, specialised SA's and vault role mappings, 
# use Consul or S3 as a backend and store admin tokens and unseal keys in AWS KMS or similar.

helm upgrade -i \
    --kube-context $CTX \
    -f $DIR/vault.yaml \
    --version 0.5.12 \
    vault banzaicloud-stable/vault

helm upgrade -i \
    --kube-context $CTX \
    --set usePassword=false \
    --version 4.1.0 \
    redis stable/redis
