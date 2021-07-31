#!/bin/bash

chart_version="1.4.0"

KUBECTL=kubectl

cert-manager_install () {
  echo -e "################################################################\n####### install cert-manager \n################################################################"

	helm repo add cert-manager https://charts.jetstack.io

	helm install cert-manager cert-manager/cert-manager --version "${chart_version}" \
	   --namespace cert-manager \
	   --create-namespace \
	   --values helm-values/cert-manager.yaml

	# wait to cert-manager up and running
	"$KUBECTL" wait -n cert-manager --timeout=2m --for=condition=available deployment cert-manager || true
	"$KUBECTL" wait -n cert-manager --timeout=2m --for=condition=available deployment cert-manager-webhook || true
	"$KUBECTL" wait -n cert-manager --timeout=2m --for=condition=available deployment cert-manager-cainjector || true

}

time cert-manager_install

# enable scheduling workload on master
kubectl taint nodes --all node-role.kubernetes.io/master- || true
