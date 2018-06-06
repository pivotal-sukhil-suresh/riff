package utils

import (
	"github.com/projectriff/riff/riff-cli/pkg/kubectl"
	"fmt"
	"errors"
	"strings"
	"github.com/projectriff/riff/riff-cli/pkg/jsonpath"
	"time"
	"github.com/projectriff/riff/riff-cli/pkg/minikube"
	"strconv"
)

type ComponentLookupHelper struct {
	kubectl kubectl.KubeCtl
	minik minikube.Minikube
}

type kubeContext struct {
	contextName string
	localCluster bool
}

type ComponentInfo struct {
	parser *jsonpath.Parser
}

func NewComponentInfo(context kubeContext, json string) ComponentInfo {
	return ComponentInfo{jsonpath.NewParser([]byte(json))}
}

func NewComponentLookupHelper(kubectl kubectl.KubeCtl, minik minikube.Minikube) ComponentLookupHelper {
	return ComponentLookupHelper{
		kubectl,
		minik,
	}
}

const LookupTimeoutDuration = 5 * time.Second

func (c ComponentLookupHelper) Lookup(component string) (address string, port string, err error) {

	cmdArgs := []string{"config", "current-context"}
	currentContext, err := c.kubectl.Exec(cmdArgs, LookupTimeoutDuration)
	if err != nil {
		return address, port, c.newError(component, err)
	}

	localCluster := false

	switch currentContext {
	case "minikube":
		localCluster = true
		address, err = c.minik.QueryIp()
		if err != nil {
			return address, port, c.newError(component, err)
		}

	case "docker-for-desktop":
		localCluster = true
		address = "127.0.0.1"
	}

	label := fmt.Sprintf("app=riff,component=%s", component)
	cmdArgs = []string{"get", "svc", "--all-namespaces", "-l", label, "-o", "json"}

	jsonString, err := c.kubectl.Exec(cmdArgs, LookupTimeoutDuration)
	if err != nil {
		return address, port, c.newError(component, err)
	}

	parser := jsonpath.NewParser([]byte(jsonString))
	portType, err := parser.StringValue(`$.items[0].spec.type`)
	if err != nil {
		return address, port, c.newError(component, errors.New("Is it deployed?"))
	}

	var portPath string

	switch portType {
	case "NodePort":
		if !localCluster {
			return address, port, c.newError(component, errors.New("NodePort supported only for local cluster."))
		}
		portPath = "$.items[0].spec.ports[?(@.name == http)].nodePort[0]"

	case "LoadBalancer":
		if localCluster {
			return address, port, c.newError(component, errors.New("LoadBalancer is supported only for remote cluster."))
		}

		portPath = "$.items[0].spec.ports[?(@.name == http)].port[0]"
		ip, _ := parser.StringValue(`$.items[0].status.loadBalancer.ingress[0].ip`)
		if ip == "" {
			hostname, _:= parser.StringValue(`$.items[0].status.loadBalancer.ingress[0].hostname`)
			if hostname == "" {
				return address, port, c.newError(component, errors.New("Unable to determine ip address or hostname."))
			} else {
				address = hostname
			}
		} else {
			address = ip
		}

	default:
		message := fmt.Sprintf("Unsupported port type `%s`", portType)
		return address, port, c.newError(component, errors.New(message))
	}

	value, err := parser.Value(portPath)
	if err != nil {
		return address, port, c.newError(component, errors.New("Missing port number."))
	}
	port = strconv.FormatFloat(value.(float64), 'f', 0, 64)

	return address, port, err
}

func (c ComponentLookupHelper) newError(component string, err error) error {
	message := fmt.Sprintf("Error looking up address for component `%s`.", component)
	if err != nil {
		message = strings.Join([]string {message, err.Error()}, " ")
	}
	return errors.New(message)
}


func (c ComponentLookupHelper) Lookup2(component string) (address string, port string, err error) {

	context, err := c.getKubeContext()
	if err != nil {
		return address, port, c.newError(component, err)
	}

	componentInfo, err := c.getComponentInfo(context, component)
	if err != nil {
		return address, port, c.newError(component, err)
	}

	portType, err := componentInfo.getPortType()
	if err != nil {
		return address, port, c.newError(component, errors.New("Is it deployed?"))
	}

	err = c.validatePortType(context, portType)
	if err != nil {
		return address, port, c.newError(component, err)
	}

	port, err = componentInfo.getPort(portType)

	if context.localCluster {
		address, err = c.getLocalAddress(context)
		if err != nil {
			return address, port, c.newError(component, err)
		}
	}

	address, err = componentInfo.getLoadBalancerAddress()
	if err != nil {
		return address, port, c.newError(component, err)
	}

	return
}

func (c ComponentLookupHelper) getKubeContext() (context kubeContext, err error) {

	var contextName string
	var localCluster bool

	cmdArgs := []string{"config", "current-context"}
	currentContext, err := c.kubectl.Exec(cmdArgs, LookupTimeoutDuration)
	if err != nil {
		return
	}

	localCluster = false

	switch currentContext {
	case "minikube":
		localCluster = true
		if err != nil {
			return
		}

	case "docker-for-desktop":
		localCluster = true
	}

	context = kubeContext{contextName, localCluster}
	return
}

func (c ComponentLookupHelper) validatePortType(context kubeContext, portType string) (err error) {

	if portType != "NodePort" && portType != "LoadBalancer" {
		return errors.New(fmt.Sprintf("Unsupported port type `%s`", portType))
	}

	if context.localCluster {
		if portType == "LoadBalancer" {
			return errors.New("LoadBalancer is supported only for remote cluster.")
		}
	} else { // remote cluster
		if portType == "NodePort" {
			return errors.New("NodePort supported only for local cluster.")
		}
	}

	return
}

func (c ComponentLookupHelper) getLocalAddress(context kubeContext) (address string, err error) {
	switch context.contextName {
	case "docker-for-desktop":
		address = "127.0.0.1"
	case "minkube":
		address, err = c.minik.QueryIp()
	}
	return
}

func (c ComponentLookupHelper) getComponentInfo(context kubeContext, component string) (componentInfo ComponentInfo, err error) {

	label := fmt.Sprintf("app=riff,component=%s", component)
	cmdArgs := []string{"get", "svc", "--all-namespaces", "-l", label, "-o", "json"}
	json, err := c.kubectl.Exec(cmdArgs, LookupTimeoutDuration)

	return NewComponentInfo(context, json), err
}

func (ci ComponentInfo) getPortType() (portType string, err error) {
	return ci.parser.StringValue(`$.items[0].spec.type`)
}

func (ci ComponentInfo) getPort(portType string) (port string, err error) {

	var portPath string

	switch portType {
	case "NodePort":
		portPath = "$.items[0].spec.ports[?(@.name == http)].nodePort[0]"

	case "LoadBalancer":
		portPath = "$.items[0].spec.ports[?(@.name == http)].port[0]"
	}

	value, err := ci.parser.Value(portPath)
	if err != nil {
		return port, errors.New("Missing port number.")
	}

	port = strconv.FormatFloat(value.(float64), 'f', 0, 64)

	return
}

func (ci ComponentInfo) getLoadBalancerAddress() (address string, err error) {
	ip, _ := ci.parser.StringValue(`$.items[0].status.loadBalancer.ingress[0].ip`)
	if ip == "" {
		hostname, _:= ci.parser.StringValue(`$.items[0].status.loadBalancer.ingress[0].hostname`)
		if hostname == "" {
			err = errors.New("Unable to determine ip address or hostname.")
		} else {
			address = hostname
		}
	} else {
		address = ip
	}
	return
}
