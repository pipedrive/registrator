package bridge

import (
	"strconv"
	"strings"

	"github.com/cenkalti/backoff"
	dockerapi "github.com/fsouza/go-dockerclient"
	"bytes"
	"text/template"
	log "github.com/pipedrive/registrator/logger"
	"syscall"
)

func retry(fn func() error) error {
	return backoff.Retry(fn, backoff.NewExponentialBackOff())
}

func mapDefault(m map[string]string, key, default_ string) string {
	v, ok := m[key]
	if !ok || v == "" {
		return default_
	}
	return v
}

// Golang regexp module does not support /(?!\\),/ syntax for spliting by not escaped comma
// Then this function is reproducing it
func recParseEscapedComma(str string) []string {
	if len(str) == 0 {
		return []string{}
	} else if str[0] == ',' {
		return recParseEscapedComma(str[1:])
	}

	offset := 0
	for len(str[offset:]) > 0 {
		index := strings.Index(str[offset:], ",")

		if index == -1 {
			break
		} else if str[offset+index-1:offset+index] != "\\" {
			return append(recParseEscapedComma(str[offset+index+1:]), str[:offset+index])
		}

		str = str[:offset+index-1] + str[offset+index:]
		offset += index
	}

	return []string{str}
}

func combineTags(tagParts ...string) []string {
	tags := make([]string, 0)
	for _, element := range tagParts {
		tags = append(tags, recParseEscapedComma(element)...)
	}
	return tags
}

func serviceMetaData(config *dockerapi.Config, port string) (map[string]string, map[string]bool) {
	meta := config.Env
	for k, v := range config.Labels {
		meta = append(meta, k+"="+v)
	}
	metadata := make(map[string]string)
	metadataFromPort := make(map[string]bool)
	for _, kv := range meta {
		kvp := strings.SplitN(kv, "=", 2)
		if strings.HasPrefix(kvp[0], "SERVICE_") && len(kvp) > 1 {
			key := strings.ToLower(strings.TrimPrefix(kvp[0], "SERVICE_"))
			if metadataFromPort[key] {
				continue
			}
			portkey := strings.SplitN(key, "_", 2)
			_, err := strconv.Atoi(portkey[0])
			if err == nil && len(portkey) > 1 {
				if portkey[0] != port {
					continue
				}
				metadata[portkey[1]] = kvp[1]
				metadataFromPort[portkey[1]] = true
			} else {
				metadata[key] = kvp[1]
			}
		}
	}
	return metadata, metadataFromPort
}

func servicePort(container *dockerapi.Container, port dockerapi.Port, published []dockerapi.PortBinding) ServicePort {
	log.Debugf("Building servicePort %s", container.ID[:12])
	var hp, hip, eip, nm string
	if len(published) > 0 {
		hp = published[0].HostPort
		hip = published[0].HostIP
		log.Debugf("Found Published Port for %s - \"%s:%s\"", container.ID[:12], hip, hp)
	}
	if hip == "" {
		hip = "0.0.0.0"
	}

	//for overlay networks
	//detect if container use overlay network, than set HostIP into NetworkSettings.Network[string].IPAddress
	//better to use registrator with -internal flag
	nm = container.HostConfig.NetworkMode
	log.Debugf("Network mode for %s is: \"%s\"", container.ID[:12], nm)
	if nm != "bridge" && nm != "default" && nm != "host" {
		hip = container.NetworkSettings.Networks[nm].IPAddress
	}

	// Nir: support docker NetworkSettings
	eip = container.NetworkSettings.IPAddress
	if eip == "" {
		for network_name, network := range container.NetworkSettings.Networks {
			if network_name != "ingress" {
				eip = network.IPAddress
				log.Debugf("Container %s exposed IP is: \"%s\"", container.ID[:12], eip)
			}
		}
	}

	return ServicePort{
		HostPort:          hp,
		HostIP:            hip,
		ExposedPort:       port.Port(),
		ExposedIP:         eip,
		PortType:          port.Proto(),
		ContainerID:       container.ID,
		ContainerHostname: container.Config.Hostname,
		container:         container,
	}
}

func (f *ContainersFilters) String() string {
	var buffer bytes.Buffer
	isFirst := true

	for key, val := range *f {
		if !isFirst {
			buffer.WriteString("&")
		}
		buffer.WriteString("(")
		buffer.WriteString(key)
		buffer.WriteString("=")
		buffer.WriteString(strings.Join(val, "|"))
		buffer.WriteString(")")
		isFirst = false
	}

	return buffer.String()
}

func (f *ContainersFilters) Set(value string) error {
	split := strings.SplitN(value, "=", 2)

	var key, val string

	if splitSize := len(split); splitSize < 2 {
		key = value
		val = ""
	} else {
		key, val = split[0], split[1]
	}

	_, exists := (*f)[key]
	if !exists {
		(*f)[key] = []string{val}
	} else {
		(*f)[key] = append((*f)[key], val)
	}

	return nil
}

func (f *ContainersFilters) WithContainerId(containerId string) ContainersFilters {
	filtersCopy := make(ContainersFilters)

	for key, val := range *f {
		valCopy := make([]string, len(val))
		copy(valCopy, val)
		filtersCopy[key] = valCopy
	}

	filtersCopy["id"] = []string{containerId}

	return filtersCopy
}

func EvaluateTemplateTags(s *string, container *dockerapi.Container) string {
	if container == nil || s == nil || *s == "" {
		return *s
	}

	tmpl := template.New("template-tags")
	tmpl, err := tmpl.Parse(*s)

	if err != nil {
		log.Printf("template tags: Unable to parse %s template", s)
		return *s
	}

	tmplVal := bytes.NewBufferString("")

	if err = tmpl.Execute(tmplVal, container); err != nil {
		log.Printf("template tags: Unable to evaluate %s template against container", s)
		return *s
	}

	return tmplVal.String()
}

func SignalFromEvent(msg *dockerapi.APIEvents) syscall.Signal {
	signal := syscall.Signal(-1)

	if _, ok := msg.Actor.Attributes["signal"]; ok {
		i, err := strconv.Atoi(msg.Actor.Attributes["signal"])
		if err == nil {
			signal = syscall.Signal(i)
		}
	}

	return signal
}