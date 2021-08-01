package e2e

import (
	"bytes"
	"github.com/onsi/gomega"

	//"context"
	//"fmt"
	"k8s.io/klog"
	//"strconv"
	//"time"

	"github.com/onsi/ginkgo"
	//v1 "k8s.io/api/core/v1"

	//"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kubernetes/test/e2e/framework"

	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	e2ekubectl "k8s.io/kubernetes/test/e2e/framework/kubectl"
	e2elog "k8s.io/kubernetes/test/e2e/framework/log"
	e2epod "k8s.io/kubernetes/test/e2e/framework/pod"
)

const (
	podNetworkAnnotation = "k8s.ovn.org/pod-networks"
	agnhostImage         = "k8s.gcr.io/e2e-test-images/agnhost:2.26"
	certManagerFullYaml  = "../assets/cert-manager-full-generated-v1.4.1.yaml"
	certMangerNamespace  = "cert-manager"
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
		//str := framework.RunKubectlOrDie(certMangerNamespace, "get", "pods")
		//gomega.Expect(str).Should(gomega.MatchRegexp("cert-manager-"))
		//gomega.Expect(str).Should(gomega.MatchRegexp("cert-manager-cainjector-"))
		//gomega.Expect(str).Should(gomega.MatchRegexp("cert-manager-webhook-"))
	})

	//ginkgo.It("should provide k8s secret with generated certificate", func() {
	//	//klog.Infof("Namespace: %v", f.Namespace.Name)
	//	str := framework.RunKubectlOrDie(certMangerNamespace, "get", "pods")
	//	gomega.Expect(str).Should(gomega.MatchRegexp(".*cert-manager-cainjector2.*"))
	//})
})

//var _ = ginkgo.Describe("e2e nettest", func() {
//	var svcname = "nettest"
//
//	f := framework.NewDefaultFramework(svcname)
//
//	ginkgo.BeforeEach(func() {
//		// Assert basic external connectivity.
//		// Since this is not really a test of kubernetes in any way, we
//		// leave it as a pre-test assertion, rather than a Ginko test.
//		ginkgo.By("Executing a successful http request from the external internet")
//		resp, err := http.Get("http://google.com")
//		if err != nil {
//			framework.Failf("Unable to connect/talk to the internet: %v", err)
//		}
//		if resp.StatusCode != http.StatusOK {
//			framework.Failf("Unexpected error code, expected 200, got, %v (%v)", resp.StatusCode, resp)
//		}
//	})
//
//	ginkgo.It("should provide connection to external host by DNS name from a pod", func() {
//		ginkgo.By("Running container which tries to connect to www.google.com. in a loop")
//
//		podChan, errChan := make(chan *v1.Pod), make(chan error)
//		go func() {
//			defer ginkgo.GinkgoRecover()
//			checkContinuousConnectivity(f, "", "connectivity-test-continuous", "www.google.com.", 443, 30, podChan, errChan)
//		}()
//
//		testPod := <-podChan
//		framework.Logf("Test pod running on %q", testPod.Spec.NodeName)
//
//		time.Sleep(10 * time.Second)
//
//		framework.ExpectNoError(<-errChan)
//	})
//
//})

//func checkContinuousConnectivity(f *framework.Framework, nodeName, podName, host string, port, timeout int, podChan chan *v1.Pod, errChan chan error) {
//	contName := fmt.Sprintf("%s-container", podName)
//
//	command := []string{
//		"bash", "-c",
//		"set -xe; for i in {1..10}; do nc -vz -w " + strconv.Itoa(timeout) + " " + host + " " + strconv.Itoa(port) + "; sleep 2; done",
//	}
//
//	pod := &v1.Pod{
//		ObjectMeta: metav1.ObjectMeta{
//			Name: podName,
//		},
//		Spec: v1.PodSpec{
//			Containers: []v1.Container{
//				{
//					Name:    contName,
//					Image:   agnhostImage,
//					Command: command,
//				},
//			},
//			NodeName:      nodeName,
//			RestartPolicy: v1.RestartPolicyNever,
//		},
//	}
//	podClient := f.ClientSet.CoreV1().Pods(f.Namespace.Name)
//	_, err := podClient.Create(context.Background(), pod, metav1.CreateOptions{})
//	if err != nil {
//		errChan <- err
//		return
//	}
//
//	// Wait for pod network setup to be almost ready
//	wait.PollImmediate(1*time.Second, 30*time.Second, func() (bool, error) {
//		pod, err := podClient.Get(context.Background(), podName, metav1.GetOptions{})
//		if err != nil {
//			return false, nil
//		}
//		_, ok := pod.Annotations[podNetworkAnnotation]
//		return ok, nil
//	})
//
//	err = e2epod.WaitForPodNotPending(f.ClientSet, f.Namespace.Name, podName)
//	if err != nil {
//		errChan <- err
//		return
//	}
//
//	podGet, err := podClient.Get(context.Background(), podName, metav1.GetOptions{})
//	if err != nil {
//		errChan <- err
//		return
//	}
//
//	podChan <- podGet
//
//	err = e2epod.WaitForPodSuccessInNamespace(f.ClientSet, podName, f.Namespace.Name)
//
//	if err != nil {
//		logs, logErr := e2epod.GetPodLogs(f.ClientSet, f.Namespace.Name, pod.Name, contName)
//		if logErr != nil {
//			framework.Logf("Warning: Failed to get logs from pod %q: %v", pod.Name, logErr)
//		} else {
//			framework.Logf("pod %s/%s logs:\n%s", f.Namespace.Name, pod.Name, logs)
//		}
//	}
//
//	errChan <- err
//}
