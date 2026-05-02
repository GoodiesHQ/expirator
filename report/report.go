package report

import (
	"slices"
	"time"

	"github.com/goodieshq/expirator/utils"
	"github.com/google/uuid"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

// Report contains all entries
type Report struct {
	Entries []Entry `json:"entries"`
}

// EntryMin is a minimal version of Entry used for output formats that don't need all details
type EntryMin struct {
	OwnerName  string         `json:"owner_name"`
	SourceKind SourceKind     `json:"source_kind"`
	CredKind   CredentialKind `json:"cred_kind"`
	CredName   string         `json:"cred_name"`
	DateExpiry time.Time      `json:"date_expiry"`
}

// Entry represents a single expiring credential, either from an app registration or service principal
type Entry struct {
	OwnerName  string         `json:"owner_name"`
	OwnerAppID string         `json:"owner_app_id"`
	OwnerObjID string         `json:"owner_obj_id"`
	SourceKind SourceKind     `json:"source_kind"`
	CredKind   CredentialKind `json:"cred_kind"`
	CredKeyID  uuid.UUID      `json:"cred_key_id"`
	CredName   string         `json:"cred_name"`
	DateStart  time.Time      `json:"date_start"`
	DateExpiry time.Time      `json:"date_expiry"`
}

// Min converts an Entry to its minimal form, EntryMin
func (e Entry) Min() EntryMin {
	return EntryMin{
		OwnerName:  e.OwnerName,
		SourceKind: e.SourceKind,
		CredKind:   e.CredKind,
		CredName:   e.CredName,
		DateExpiry: e.DateExpiry,
	}
}

// FromApplications converts a list of Graph Application models to a list of report Entries
func FromApplications(apps []models.Applicationable) []Entry {
	entries := make([]Entry, 0)

	// Iterate over applications and extract credentials
	for _, app := range apps {
		owner := utils.Deref(app.GetDisplayName())
		appID := utils.Deref(app.GetAppId())
		objID := utils.Deref(app.GetId())

		// Iterate over password credentials (secrets)
		for _, cred := range app.GetPasswordCredentials() {
			entries = append(entries, Entry{
				OwnerName:  owner,
				OwnerAppID: appID,
				OwnerObjID: objID,
				SourceKind: SourceAppRegistration,
				CredKind:   KindClientSecret,
				CredName:   utils.Deref(cred.GetDisplayName()),
				CredKeyID:  utils.Deref(cred.GetKeyId()),
				DateStart:  utils.Deref(cred.GetStartDateTime()),
				DateExpiry: utils.Deref(cred.GetEndDateTime()),
			})
		}

		// Iterate over key credentials (certificates)
		for _, k := range app.GetKeyCredentials() {
			entries = append(entries, Entry{
				OwnerName:  owner,
				OwnerAppID: appID,
				OwnerObjID: objID,
				SourceKind: SourceAppRegistration,
				CredKind:   KindAppCert,
				CredName:   utils.Deref(k.GetDisplayName()),
				CredKeyID:  utils.Deref(k.GetKeyId()),
				DateStart:  utils.Deref(k.GetStartDateTime()),
				DateExpiry: utils.Deref(k.GetEndDateTime()),
			})
		}
	}

	// Sort entries by expiry date (soonest first)
	slices.SortFunc(entries, func(e1, e2 Entry) int {
		if e1.DateExpiry.After(e2.DateExpiry) {
			return 1
		} else if e1.DateExpiry.Before(e2.DateExpiry) {
			return -1
		}
		return 0
	})
	return entries
}

// FromServicePrincipals converts a list of Graph ServicePrincipal models to a list of report Entries
func FromServicePrincipals(sps []models.ServicePrincipalable) []Entry {
	entries := make([]Entry, 0)

	// Iterate over service principals and extract credentials
	for _, sp := range sps {
		owner := utils.Deref(sp.GetDisplayName())
		appID := utils.Deref(sp.GetAppId())
		objID := utils.Deref(sp.GetId())

		// Iterate over password credentials (secrets)
		for _, cred := range sp.GetPasswordCredentials() {
			entries = append(entries, Entry{
				OwnerName:  owner,
				OwnerAppID: appID,
				OwnerObjID: objID,
				SourceKind: SourceEnterpriseApp,
				CredKind:   KindSPSecret,
				CredName:   utils.Deref(cred.GetDisplayName()),
				CredKeyID:  utils.Deref(cred.GetKeyId()),
				DateStart:  utils.Deref(cred.GetStartDateTime()),
				DateExpiry: utils.Deref(cred.GetEndDateTime()),
			})
		}

		// Iterate over key credentials (certificates)
		for _, k := range sp.GetKeyCredentials() {
			// Determine if this is a regular SP cert or a SAML cert based on usage and type
			kind := KindSPCert
			if utils.Deref(k.GetUsage()) == "Verify" && utils.Deref(k.GetTypeEscaped()) == "AsymmetricX509Cert" {
				kind = KindSAMLCert
			}
			entries = append(entries, Entry{
				OwnerName:  owner,
				OwnerAppID: appID,
				OwnerObjID: objID,
				SourceKind: SourceEnterpriseApp,
				CredKind:   kind,
				CredName:   utils.Deref(k.GetDisplayName()),
				CredKeyID:  utils.Deref(k.GetKeyId()),
				DateStart:  utils.Deref(k.GetStartDateTime()),
				DateExpiry: utils.Deref(k.GetEndDateTime()),
			})
		}
	}

	// Sort entries by expiry date (soonest first)
	slices.SortFunc(entries, func(e1, e2 Entry) int {
		if e1.DateExpiry.After(e2.DateExpiry) {
			return 1
		} else if e1.DateExpiry.Before(e2.DateExpiry) {
			return -1
		}
		return 0
	})
	return entries
}
