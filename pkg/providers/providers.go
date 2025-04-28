package providers

type Provider interface {
	UpdatePTR(ip, nodeName string) error
}
