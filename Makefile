.PHONY: kubetest-up
kubetest-up:
	kubetest2 kind --config kind-config2-without-cni.yaml --up --stderrthreshold 5

.PHONY: kubetest-cilium-install
kubetest-cilium-install:
	kubetest2 kind --test exec -- sh -c scripts/cilium_install.sh

.PHONY: kubetest-exec-shellspec
kubetest-exec-shellspec:
	kubetest2 kind --test exec -- shellspec --default-path tests/shellspec --load-path tests/shellspec -o j -f d --reportdir ./ tests/shellspec/cert-manager-test-csi-driver_spec.sh

.PHONY: kubetest-exec-ginkgo
kubetest-exec-ginkgo:
	cd tests/e2e && kubetest2 kind --test exec -- ginkgo -v

.PHONY: kubetest-down
kubetest-down:
	kubetest2 kind --down

.PHONY: kubetest-all-w-shellspec
kubetest-all: kubetest-up kubetest-cilium-install kubetest-exec-shellspec kubetest-down

.PHONY: kubetest-all-w-ginkgo
kubetest-all: kubetest-up kubetest-cilium-install kubetest-exec-ginkgo kubetest-down

.PHONY: deploy-cert-manager
deploy-cert-manager:
	scripts/cert-manager_install.sh

.PHONY: deploy-cert-manager-csi-driver
deploy-cert-manager-csi-driver:
	kubectl apply -f tests/assets/cert-manager-csi-driver.yaml
	kubectl apply -f tests/assets/cert-manager-csi-driver-example-app.yaml

.PHONY: deploy-cert-manager-test-ca
deploy-cert-manager-test-ca:
	kubectl apply -f tests/assets/cert-manager-issuer-kind-test.yaml
	kubectl apply -f tests/assets/cert-manager-issuer-kind-ca-test.yaml
	kubectl apply -f tests/assets/cert-manager-certificate-test1.yaml
	kubectl apply -f tests/assets/cert-manager-certificate-test2.yaml

.PHONY: run-shellspec
run-shellspec:
	shellspec --default-path tests/shellspec --load-path tests/shellspec -o j -f d --reportdir ./ tests/shellspec/cert-manager-test-csi-driver_spec.sh

.PHONY: run-ginkgo
run-ginkgo:
	cd tests/e2e && ginkgo

.PHONY: get-kubetest2-latest
get-kubetest2-latest:
	wget -q https://storage.googleapis.com/k8s-staging-kubetest2/latest/linux-amd64.tgz -O kubetest2.tgz
	mkdir kubetest2
	tar -xzf kubetest2.tgz -C kubetest2
	sudo cp kubetest2/* /usr/local/bin
	rm -rf kubetest2 kubetest2.tgz
