
check_kubectl() {
  if [ -x /usr/bin/kubectl ]; then
    KUBECTL=kubectl
  else
    KUBECTL="ci-utils/kubectl"
  fi;
}

create_workload() {
  echo "creating workload..."
  sleep "0.$(( ( RANDOM % 100 )  + 1 ))"
  # deploy cert-manager
	scripts/cert-manager_install.sh
	# wait to cert-manager up and running
	$KUBECTL wait -n cert-manager --timeout=2m --for=condition=available deployment cert-manager
	$KUBECTL wait -n cert-manager --timeout=2m --for=condition=available deployment cert-manager-webhook
	$KUBECTL wait -n cert-manager --timeout=2m --for=condition=available deployment cert-manager-cainjector

	# create namespaces, certs and issuers
	$KUBECTL apply -f tests/assets/cert-manager-issuer-kind-test.yaml
	$KUBECTL apply -f tests/assets/cert-manager-issuer-kind-ca-test.yaml
	$KUBECTL apply -f tests/assets/cert-manager-certificate-test1.yaml
	$KUBECTL apply -f tests/assets/cert-manager-certificate-test2.yaml

  # install csi driver and example app
  $KUBECTL apply -f tests/assets/cert-manager-csi-driver.yaml
	$KUBECTL apply -f tests/assets/cert-manager-csi-driver-example-app.yaml
  # wait to app
  $KUBECTL wait -n cert-manager-local-ca2 --timeout=2m --for=condition=available deployment my-csi-app

}

delete_workload() {
  echo "deleting workload..."
  sleep "0.$(( ( RANDOM % 100 )  + 1 ))"
  #
  $KUBECTL delete -f tests/assets/cert-manager-csi-driver.yaml
	$KUBECTL delete -f tests/assets/cert-manager-csi-driver-example-app.yaml
#    sleep 2s
	$KUBECTL delete -f tests/assets/cert-manager-issuer-kind-test.yaml
	$KUBECTL delete -f tests/assets/cert-manager-issuer-kind-ca-test.yaml
	$KUBECTL delete -f tests/assets/cert-manager-certificate-test1.yaml
	$KUBECTL delete -f tests/assets/cert-manager-certificate-test2.yaml
}

check_cert_ca_in_pod_folder() {
  podname=$($KUBECTL -n cert-manager-local-ca2 get pods -l app=my-csi-app --no-headers -o name |head -1)
  $KUBECTL -n cert-manager-local-ca2 exec -t "$podname" -- stat /tls/ca.pem
}

check_cert_crt_in_pod_folder() {
  podname=$($KUBECTL -n cert-manager-local-ca2 get pods -l app=my-csi-app --no-headers -o name |head -1)
  $KUBECTL -n cert-manager-local-ca2 exec -t "$podname" -- stat /tls/crt.pem
}

check_cert_key_in_pod_folder() {
  podname=$($KUBECTL -n cert-manager-local-ca2 get pods -l app=my-csi-app --no-headers -o name |head -1)
  $KUBECTL -n cert-manager-local-ca2 exec -t "$podname" -- stat /tls/key.pem
}

Describe 'Test cert-manager and cert-manager csi-driver'
  setup() {
    check_kubectl
    create_workload
  }
  cleanup() {
#    :
    delete_workload
  }
  BeforeAll 'setup'
  AfterAll 'cleanup'

  It 'check cert ca in pod folder'
    When call check_cert_ca_in_pod_folder
    The stdout should include "Size:"
    The status should be success
  End

  It 'check cert crt in pod folder'
    When call check_cert_crt_in_pod_folder
    The stdout should include "Size:"
    The status should be success
  End

  It 'check cert key in pod folder'
    When call check_cert_key_in_pod_folder
    The stdout should include "Size:"
    The status should be success
  End
End
