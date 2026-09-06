//go:build !darwin

package platform

import "fmt"

type Stub struct{}

func Current() Platform {
	return &Stub{}
}

func (s *Stub) NeedsSudo() bool                        { return false }
func (s *Stub) SetupDNS(tld string, dnsPort int) error { return fmt.Errorf("platform not supported") }
func (s *Stub) TeardownDNS(tld string) error           { return fmt.Errorf("platform not supported") }
func (s *Stub) SetupPortForward(from, to int) error    { return fmt.Errorf("platform not supported") }
func (s *Stub) TeardownPortForward(from, to int) error { return fmt.Errorf("platform not supported") }
func (s *Stub) InstallDaemon(bin, sock string) error   { return fmt.Errorf("platform not supported") }
func (s *Stub) UninstallDaemon() error                 { return fmt.Errorf("platform not supported") }
