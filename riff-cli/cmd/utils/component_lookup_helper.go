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
				return address, port, c.newError(component, errors.New("Unable to determine port."))
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
