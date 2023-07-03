package client_test

import (
	"fmt"
	"net/http"

	//
	"github.com/nikolalohinski/free-go/client"
	"github.com/nikolalohinski/free-go/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("virtual machines", func() {
	var (
		freeboxClient client.Client

		server   *ghttp.Server
		endpoint = new(string)

		sessionToken = new(string)

		returnedErr = new(error)
	)
	BeforeEach(func() {
		server = ghttp.NewServer()
		*endpoint = server.Addr()

		freeboxClient = Must(client.New(*endpoint, version)).(client.Client).
			WithAppID(appID).
			WithPrivateToken(privateToken)

		*sessionToken = setupLoginFlow(server)
	})
	AfterEach(func() {
		server.Close()
	})
	Context("getting virtual machine info", func() {
		returnedInfo := new(types.VirtualMachinesInfo)
		JustBeforeEach(func() {
			*returnedInfo, *returnedErr = freeboxClient.GetVirtualMachineInfo()
		})
		Context("default", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, fmt.Sprintf("/api/%s/vm/info/", version)),
						ghttp.RespondWith(http.StatusOK, `{
							"success": true,
							"result": {
								"usb_used": false,
								"sata_used": false,
								"sata_ports": [
									"sata-internal-p0",
									"sata-internal-p1",
									"sata-internal-p2",
									"sata-internal-p3"
								],
								"used_memory": 0,
								"usb_ports": [
									"usb-external-type-a",
									"usb-external-type-c"
								],
								"used_cpus": 0,
								"total_memory": 1024,
								"total_cpus": 2
							}
						}`),
					),
				)
			})
			It("should return the correct virtual machine info", func() {
				Expect(*returnedErr).To(BeNil())
				Expect((*returnedInfo)).To(Equal(types.VirtualMachinesInfo{
					USBUsed:  false,
					SATAUsed: false,
					SATAPorts: []string{
						"sata-internal-p0",
						"sata-internal-p1",
						"sata-internal-p2",
						"sata-internal-p3",
					},
					UsedMemory: 0,
					USBPorts: []string{
						"usb-external-type-a",
						"usb-external-type-c",
					},
					UsedCPUs:    0,
					TotalMemory: 1024,
					TotalCPUs:   2,
				},
				))
			})
		})
		Context("when server fails to respond", func() {
			BeforeEach(func() {
				server.Close()
			})
			It("should return an error", func() {
				Expect(*returnedErr).ToNot(BeNil())
			})
		})
		Context("when the server returns an unexpected payload", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, fmt.Sprintf("/api/%s/vm/info/", version)),
						ghttp.RespondWith(http.StatusOK, `{
							"success": true,
							"result": []
						}`),
					),
				)
			})
			It("should return an error", func() {
				Expect(*returnedErr).ToNot(BeNil())
			})
		})
	})
	Context("getting virtual machine distributions", func() {
		returnedDistros := new([]types.VirtualMachineDistribution)
		JustBeforeEach(func() {
			*returnedDistros, *returnedErr = freeboxClient.GetVirtualMachineDistributions()
		})
		Context("default", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, fmt.Sprintf("/api/%s/vm/distros/", version)),
						ghttp.RespondWith(http.StatusOK, `{
							"success": true,
							"result": [
								{
									"hash": "http://ftp.free.fr/.private/ubuntu-cloud/releases/jammy/release/SHA256SUMS",
									"os": "ubuntu",
									"url": "http://ftp.free.fr/.private/ubuntu-cloud/releases/jammy/release/ubuntu-22.04-server-cloudimg-arm64.img",
									"name": "Ubuntu 22.04 LTS (Jammy)"
								},
								{
									"hash": "http://ftp.free.fr/.private/ubuntu-cloud/releases/impish/release/SHA256SUMS",
									"os": "ubuntu",
									"url": "http://ftp.free.fr/.private/ubuntu-cloud/releases/impish/release/ubuntu-21.10-server-cloudimg-arm64.img",
									"name": "Ubuntu 21.10 (Impish)"
								}
							]
						}`),
					),
				)
			})
			It("should return the correct virtual machine info", func() {
				Expect(*returnedErr).To(BeNil())
				Expect((*returnedDistros)).To(Equal([]types.VirtualMachineDistribution{
					{
						Hash: "http://ftp.free.fr/.private/ubuntu-cloud/releases/jammy/release/SHA256SUMS",
						OS:   "ubuntu",
						URL:  "http://ftp.free.fr/.private/ubuntu-cloud/releases/jammy/release/ubuntu-22.04-server-cloudimg-arm64.img",
						Name: "Ubuntu 22.04 LTS (Jammy)",
					},
					{
						Hash: "http://ftp.free.fr/.private/ubuntu-cloud/releases/impish/release/SHA256SUMS",
						OS:   "ubuntu",
						URL:  "http://ftp.free.fr/.private/ubuntu-cloud/releases/impish/release/ubuntu-21.10-server-cloudimg-arm64.img",
						Name: "Ubuntu 21.10 (Impish)",
					},
				}))
			})
		})
		Context("when server fails to respond", func() {
			BeforeEach(func() {
				server.Close()
			})
			It("should return an error", func() {
				Expect(*returnedErr).ToNot(BeNil())
			})
		})
		Context("when the server returns an unexpected payload", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, fmt.Sprintf("/api/%s/vm/distros/", version)),
						ghttp.RespondWith(http.StatusOK, `{
							"success": true,
							"result": {}
						}`),
					),
				)
			})
			It("should return an error", func() {
				Expect(*returnedErr).ToNot(BeNil())
			})
		})
	})
	Context("listing virtual machines", func() {
		returnedMachines := new([]types.VirtualMachine)
		JustBeforeEach(func() {
			*returnedMachines, *returnedErr = freeboxClient.ListVirtualMachines()
		})
		Context("default", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, fmt.Sprintf("/api/%s/vm/", version)),
						ghttp.RespondWith(http.StatusOK, `{
							"success": true,
							"result": [
								{
									"mac": "f6:69:9c:d9:4f:3d",
									"cloudinit_userdata": "\n#cloud-config\n\n\nsystem_info:\n  default_user:\n    name: freemind\n",
									"cd_path": "L0ZyZWVib3gvcGF0aC90by9pbWFnZQ==",
									"id": 0,
									"os": "debian",
									"enable_cloudinit": true,
									"disk_path": "RnJlZWJveC9WTXMvZGViaWFuLnFjb3cy",
									"vcpus": 1,
									"memory": 300,
									"name": "testing",
									"cloudinit_hostname": "testing",
									"status": "stopped",
									"bind_usb_ports": "",
									"enable_screen": false,
									"disk_type": "qcow2"
								}
							]
						}`),
					),
				)
			})
			It("should return the correct virtual machine info", func() {
				Expect(*returnedErr).To(BeNil())
				Expect((*returnedMachines)).To(Equal([]types.VirtualMachine{{
					ID:                0,
					Name:              "testing",
					Mac:               "f6:69:9c:d9:4f:3d",
					DiskPath:          "RnJlZWJveC9WTXMvZGViaWFuLnFjb3cy",
					DiskType:          types.QCow2Disk,
					CDPath:            "/Freebox/path/to/image",
					Memory:            300,
					OS:                types.DebianOS,
					VCPUs:             1,
					Status:            types.StoppedStatus,
					EnableScreen:      false,
					BindUSBPorts:      []string{},
					EnableCloudInit:   true,
					CloudInitUserData: "\n#cloud-config\n\n\nsystem_info:\n  default_user:\n    name: freemind\n",
					CloudHostName:     "testing",
				}}))
			})
		})
		Context("when server fails to respond", func() {
			BeforeEach(func() {
				server.Close()
			})
			It("should return an error", func() {
				Expect(*returnedErr).ToNot(BeNil())
			})
		})
		Context("when the server returns an unexpected payload", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, fmt.Sprintf("/api/%s/vm/", version)),
						ghttp.RespondWith(http.StatusOK, `{
							"success": true,
							"result": {}
						}`),
					),
				)
			})
			It("should return an error", func() {
				Expect(*returnedErr).ToNot(BeNil())
			})
		})
	})
})
