package client_test

import (
	"fmt"
	"net/http"

	"github.com/MakeNowJust/heredoc"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/nikolalohinski/free-go/client"
	"github.com/nikolalohinski/free-go/types"
)

var _ = Describe("login", func() {
	var (
		server   *ghttp.Server
		endpoint = new(string)

		freeboxClient client.Client

		permissions = new(types.Permissions)
		returnedErr = new(error)
	)
	BeforeEach(func() {
		server = ghttp.NewServer()
		*endpoint = server.Addr()

		freeboxClient = Must(client.New(*endpoint, version)).(client.Client).
			WithAppID(appID).
			WithPrivateToken(privateToken)
	})
	JustBeforeEach(func() {
		*permissions, *returnedErr = freeboxClient.Login()
	})
	AfterEach(func() {
		server.Close()
	})
	Context("default", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, fmt.Sprintf("/api/%s/login", version)),
					ghttp.RespondWith(http.StatusOK, heredoc.Doc(`{
						"success": true,
						"result": {
							"logged_in": false,
							"challenge": "9Va31tSgQWM853j0kSCtBUyzYNhPN7IY",
							"password_salt": "PJpG867vNjvbYY2z67Yy4164kEmmfrOC",
							"password_set": true
						}
					}`)),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodPost, fmt.Sprintf("/api/%s/login/session", version)),
					ghttp.VerifyContentType("application/json"),
					ghttp.VerifyJSON(`{
					    "app_id": "`+appID+`",
					    "password": "c3464d210c1be4f1ef6f34c578d463fc28d40a61"
					}`),
					ghttp.RespondWith(http.StatusOK, heredoc.Doc(`{
						"result": {
							"session_token": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
							"challenge": "9Va31tSgQWM853j0kSCtBUyzYNhPN7IY",
							"password_salt": "PJpG867vNjvbYY2z67Yy4164kEmmfrOC",
							"permissions": {
								"parental": false,
								"player": false,
								"explorer": false,
								"tv": false,
								"wdo": false,
								"downloader": false,
								"profile": false,
								"camera": false,
								"settings": true,
								"calls": false,
								"home": false,
								"pvr": false,
								"vm": true,
								"contacts": false
							},
							"password_set": true
						},
						"success": true
					}`)),
				),
			)
		})
		It("should return the correct permissions", func() {
			Expect(*returnedErr).To(BeNil())
			Expect(*permissions).To(Equal(types.Permissions{
				Settings: true,
				VM:       true,
			}))
		})
	})
	Context("when the server is unavailable", func() {
		Context("before the first call", func() {
			BeforeEach(func() {
				server.Close()
			})
			It("should return an error", func() {
				Expect(*returnedErr).ToNot(BeNil())
				Expect((*returnedErr).Error()).To(MatchRegexp(".* connect: connection refused"))
			})
		})
		Context("between both calls", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, fmt.Sprintf("/api/%s/login", version)),
						ghttp.RespondWith(http.StatusOK, heredoc.Doc(`{
							"success": true,
							"result": {
								"logged_in": false,
								"challenge": "9Va31tSgQWM853j0kSCtBUyzYNhPN7IY",
								"password_salt": "PJpG867vNjvbYY2z67Yy4164kEmmfrOC",
								"password_set": true
							}
						}`)),
					),
					func(w http.ResponseWriter, r *http.Request) {
						server.CloseClientConnections()
					},
				)
			})
			It("should return an error", func() {
				Expect(*returnedErr).ToNot(BeNil())
				Expect((*returnedErr).Error()).To(MatchRegexp(".* EOF"))
			})
		})
	})
	Context("when getting the challenge fails", func() {
		Context("with an unexpected status code", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, fmt.Sprintf("/api/%s/login", version)),
						ghttp.RespondWith(http.StatusBadGateway, "test body"),
					),
				)
			})
			It("should return an error", func() {
				Expect(*returnedErr).ToNot(BeNil())
				Expect((*returnedErr).Error()).To(MatchRegexp("failed with status '502': server returned 'test body'"))
			})
		})
		Context("with an error status code", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, fmt.Sprintf("/api/%s/login", version)),
						ghttp.RespondWith(http.StatusBadRequest, heredoc.Doc(`{
							"success": false,
							"error_code": "bad_request",
							"msg": "some error"
						}`)),
					),
				)
			})
			It("should return an error", func() {
				Expect(*returnedErr).ToNot(BeNil())
				Expect((*returnedErr).Error()).To(MatchRegexp("failed with error code 'bad_request': some error"))
			})
		})
		Context("because the returned body is an invalid JSON object", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, fmt.Sprintf("/api/%s/login", version)),
						ghttp.RespondWith(http.StatusBadRequest, "{"),
					),
				)
			})
			It("should return an error", func() {
				Expect(*returnedErr).ToNot(BeNil())
				Expect((*returnedErr).Error()).To(MatchRegexp("failed to unmarshal response body '{': .*"))
			})
		})
	})
	Context("when getting the session token fails", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, fmt.Sprintf("/api/%s/login", version)),
					ghttp.RespondWith(http.StatusOK, heredoc.Doc(`{
						"success": true,
						"result": {
							"logged_in": false,
							"challenge": "9Va31tSgQWM853j0kSCtBUyzYNhPN7IY",
							"password_salt": "PJpG867vNjvbYY2z67Yy4164kEmmfrOC",
							"password_set": true
						}
					}`)),
				),
			)
		})
		Context("with an unexpected status code", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPost, fmt.Sprintf("/api/%s/login/session", version)),
						ghttp.VerifyContentType("application/json"),
						ghttp.VerifyJSON(`{
						    "app_id": "`+appID+`",
						    "password": "c3464d210c1be4f1ef6f34c578d463fc28d40a61"
						}`),
						ghttp.RespondWith(http.StatusBadGateway, "test body"),
					),
				)
			})
			It("should return an error", func() {
				Expect(*returnedErr).ToNot(BeNil())
				Expect((*returnedErr).Error()).To(MatchRegexp("failed with status '502': server returned 'test body'"))
			})
		})
		Context("with an error status code", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPost, fmt.Sprintf("/api/%s/login/session", version)),
						ghttp.VerifyContentType("application/json"),
						ghttp.VerifyJSON(`{
						    "app_id": "`+appID+`",
						    "password": "c3464d210c1be4f1ef6f34c578d463fc28d40a61"
						}`),
						ghttp.RespondWith(http.StatusForbidden, heredoc.Doc(`{
						    "uid": "9bb8f32441fcb41e4c9f2d9b60af3b13",
						    "success": false,
						    "msg": "Erreur d'authentification de l'application",
						    "result": {
								"challenge": "9Va31tSgQWM853j0kSCtBUyzYNhPN7IY",
								"password_salt": "PJpG867vNjvbYY2z67Yy4164kEmmfrOC"
						    },
						    "error_code": "invalid_token"
						}`)),
					),
				)
			})
			It("should return an error", func() {
				Expect(*returnedErr).ToNot(BeNil())
				Expect((*returnedErr).Error()).To(MatchRegexp("failed with error code 'invalid_token': Erreur d'authentification de l'application"))
			})
		})
		Context("because the returned body is an invalid JSON object", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPost, fmt.Sprintf("/api/%s/login/session", version)),
						ghttp.VerifyContentType("application/json"),
						ghttp.VerifyJSON(`{
						    "app_id": "`+appID+`",
						    "password": "c3464d210c1be4f1ef6f34c578d463fc28d40a61"
						}`),
						ghttp.RespondWith(http.StatusForbidden, "{"),
					),
				)
			})
			It("should return an error", func() {
				Expect(*returnedErr).ToNot(BeNil())
				Expect((*returnedErr).Error()).To(MatchRegexp("failed to unmarshal response body '{': .*"))
			})
		})
	})
	Context("when the server returned an unexpected payload for the login challenge", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, fmt.Sprintf("/api/%s/login", version)),
					ghttp.RespondWith(http.StatusOK, heredoc.Doc(`{
						"success": true,
						"result": []
					}`)),
				),
			)
		})
		It("should return an error", func() {
			Expect(*returnedErr).ToNot(BeNil())
		})
	})
	Context("when the server returned an unexpected payload for the session result", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodGet, fmt.Sprintf("/api/%s/login", version)),
					ghttp.RespondWith(http.StatusOK, heredoc.Doc(`{
						"success": true,
						"result": {
							"logged_in": false,
							"challenge": "9Va31tSgQWM853j0kSCtBUyzYNhPN7IY",
							"password_salt": "PJpG867vNjvbYY2z67Yy4164kEmmfrOC",
							"password_set": true
						}
					}`)),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodPost, fmt.Sprintf("/api/%s/login/session", version)),
					ghttp.VerifyContentType("application/json"),
					ghttp.VerifyJSON(`{
					    "app_id": "`+appID+`",
					    "password": "c3464d210c1be4f1ef6f34c578d463fc28d40a61"
					}`),
					ghttp.RespondWith(http.StatusOK, heredoc.Doc(`{
						"success": true,
						"result": []
					}`)),
				),
			)
		})
		It("should return an error", func() {
			Expect(*returnedErr).ToNot(BeNil())
		})
	})
	Context("when app id is not set", func() {
		BeforeEach(func() {
			freeboxClient = Must(client.New(*endpoint, version)).(client.Client).
				WithPrivateToken(privateToken)
		})

		It("should return the correct error", func() {
			Expect(*returnedErr).ToNot(BeNil())
			Expect(*returnedErr).To(Equal(client.ErrAppIDIsNotSet))
		})
	})
	Context("when private token is not set", func() {
		BeforeEach(func() {
			freeboxClient = Must(client.New(*endpoint, version)).(client.Client).
				WithAppID(appID)
		})

		It("should return the correct error", func() {
			Expect(*returnedErr).ToNot(BeNil())
			Expect(*returnedErr).To(Equal(client.ErrPrivateTokenIsNotSet))
		})
	})
})
