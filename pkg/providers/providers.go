package providers

type HostingProvider interface {
	UpdatePTR(ip, nodeName string) error
}

type DnsProvider interface {
	UpdateA(ip, nodeName string) error
}
