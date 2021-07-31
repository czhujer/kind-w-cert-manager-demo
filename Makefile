.PHONY: kubetest-up
kubetest-up:
	kubetest2 kind --config kind-config2-without-cni.yaml --up --stderrthreshold 5

.PHONY: kubetest-cilium-install
kubetest-cilium-install:
	kubetest2 kind --test exec -- sh -c scripts/cilium_install.sh

.PHONY: kubetest-down
kubetest2-down:
	kubetest2 kind --down

.PHONY: kind-integration-deploy-cert-manager
kind-integration-deploy-cert-manager:
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

.PHONY: kind-integration-run-shellspec
kind-integration-run-shellspec:
	bash -c ' \
	SHELLSPEC_TEST_SUITES+="tests/shellspec/run_pytest_and_rspec_spec.sh" \
	shellspec --default-path tests/shellspec --load-path tests/shellspec -o j -f d --reportdir ./ $$SHELLSPEC_TEST_SUITES \
	'
