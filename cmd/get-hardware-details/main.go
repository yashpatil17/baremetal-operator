// get-hardware-details is a tool that can be used to convert raw Ironic introspection data into the HardwareDetails
// type used by Metal3.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/gophercloud/gophercloud/openstack/baremetalintrospection/v1/introspection"

	"github.com/yashpatil17/baremetal-operator/pkg/provisioner/ironic/clients"
	"github.com/yashpatil17/baremetal-operator/pkg/provisioner/ironic/hardwaredetails"
)

type options struct {
	Endpoint   string
	AuthConfig clients.AuthConfig
	NodeID     string
}

func main() {
	opts := getOptions()
	ironicTrustedCAFile := os.Getenv("IRONIC_CACERT_FILE")
	ironicInsecureStr := os.Getenv("IRONIC_INSECURE")
	ironicInsecure := false
	if strings.ToLower(ironicInsecureStr) == "true" {
		ironicInsecure = true
	}

	tlsConf := clients.TLSConfig{
		TrustedCAFile:      ironicTrustedCAFile,
		InsecureSkipVerify: ironicInsecure,
	}

	inspector, err := clients.InspectorClient(opts.Endpoint, opts.AuthConfig, tlsConf)
	if err != nil {
		fmt.Printf("could not get inspector client: %s", err)
		os.Exit(1)
	}

	introData := introspection.GetIntrospectionData(inspector, opts.NodeID)
	data, err := introData.Extract()
	if err != nil {
		fmt.Printf("could not get introspection data: %s", err)
		os.Exit(1)
	}

	json, err := json.MarshalIndent(hardwaredetails.GetHardwareDetails(data), "", "\t")
	if err != nil {
		fmt.Printf("could not convert introspection data: %s", err)
		os.Exit(1)
	}

	fmt.Println(string(json))
}

func getOptions() (o options) {
	if len(os.Args) != 3 {
		fmt.Println("Usage: get-hardware-details <inspector URI> <node UUID>")
		os.Exit(1)
	}

	var err error
	o.Endpoint, o.AuthConfig, err = clients.ConfigFromEndpointURL(os.Args[1])
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	o.NodeID = os.Args[2]
	return
}
