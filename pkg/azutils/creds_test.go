package azutils_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/quara-dev/beyond/pkg/azutils"
)

var _ = Describe("Creds", func() {
	Context("using client secret auth", func() {
		creds := new(azutils.CredentialConfig)
		creds.ClientId = "client-id"
		creds.ClientSecret = "client-secret"
		creds.TenantId = "tenant-id"
		It("return error if tenant id is not set", func() {
			creds := *creds
			creds.TenantId = ""
			_, err := creds.GetTenantId()
			Expect(err).To(HaveOccurred())
		})
		It("return an error if only client id is set", func() {
			creds := *creds
			creds.ClientSecret = ""
			err := creds.Build()
			Expect(err).To(HaveOccurred())
		})
		It("return an error if only client secret is set", func() {
			creds := *creds
			creds.ClientId = ""
			err := creds.Build()
			Expect(err).To(HaveOccurred())
		})
		It("return an error if both client_id and client_id_file are set", func() {
			creds := *creds
			creds.ClientIdFile = "client-id-file"
			err := creds.Build()
			Expect(err).To(HaveOccurred())
		})
		It("return an error if both client_secret and client_secret_file are set", func() {
			creds := *creds
			creds.ClientSecretFile = "client-secret-file"
			err := creds.Build()
			Expect(err).To(HaveOccurred())
		})
		It("return an error if both tenant_id and tenant_id_file are set", func() {
			creds := *creds
			creds.TenantIdFile = "tenant-id-file"
			err := creds.Build()
			Expect(err).To(HaveOccurred())
		})
		It("return client id", func() {
			Expect(creds.GetClientId()).To(Equal("client-id"))
		})
		It("return client secret", func() {
			Expect(creds.GetClientSecret()).To(Equal("client-secret"))
		})
		It("return tenant id", func() {
			Expect(creds.GetTenantId()).To(Equal("tenant-id"))
		})
		It("build credentials", func() {
			Expect(creds.Build()).To(Succeed())
		})
	})
	Context("using default identity", func() {
		creds := new(azutils.CredentialConfig)
		It("cannot build if no auth is set", func() {
			creds := *creds
			creds.NoDefaultCredentials = true
			creds.NoManagedIdentity = true
			err := creds.Build()
			Expect(err).To(HaveOccurred())
		})
		It("cannot return tenant id", func() {
			_, err := creds.GetTenantId()
			Expect(err).To(HaveOccurred())
		})
		It("cannot return client id", func() {
			_, err := creds.GetClientId()
			Expect(err).To(HaveOccurred())
		})
		It("cannot return client secret", func() {
			_, err := creds.GetClientSecret()
			Expect(err).To(HaveOccurred())
		})
		It("build credentials", func() {
			Expect(creds.Build()).To(Succeed())
		})
	})
	Context("using managed identity", func() {
		creds := new(azutils.CredentialConfig)
		creds.NoDefaultCredentials = true
		It("return an error if both resource id and resource id are set", func() {
			creds := *creds
			creds.ManagedIdentityClientId = "client-id"
			creds.ManagedIdentityResourceId = "resource-id"
			err := creds.Build()
			Expect(err).To(HaveOccurred())
		})
		It("return an error if both managed identity client id and client_id are set", func() {
			creds := *creds
			creds.ManagedIdentityClientId = "managed-client-id"
			creds.ClientId = "client-id"
			err := creds.Build()
			Expect(err).To(HaveOccurred())
		})
		It("return an error if both managed identity client id and client_id_file are set", func() {
			creds := *creds
			creds.ManagedIdentityClientId = "managed-client-id"
			creds.ClientIdFile = "client-id-file"
			err := creds.Build()
			Expect(err).To(HaveOccurred())
		})
		It("return an error if both managed identity resource id and client_id are set", func() {
			creds := *creds
			creds.ManagedIdentityResourceId = "managed-resource-id"
			creds.ClientId = "client-id"
			err := creds.Build()
			Expect(err).To(HaveOccurred())
		})
		It("return an error if both managed identity resource id and client_id_file are set", func() {
			creds := *creds
			creds.ManagedIdentityResourceId = "managed-resource-id"
			creds.ClientIdFile = "client-id-file"
			err := creds.Build()
			Expect(err).To(HaveOccurred())
		})
		It("build credentials", func() {
			Expect(creds.Build()).To(Succeed())
		})
	})
})
