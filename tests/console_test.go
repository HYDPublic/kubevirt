/*
 * This file is part of the KubeVirt project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2017 Red Hat, Inc.
 *
 */

package tests_test

import (
	"flag"
	"time"

	"github.com/google/goexpect"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"kubevirt.io/kubevirt/pkg/api/v1"
	"kubevirt.io/kubevirt/pkg/kubecli"
	"kubevirt.io/kubevirt/tests"
)

var _ = Describe("Console", func() {

	flag.Parse()

	virtClient, err := kubecli.GetKubevirtClient()
	tests.PanicOnError(err)

	BeforeEach(func() {
		tests.BeforeTestCleanup()
	})

	RunVMAndExpectConsoleOutput := func(vm *v1.VirtualMachine, expected string) {

		By("Creating a new VM")
		Expect(virtClient.RestClient().Post().Resource("virtualmachines").Namespace(tests.NamespaceTestDefault).Body(vm).Do().Error()).To(Succeed())
		tests.WaitForSuccessfulVMStart(vm)

		By("Expecting the VM console")
		expecter, _, err := tests.NewConsoleExpecter(virtClient, vm, 10*time.Second)
		defer expecter.Close()
		Expect(err).ToNot(HaveOccurred())

		By("Checking that the console output equals to expected one")
		_, err = expecter.ExpectBatch([]expect.Batcher{
			&expect.BExp{R: expected},
		}, 120*time.Second)
		Expect(err).ToNot(HaveOccurred())
	}

	Describe("A new VM", func() {
		Context("with a serial console", func() {
			Context("with a cirros image", func() {
				It("should return that we are running cirros", func() {
					vm := tests.NewRandomVMWithEphemeralDiskAndUserdata(tests.RegistryDiskFor(tests.RegistryDiskCirros), "#!/bin/bash\necho 'hello'\n")
					RunVMAndExpectConsoleOutput(
						vm,
						"login as 'cirros' user",
					)
				}, 140)
			})

			Context("with a fedora image", func() {
				It("should return that we are running fedora", func() {
					vm := tests.NewRandomVMWithEphemeralDiskHighMemory(tests.RegistryDiskFor(tests.RegistryDiskFedora))
					RunVMAndExpectConsoleOutput(
						vm,
						"Welcome to",
					)
				}, 140)
			})

			It("should be able to reconnect to console multiple times", func() {
				vm := tests.NewRandomVMWithEphemeralDisk(tests.RegistryDiskFor(tests.RegistryDiskAlpine))

				By("Creating a new VM")
				Expect(virtClient.RestClient().Post().Resource("virtualmachines").Namespace(tests.NamespaceTestDefault).Body(vm).Do().Error()).To(Succeed())
				tests.WaitForSuccessfulVMStart(vm)

				for i := 0; i < 5; i++ {
					By("Expecting a VM console")
					expecter, _, err := tests.NewConsoleExpecter(virtClient, vm, 10*time.Second)
					defer expecter.Close()
					Expect(err).ToNot(HaveOccurred())

					By("Checking that the console output equals to expected one")
					_, err = expecter.ExpectBatch([]expect.Batcher{
						&expect.BSnd{S: "\n"},
						&expect.BExp{R: "login"},
					}, 160*time.Second)
					Expect(err).ToNot(HaveOccurred())
				}
			}, 220)
		})
	})
})
