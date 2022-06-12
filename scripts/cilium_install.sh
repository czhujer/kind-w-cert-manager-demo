#!/bin/bash

cilium_version="1.11.5"

KUBECTL=kubectl

cilium_install () {
  echo -e "################################################################\n####### install cilium \n################################################################"

	helm repo add cilium https://helm.cilium.io/

	helm upgrade --install \
	  cilium cilium/cilium --version "${cilium_version}" \
	   --namespace kube-system \
	   --set nodeinit.enabled=true \
	   --set kubeProxyReplacement=partial \
	   --set hostServices.enabled=false \
	   --set externalIPs.enabled=true \
	   --set nodePort.enabled=true \
	   --set hostPort.enabled=true \
	   --set bpf.masquerade=false \
	   --set image.pullPolicy=IfNotPresent \
	   --set ipam.mode=kubernetes

	# wait to cilium pods (and nodes) ready
  "$KUBECTL" -n kube-system rollout status --watch --timeout=1m  daemonset.apps/cilium || true

	# wait to coredns up and running
	"$KUBECTL" -n kube-system rollout status --watch --timeout=1m  deployments.apps coredns || true
	"$KUBECTL" -n kube-system wait --for=condition=available --timeout=1m deployments.apps coredns || true
}

cilium_install

# enable scheduling workload on master
"$KUBECTL" taint nodes --all node-role.kubernetes.io/master- || true
