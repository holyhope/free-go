//go:build integration

package integration_test

import (
	"fmt"

	"github.com/nikolalohinski/free-go/client"
	"github.com/nikolalohinski/free-go/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("port forwarding scenarios", func() {

	BeforeEach(func() {
		freeboxClient = freeboxClient.WithAppID(appID).WithPrivateToken(token)

		permissions := Must(freeboxClient.Login()).(types.Permissions)
		if !permissions.Settings {
			panic(fmt.Sprintf("the token for the '%s' app does not appear to have the permissions to modify freebox settings", appID))
		}
	})

	Context("full lifecycle of a port forwarding rule", func() {
		It("should not return an error nor unexpected responses", func() {
			// create
			enabled := true
			payload := types.PortForwardingRulePayload{
				Enabled:      &enabled,
				IPProtocol:   types.TCP,
				WanPortStart: 12345,
				WanPortEnd:   12345,
				LanIP:        "192.168.1.128",
				SourceIP:     "0.0.0.0",
				LanPort:      8080,
				Comment:      "free-go integration tests",
			}
			createdRule, err := freeboxClient.CreatePortForwardingRule(payload)
			Expect(err).To(BeNil())
			Expect(createdRule).To(MatchFields(IgnoreExtras, Fields{
				"Valid":                     BeTrue(),
				"ID":                        Not(BeZero()),
				"PortForwardingRulePayload": Equal(payload),
			}))

			// read
			readRule, err := freeboxClient.GetPortForwardingRule(createdRule.ID)
			Expect(err).To(BeNil())
			Expect(readRule).To(Equal(createdRule))

			// update
			updatedRule, err := freeboxClient.UpdatePortForwardingRule(readRule.ID, types.PortForwardingRulePayload{
				Enabled: new(bool),
			})
			Expect(err).To(BeNil())
			Expect(updatedRule).To(MatchFields(IgnoreExtras, Fields{
				"PortForwardingRulePayload": MatchFields(IgnoreExtras, Fields{
					"Enabled": PointTo(BeFalse()),
				}),
			}))

			// list
			rules, err := freeboxClient.ListPortForwardingRules()
			Expect(err).To(BeNil())
			Expect(rules).ToNot(BeEmpty())
			Expect(rules).To(ContainElement(Equal(updatedRule)))

			// delete
			err = freeboxClient.DeletePortForwardingRule(updatedRule.ID)
			Expect(err).To(BeNil())

			// Check rule was deleted
			_, err = freeboxClient.GetPortForwardingRule(updatedRule.ID)
			Expect(err).To(MatchError(client.ErrPortForwardingRuleNotFound))
		})
	})
})
