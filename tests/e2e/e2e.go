package e2e

import (
	"bytes"
	"context"
	"fmt"
	"github.com/k0kubun/pp"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/klog"
	"k8s.io/kubernetes/test/e2e/framework"
	e2ekubectl "k8s.io/kubernetes/test/e2e/framework/kubectl"
	e2elog "k8s.io/kubernetes/test/e2e/framework/log"
	e2epod "k8s.io/kubernetes/test/e2e/framework/pod"
	"time"
)

const (
	//podNetworkAnnotation = "k8s.ovn.org/pod-networks"
	//agnhostImage         = "k8s.gcr.io/e2e-test-images/agnhost:2.26"
	certManagerFullYaml = "../assets/cert-manager-full-generated-v1.4.1.yaml"
	certMangerNamespace = "cert-manager"
)

func applyManifest(yamlFile string) {
	var stdout, stderr bytes.Buffer
	var err error

	tk := e2ekubectl.NewTestKubeconfig(framework.TestContext.CertDir, framework.TestContext.Host, framework.TestContext.KubeConfig, framework.TestContext.KubeContext, framework.TestContext.KubectlPath, "")
	cmd := tk.KubectlCmd("apply", "-f", yamlFile)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		e2elog.Logf("Command finished with error: %v", err)
	}
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	klog.Infof("command stdout: %v", outStr)
	klog.Infof("command stderr: %v", errStr)

	framework.ExpectNoError(err)
}

var _ = ginkgo.Describe("e2e cert-manager", func() {
	var svcname = "certmanager"

	f := framework.NewDefaultFramework(svcname)

	ginkgo.BeforeEach(func() {
		//ensure if cert-manager and Issuer(s) is installed
		ginkgo.By("Executing cert-manager installation")
		applyManifest(certManagerFullYaml)

		ginkgo.By("Waiting to cert-manager's pods ready")
		err := e2epod.WaitForPodsRunningReady(f.ClientSet, certMangerNamespace, 3, 0, framework.PodStartShortTimeout, make(map[string]string))
		framework.ExpectNoError(err)

		ginkgo.By("Executing certs and Issuer objects")
		applyManifest("../assets/cert-manager-issuer-kind-test.yaml")
		applyManifest("../assets/cert-manager-issuer-kind-ca-test.yaml")
		applyManifest("../assets/cert-manager-certificate-test1.yaml")
		applyManifest("../assets/cert-manager-certificate-test2.yaml")

	})

	ginkgo.It("should cert-manager's pods running", func() {
		str := framework.RunKubectlOrDie(certMangerNamespace, "get", "pods")
		gomega.Expect(str).Should(gomega.MatchRegexp("cert-manager-"))
		gomega.Expect(str).Should(gomega.MatchRegexp("cert-manager-cainjector-"))
		gomega.Expect(str).Should(gomega.MatchRegexp("cert-manager-webhook-"))

	})

	ginkgo.It("should cert-manager Issuer exists", func() {
		ret, err := getCrdObjects(f.ClientSet)
		//klog.Infof("XXX crd list: %v", ret)
		pp.Print(ret)
		klog.Infof("XXX crd list err: %v", err)
	})

	//ginkgo.It("should provide k8s secret with generated certificate", func() {
	//	//klog.Infof("Namespace: %v", f.Namespace.Name)
	//	str := framework.RunKubectlOrDie(certMangerNamespace, "get", "pods")
	//	gomega.Expect(str).Should(gomega.MatchRegexp(".*cert-manager-cainjector2.*"))
	//})
})

func getCrdObjects(c clientset.Interface) (*apiextensionsv1.CustomResourceDefinitionList, error) {
	var client restclient.Result
	finished := make(chan struct{}, 1)
	go func() {
		// call chain tends to hang in some cases when Node is not ready. Add an artificial timeout for this call. #22165
		client = c.CoreV1().RESTClient().Get().
			AbsPath("/apis/cert-manager.io/v1/namespaces/cert-manager-local-ca/issuers").
			Do(context.TODO())

		finished <- struct{}{}
	}()
	select {
	case <-finished:
		result := &apiextensionsv1.CustomResourceDefinitionList{}
		if err := client.Into(result); err != nil {
			return &apiextensionsv1.CustomResourceDefinitionList{}, err
		}
		return result, nil
	case <-time.After(framework.PodStartShortTimeout):
		return &apiextensionsv1.CustomResourceDefinitionList{}, fmt.Errorf("Waiting up to %v for getting the list of CRDs", framework.PodStartShortTimeout)
	}
}
