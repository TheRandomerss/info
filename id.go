package info

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type DeviceInfo struct {
	MacAddress  string
	Ram         string
	Cpu         string
	User        string
	Path        string
	FingerPrint string
}

func IsVirtualInterface(name string) bool {
	virtualInterfaceKeywords := []string{"virtual", "pseudo", "loopback", "tunnel", "software"}
	for _, keyword := range virtualInterfaceKeywords {
		if strings.Contains(strings.ToLower(name), keyword) {
			return true
		}
	}
	return false
}

func GetPhysicalMacAddr() (string, error) {
	ifas, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, ifa := range ifas {
		if !IsVirtualInterface(ifa.Name) {
			return ifa.HardwareAddr.String(), nil
		}
	}

	return "", fmt.Errorf("No physical MAC address found")
}

func GetSystemInfo() DeviceInfo {
	var info DeviceInfo

	info.MacAddress, _ = GetPhysicalMacAddr()

	// Choose the appropriate commands for RAM and CPU based on the operating system
	var ramCmd, cpuCmd *exec.Cmd

	if runtime.GOOS == "windows" {
		ramCmd = exec.Command("wmic", "ComputerSystem", "get", "TotalPhysicalMemory")
		cpuCmd = exec.Command("wmic", "cpu", "get", "name")

		ramOutput, _ := ramCmd.Output()
		info.Ram = strings.Fields(strings.TrimSpace(string(ramOutput)))[1]

		// Get CPU information
		_cpuOutput, _ := cpuCmd.Output()
		cpuOutput := strings.Fields(strings.TrimSpace(string(_cpuOutput)))
		for i := 1; i < len(cpuOutput); i++ {
			info.Cpu += cpuOutput[i]
		}

		// Get current user
		info.User = os.Getenv("USERNAME")

		// Get home directory path
		info.Path = os.Getenv("USERPROFILE")
		hash := HashDeviceInfo(info)
		info.FingerPrint = hash
		return info
	} else {
		ramCmd = exec.Command("grep", "MemTotal", "/proc/meminfo")
		cpuCmd = exec.Command("lscpu")

		ramOutput, _ := ramCmd.Output()
		info.Ram = strings.Split(strings.ReplaceAll(strings.TrimSpace(string(ramOutput)), " ", ""), ":")[1]

		// Get CPU information
		cpuOutput, _ := cpuCmd.Output()
		info.Cpu = strings.TrimSpace(string(cpuOutput))

		// Get current user
		info.User = os.Getenv("USER")

		// Get home directory path
		info.Path = os.Getenv("HOME")
		hash := HashDeviceInfo(info)
		info.FingerPrint = hash
		return info
	}

}

func HashDeviceInfo(data DeviceInfo) string {
	// Concatenate the strings
	concatenatedString := data.MacAddress + data.Ram + data.Cpu + data.User + data.Path
	// Hash the concatenated string using SHA-256
	hasher := sha256.New()
	hasher.Write([]byte(concatenatedString))
	hash := hex.EncodeToString(hasher.Sum(nil))
	return hash
}
