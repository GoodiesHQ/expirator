package report

// CredentialKind distinguishes the various types of credentials that can expire
type CredentialKind string

const (
	KindClientSecret CredentialKind = "ClientSecret"    // Azure AD app registration client secret
	KindAppCert      CredentialKind = "AppCertificate"  // Azure AD app registration certificate credential
	KindSAMLCert     CredentialKind = "SAMLSigningCert" // SAML signing certificate on an enterprise application
	KindSPSecret     CredentialKind = "SPSecret"        // Service principal password credential
	KindSPCert       CredentialKind = "SPCertificate"   // Service principal certificate credential
)

// SourceKind distinguishes app registrations from enterprise apps
type SourceKind string

const (
	SourceAppRegistration SourceKind = "AppRegistration" // Azure AD app registration
	SourceEnterpriseApp   SourceKind = "EnterpriseApp"   // Azure AD enterprise application (service principal)
)
