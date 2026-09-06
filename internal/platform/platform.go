package platform

type Platform interface {
	SetupDNS(tld string, dnsPort int) error
	TeardownDNS(tld string) error
	SetupPortForward(fromPort, toPort int) error
	TeardownPortForward(fromPort, toPort int) error
	InstallDaemon(binaryPath, sockPath string) error
	UninstallDaemon() error
	NeedsSudo() bool
}
