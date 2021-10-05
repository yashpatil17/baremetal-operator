package ironic

import (
	"fmt"
	"testing"

	"github.com/gophercloud/gophercloud/openstack/baremetal/v1/nodes"
	"github.com/stretchr/testify/assert"

	"github.com/yashpatil17/baremetal-operator/pkg/bmc"
	"github.com/yashpatil17/baremetal-operator/pkg/provisioner/ironic/clients"
	"github.com/yashpatil17/baremetal-operator/pkg/provisioner/ironic/testserver"
)

func TestHasCapacity(t *testing.T) {

	states := []nodes.ProvisionState{
		nodes.Cleaning, nodes.CleanWait, nodes.Inspecting, nodes.InspectWait, nodes.Deploying, nodes.DeployWait,
		nodes.Deleting,
	}

	cases := []struct {
		name              string
		provisioningLimit int
		nodeStates        []nodes.ProvisionState
		hostName          string

		expectedHasCapacity bool
		expectedError       string
	}{
		{
			name:              "no-capacity",
			provisioningLimit: len(states),
			nodeStates:        states,

			expectedHasCapacity: false,
		},
		{
			name:              "enough-capacity",
			provisioningLimit: len(states) + 1,
			nodeStates:        states,

			expectedHasCapacity: true,
		},
		{
			name:              "ignore-check-if-already-provisioning",
			provisioningLimit: len(states),
			nodeStates:        states,
			hostName:          "node-1",

			expectedHasCapacity: true,
		},
		{
			name:              "enough-capacity-due-not-provisioning-states",
			provisioningLimit: 1,
			nodeStates:        []nodes.ProvisionState{nodes.Active, nodes.AdoptFail, nodes.Adopting, nodes.Available, nodes.CleanFail},

			expectedHasCapacity: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			allNodes := []nodes.Node{}
			for n, state := range tc.nodeStates {
				allNodes = append(allNodes, nodes.Node{
					Name:           fmt.Sprintf("myns%snode-%d", nameSeparator, n),
					ProvisionState: string(state),
				})
			}

			ironic := testserver.NewIronic(t).Nodes(allNodes).Start()
			defer ironic.Stop()

			inspector := testserver.NewInspector(t).Start()
			defer inspector.Stop()

			host := makeHost()
			host.Name = tc.hostName

			auth := clients.AuthConfig{Type: clients.NoAuth}

			prov, err := newProvisionerWithSettings(host, bmc.Credentials{}, nullEventPublisher,
				ironic.Endpoint(), auth, inspector.Endpoint(), auth,
			)
			if err != nil {
				t.Fatalf("could not create provisioner: %s", err)
			}
			prov.config.maxBusyHosts = tc.provisioningLimit

			result, err := prov.HasCapacity()

			assert.Equal(t, tc.expectedHasCapacity, result)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Regexp(t, tc.expectedError, err.Error())
			}
		})
	}
}
