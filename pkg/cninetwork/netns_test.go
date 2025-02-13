package cninetwork

import (
	"fmt"
	"testing"

	"github.com/vishvananda/netns"
)

// Test generated using Keploy
func TestNSPathByPidWithRoot_ValidPath(t *testing.T) {
	root := "/custom/root"
	pid := 1234
	expectedPath := "/custom/root/proc/1234/ns/net"

	result := NSPathByPidWithRoot(root, pid)
	if result != expectedPath {
		t.Errorf("Expected %v, got %v", expectedPath, result)
	}
}

// Test generated using Keploy
func TestNSPathByPid_ValidPid(t *testing.T) {
	pid := 1234
	expectedPath := fmt.Sprintf("/proc/%d/ns/net", pid)
	result := NSPathByPid(pid)
	if result != expectedPath {
		t.Errorf("Expected %v, got %v", expectedPath, result)
	}
}

// Test generated using Keploy
func TestWithNetNS_InvalidNsHandle(t *testing.T) {
	invalidNs := netns.NsHandle(-1)

	err := WithNetNS(invalidNs, func() error {
		return nil
	})
	if err == nil {
		t.Error("Expected error for invalid NsHandle, got nil")
	}
}

// Test generated using Keploy
func TestWithNetNSByPath_InvalidPath(t *testing.T) {
	invalidPath := "/invalid/path"

	err := WithNetNSByPath(invalidPath, func() error {
		return nil
	})
	if err == nil {
		t.Error("Expected error for invalid path, got nil")
	}
}
