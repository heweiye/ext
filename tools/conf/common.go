package conf

import (
	"encoding/json"
	"os"
	"strings"

	"v2ray.com/core/common/net"
	"v2ray.com/core/common/protocol"
)

type StringList []string

func NewStringList(raw []string) *StringList {
	list := StringList(raw)
	return &list
}

func (v StringList) Len() int {
	return len(v)
}

func (v *StringList) UnmarshalJSON(data []byte) error {
	var strarray []string
	if err := json.Unmarshal(data, &strarray); err == nil {
		*v = *NewStringList(strarray)
		return nil
	}

	var rawstr string
	if err := json.Unmarshal(data, &rawstr); err == nil {
		strlist := strings.Split(rawstr, ",")
		*v = *NewStringList(strlist)
		return nil
	}
	return newError("unknown format of a string list: " + string(data))
}

type Address struct {
	net.Address
}

func (v *Address) UnmarshalJSON(data []byte) error {
	var rawStr string
	if err := json.Unmarshal(data, &rawStr); err != nil {
		return newError("invalid address: ", string(data)).Base(err)
	}
	v.Address = net.ParseAddress(rawStr)

	return nil
}

func (v *Address) Build() *net.IPOrDomain {
	return net.NewIPOrDomain(v.Address)
}

type Network string

func (v Network) Build() net.Network {
	switch strings.ToLower(string(v)) {
	case "tcp":
		return net.Network_TCP
	case "udp":
		return net.Network_UDP
	default:
		return net.Network_Unknown
	}
}

type NetworkList []Network

func (v *NetworkList) UnmarshalJSON(data []byte) error {
	var strarray []Network
	if err := json.Unmarshal(data, &strarray); err == nil {
		nl := NetworkList(strarray)
		*v = nl
		return nil
	}

	var rawstr Network
	if err := json.Unmarshal(data, &rawstr); err == nil {
		strlist := strings.Split(string(rawstr), ",")
		nl := make([]Network, len(strlist))
		for idx, network := range strlist {
			nl[idx] = Network(network)
		}
		*v = nl
		return nil
	}
	return newError("unknown format of a string list: " + string(data))
}

func (v *NetworkList) Build() *net.NetworkList {
	if v == nil {
		return &net.NetworkList{
			Network: []net.Network{net.Network_TCP},
		}
	}

	list := new(net.NetworkList)
	for _, network := range *v {
		list.Network = append(list.Network, network.Build())
	}
	return list
}

func parseIntPort(data []byte) (net.Port, error) {
	var intPort uint32
	err := json.Unmarshal(data, &intPort)
	if err != nil {
		return net.Port(0), err
	}
	return net.PortFromInt(intPort)
}

func parseStringPort(data []byte) (net.Port, net.Port, error) {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return net.Port(0), net.Port(0), err
	}
	if strings.HasPrefix(s, "env:") {
		s = s[4:]
		s = os.Getenv(s)
	}

	pair := strings.SplitN(s, "-", 2)
	if len(pair) == 0 {
		return net.Port(0), net.Port(0), newError("Config: Invalid port range: ", s)
	}
	if len(pair) == 1 {
		port, err := net.PortFromString(pair[0])
		return port, port, err
	}

	fromPort, err := net.PortFromString(pair[0])
	if err != nil {
		return net.Port(0), net.Port(0), err
	}
	toPort, err := net.PortFromString(pair[1])
	if err != nil {
		return net.Port(0), net.Port(0), err
	}
	return fromPort, toPort, nil
}

type PortRange struct {
	From uint32
	To   uint32
}

func (v *PortRange) Build() *net.PortRange {
	return &net.PortRange{
		From: v.From,
		To:   v.To,
	}
}

// UnmarshalJSON implements encoding/json.Unmarshaler.UnmarshalJSON
func (v *PortRange) UnmarshalJSON(data []byte) error {
	port, err := parseIntPort(data)
	if err == nil {
		v.From = uint32(port)
		v.To = uint32(port)
		return nil
	}

	from, to, err := parseStringPort(data)
	if err == nil {
		v.From = uint32(from)
		v.To = uint32(to)
		if v.From > v.To {
			return newError("invalid port range ", v.From, " -> ", v.To)
		}
		return nil
	}

	return newError("invalid port range: ", string(data))
}

type User struct {
	EmailString string `json:"email"`
	LevelByte   byte   `json:"level"`
}

func (v *User) Build() *protocol.User {
	return &protocol.User{
		Email: v.EmailString,
		Level: uint32(v.LevelByte),
	}
}
