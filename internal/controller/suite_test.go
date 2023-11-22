/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	aiv1a1 "github.com/weave-ai/lm-controller/api/v1alpha1"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	cfg        *rest.Config
	k8sClient  client.Client
	testEnv    *envtest.Environment
	testScheme *runtime.Scheme
)

func TestMain(m *testing.M) {
	g := NewGomega(func(message string, callerSkip ...int) {
		logf.Log.WithName("Gomega").Info(message)
	})
	logf.SetLogger(zap.New(zap.WriteTo(os.Stderr), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	{
		testEnv = &envtest.Environment{
			CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
			ErrorIfCRDPathMissing: true,
		}

		var err error
		// cfg is defined in this file globally.
		cfg, err = testEnv.Start()
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(cfg).NotTo(BeNil())

		testScheme = runtime.NewScheme()

		err = sourcev1.AddToScheme(testScheme)
		g.Expect(err).NotTo(HaveOccurred())

		err = aiv1a1.AddToScheme(testScheme)
		g.Expect(err).NotTo(HaveOccurred())

		//+kubebuilder:scaffold:scheme

		k8sClient, err = client.New(cfg, client.Options{Scheme: testScheme})
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(k8sClient).NotTo(BeNil())
	}

	By("running the test suite")
	code := m.Run()

	By("tearing down the test environment")
	{
		err := testEnv.Stop()
		g.Expect(err).NotTo(HaveOccurred())
	}
	os.Exit(code)
}

func By(s string) {
	logf.Log.WithName("By").Info(s)
}
