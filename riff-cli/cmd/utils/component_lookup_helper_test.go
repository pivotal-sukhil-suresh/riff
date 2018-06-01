package utils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/riff-cli/pkg/kubectl"
	"errors"
	"github.com/projectriff/riff/riff-cli/cmd/utils"
	"fmt"
	"github.com/projectriff/riff/riff-cli/pkg/minikube"
	"io/ioutil"
	"strings"
	"github.com/projectriff/riff/riff-cli/pkg/jsonpath"
)

var _ = Describe("Looking up address and port of a riff component", func() {

	var (
		kubeClient            *kubectl.MockKubeCtl
		minik          *minikube.MockMinikube
		lookupHelper          utils.ComponentLookupHelper

		currentContextCmdArgs []string
		currentContext string
		currentContextError error

		componentName string
	)

	BeforeEach(func() {
		kubeClient = new(kubectl.MockKubeCtl)
		minik = new(minikube.MockMinikube)
		lookupHelper = utils.NewComponentLookupHelper(kubeClient, minik)

		componentName = "http-gateway"
		currentContextCmdArgs = []string{"config", "current-context"}
	})

	JustBeforeEach(func() {
		kubeClient.
			On("Exec", currentContextCmdArgs, utils.LookupTimeoutDuration).
			Return(currentContext, currentContextError)
	})

	Context("when kubectl current-context is NOT configured", func() {

		BeforeEach(func() {
			currentContextError = errors.New("error: current-context is not set")
		})

		It("should error when fetching current context", func() {
			_, _, err := lookupHelper.Lookup(componentName)
			Expect(err).To(HaveOccurred())

			expectedErrorMessage := fmt.Sprintf("Error looking up address for component `%s`. %s", componentName, currentContextError.Error())
			Expect(err.Error()).To(Equal(expectedErrorMessage))
		})
	})

	Context("when kubectl current-context is configured", func() {

		var (
			lookupCmdArgs []string
			lookupResponse string
			lookupError error
		)

		BeforeEach(func() {
			currentContextError = nil
			lookupCmdArgs = []string{"get", "svc", "--all-namespaces", "-l", fmt.Sprintf("app=riff,component=%s", componentName), "-o", "json"}
		})

		JustBeforeEach(func() {
			kubeClient.
				On("Exec", lookupCmdArgs, utils.LookupTimeoutDuration).
				Return(lookupResponse, lookupError)
		})

		AssertUnsupportedPortTypeBehaviour := func() {

			Context("when portType for component is NOT supported", func() {

				var portType string

				BeforeEach(func() {
					bytes, err := ioutil.ReadFile("../../test_data/command/utils/UnsupportedPortType.json")
					Expect(err).NotTo(HaveOccurred())
					lookupResponse = string(bytes)

					parser := jsonpath.NewParser(bytes)
					portType, err = parser.StringValue(`$.items[0].spec.type`)
					Expect(err).ToNot(HaveOccurred())
				})

				It("should error with unsupported portType", func() {
					_, _, err := lookupHelper.Lookup(componentName)
					Expect(err).To(HaveOccurred())
					expectedMessage := fmt.Sprintf("Error looking up address for component `%s`. Unsupported port type `%s`", componentName, portType)
					Expect(err.Error()).To(Equal(expectedMessage))
				})
			})
		}

		AssertComponentNotDeployedBehaviour := func() {

			Context("when component is NOT deployed", func() {

				BeforeEach(func() {
					bytes, err := ioutil.ReadFile("../../test_data/command/utils/ComponentNotDeployed.json")
					Expect(err).NotTo(HaveOccurred())
					lookupResponse = string(bytes)
				})

				It("should error saying component may not be deployed", func() {
					_, _, err := lookupHelper.Lookup(componentName)
					Expect(err).To(HaveOccurred())
					expectedMessage := fmt.Sprintf("Error looking up address for component `%s`. Is it deployed?", componentName)
					Expect(err.Error()).To(Equal(expectedMessage))
				})
			})
		}

		Context("when cluster is remote (current-context is NOT minikube or docker-for-desktop)", func() {

			BeforeEach(func() {
				currentContext = "remote\n"
			})

			AssertComponentNotDeployedBehaviour()

			Context("when component is deployed", func() {

				AssertUnsupportedPortTypeBehaviour()

				Context("when portType for component is NodePort", func() {

					BeforeEach(func() {
						bytes, err := ioutil.ReadFile("../../test_data/command/utils/NodePortType.json")
						Expect(err).NotTo(HaveOccurred())
						lookupResponse = strings.Replace(string(bytes), "<port>", "3206", 1)
					})

					It("should error because NodePort is NOT supported for remote cluster", func() {
						_, _, err := lookupHelper.Lookup(componentName)
						Expect(err).To(HaveOccurred())
						expectedMessage := fmt.Sprintf("Error looking up address for component `%s`. NodePort supported only for local cluster.", componentName)
						Expect(err.Error()).To(Equal(expectedMessage))
					})
				})

				Context("when portType for component is LoadBalancer", func() {

					var componentAddress string
					var componentPort string

					BeforeEach(func() {
						componentPort = "3206"
					})

					Context("when hostname and IP is NOT available for LoadBalancer", func() {

						BeforeEach(func() {
							bytes, err := ioutil.ReadFile("../../test_data/command/utils/LoadBalancerPortType_Invalid_MissingIpHostname.json")
							Expect(err).NotTo(HaveOccurred())
							lookupResponse = string(bytes)
						})

						It("errors with failure to determine address", func() {
							_, _, err := lookupHelper.Lookup(componentName)
							Expect(err).To(HaveOccurred())
							expectedMessage := fmt.Sprintf("Error looking up address for component `%s`. Unable to determine ip address or hostname.", componentName)
							Expect(err.Error()).To(Equal(expectedMessage))
						})
					})

					Context("when IP is used for LoadBalancer", func() {

						BeforeEach(func() {
							componentAddress = "172.217.0.227"

							bytes, err := ioutil.ReadFile("../../test_data/command/utils/LoadBalancerIpPortType.json")
							Expect(err).NotTo(HaveOccurred())

							lookupResponse = strings.Replace(string(bytes), "<ip-address>", componentAddress, 1)
							lookupResponse = strings.Replace(lookupResponse, "<port>", componentPort, 1)
						})

						It("returns the IP and port", func() {
							address, port, err := lookupHelper.Lookup(componentName)
							Expect(err).ToNot(HaveOccurred())
							Expect(address).To(Equal(componentAddress))
							Expect(port).To(Equal(componentPort))
						})
					})

					Context("when hostname is used for LoadBalancer", func() {

						BeforeEach(func() {
							componentAddress = "riff.http-gateway"

							bytes, err := ioutil.ReadFile("../../test_data/command/utils/LoadBalancerHostnamePortType.json")
							Expect(err).NotTo(HaveOccurred())

							lookupResponse = strings.Replace(string(bytes), "<port>", componentPort, 1)
							lookupResponse = strings.Replace(lookupResponse, "<hostname>", componentAddress, 1)
						})

						It("returns the hostname and port", func() {
							address, port, err := lookupHelper.Lookup(componentName)
							Expect(err).ToNot(HaveOccurred())
							Expect(address).To(Equal(componentAddress))
							Expect(port).To(Equal(componentPort))
						})
					})
				})
			})
		})

		Context("when cluster is NOT remote", func() {

			var componentAddress string

			AssertNodePortBehaviour := func() {

				Context("when portType for component is NodePort", func() {

					Context("when NodePort number is NOT available", func() {

						BeforeEach(func() {
							bytes, err := ioutil.ReadFile("../../test_data/command/utils/NodePortType_Invalid_NodePortAddressMissing.json")
							Expect(err).NotTo(HaveOccurred())
							lookupResponse = string(bytes)
						})

						It("should error about missing port number", func() {
							_, _, err := lookupHelper.Lookup(componentName)
							Expect(err).To(HaveOccurred())
							expectedMessage := fmt.Sprintf("Error looking up address for component `%s`. Missing port number.", componentName)
							Expect(err.Error()).To(Equal(expectedMessage))
						})
					})

					Context("when NodePort number is available", func() {

						var componentPort string

						BeforeEach(func() {
							bytes, err := ioutil.ReadFile("../../test_data/command/utils/NodePortType.json")
							Expect(err).NotTo(HaveOccurred())

							componentPort = "3206"
							lookupResponse = strings.Replace(string(bytes), "<port>", componentPort, 1)
						})

						It("should return address and port number", func() {
							address, port, err := lookupHelper.Lookup(componentName)
							Expect(err).ToNot(HaveOccurred())
							Expect(address).To(Equal(componentAddress))
							Expect(port).To(Equal(componentPort))
						})
					})
				})
			}

			AssertIncompatiblePortTypeBehaviour := func() {

				Context("when portType for component is LoadBalancer", func() {

					BeforeEach(func() {
						bytes, err := ioutil.ReadFile("../../test_data/command/utils/LoadBalancerIpPortType.json")
						Expect(err).NotTo(HaveOccurred())

						lookupResponse = strings.Replace(string(bytes), "<port>", "32607", 1)
						lookupResponse = strings.Replace(lookupResponse, "<ip-address>", "172.217.0.227", 1)
					})

					It("should error because LoadBalancer is NOT supported for local cluster", func() {
						_, _, err := lookupHelper.Lookup(componentName)
						Expect(err).To(HaveOccurred())
						expectedMessage := fmt.Sprintf("Error looking up address for component `%s`. LoadBalancer is supported only for remote cluster.", componentName)
						Expect(err.Error()).To(Equal(expectedMessage))
					})
				})
			}

			Context("when current-context is minikube", func() {

				BeforeEach(func() {
					currentContext = "minikube\n"
				})

				XContext("when minikube is NOT running", func() {

					var minikubeError error

					BeforeEach(func() {
						minikubeError = errors.New("Minikube is not running")
						//minik.On("Status").Return(...)
					})

					It("should fail fast", func() {
						_, _, err := lookupHelper.Lookup(componentName)
						Expect(err).To(HaveOccurred())
						expectedErrorMessage := fmt.Sprintf("Error looking up address for component `%s`. %s", componentName, minikubeError.Error())
						Expect(err.Error()).To(Equal(expectedErrorMessage))
					})
				})

				Context("when minikube is running", func() {

					var queryIpError error

					JustBeforeEach(func() {
						minik.On("QueryIp").Return(componentAddress, queryIpError)
					})

					Context("when minikube QueryIp does NOT work", func() {

						BeforeEach(func() {
							queryIpError = errors.New("querying for minikube ip failed")
						})

						It("should error", func() {
							_, _, err := lookupHelper.Lookup(componentName)
							Expect(err).To(HaveOccurred())
							expectedErrorMessage := fmt.Sprintf("Error looking up address for component `%s`. %s", componentName, queryIpError.Error())
							Expect(err.Error()).To(Equal(expectedErrorMessage))
						})
					})

					Context("when minikube QueryIp works", func() {

						BeforeEach(func() {
							queryIpError = nil
							componentAddress = "192.168.99.100"
						})

						AssertComponentNotDeployedBehaviour()
						AssertNodePortBehaviour()
						AssertUnsupportedPortTypeBehaviour()
						AssertIncompatiblePortTypeBehaviour()
					})
				})
			})

			Context("when current-context is docker-for-desktop", func() {

				BeforeEach(func() {
					currentContext = "docker-for-desktop\n"
					componentAddress = "127.0.0.1"
				})

				AssertComponentNotDeployedBehaviour()
				AssertNodePortBehaviour()
				AssertUnsupportedPortTypeBehaviour()
				AssertIncompatiblePortTypeBehaviour()
			})
		})
	})
})
