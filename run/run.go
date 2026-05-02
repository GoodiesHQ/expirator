package run

import (
	"context"
	"fmt"
	"slices"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/goodieshq/expirator/config"
	"github.com/goodieshq/expirator/report"
	"github.com/goodieshq/expirator/utils"
	graphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	graphcore "github.com/microsoftgraph/msgraph-sdk-go-core"
	"github.com/microsoftgraph/msgraph-sdk-go/applications"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/serviceprincipals"
)

const pageSize = int32(50)

// Expirator is the main entry point for running the expirator logic based on the provided config and returns a report/error.
func Expirator(ctx context.Context, cfg *config.Config) (*report.Report, error) {
	// Build Azure credential from a tenant ID and a client ID/secret
	cred, err := azidentity.NewClientSecretCredential(
		cfg.Azure.TenantID, cfg.Azure.ClientID, cfg.Azure.ClientSecret, nil,
	)
	if err != nil {
		return nil, fmt.Errorf("building Azure credential: %w", err)
	}

	// Build Graph client with the credential
	graphClient, err := graphsdk.NewGraphServiceClientWithCredentials(
		cred,
		[]string{"https://graph.microsoft.com/.default"},
	)
	if err != nil {
		return nil, fmt.Errorf("building Graph client: %w", err)
	}

	// Helper function to determine if an entry should be included in the report based on the config
	shouldReport := utils.WithExpiringWithin(cfg.ExpiringWithinDays, cfg.IncludeExpired)

	// Slice of all entries to include in the report
	var entries []report.Entry

	// Add application registrations
	apps, err := runAzAppRegistrations(ctx, graphClient)
	if err != nil {
		return nil, fmt.Errorf("checking app registrations: %w", err)
	}
	entries = slices.Concat(entries, report.FromApplications(apps))

	// Add service principals (enterprise applications)
	sps, err := runAzServicePrincipals(ctx, graphClient)
	if err != nil {
		return nil, fmt.Errorf("checking service principals: %w", err)
	}
	entries = slices.Concat(entries, report.FromServicePrincipals(sps))

	// At this point, the list of entries are already grouped by source

	// Delete any entries that are outside the configured reporting window
	entries = slices.DeleteFunc(
		entries,
		func(e report.Entry) bool {
			return !shouldReport(e.DateExpiry)
		},
	)

	// If not grouping by source, sort the entries by expiry date
	if !cfg.GroupSources {
		slices.SortFunc(entries, func(e1, e2 report.Entry) int {
			if e1.DateExpiry.After(e2.DateExpiry) {
				return 1
			} else if e1.DateExpiry.Before(e2.DateExpiry) {
				return -1
			}
			return 0
		})
	}

	// Return the report
	return &report.Report{
		Entries: entries,
	}, nil
}

// runAzAppRegistrations fetches all app registrations in the tenant and returns them as a slice of models.Applicationable.
func runAzAppRegistrations(ctx context.Context, gc *graphsdk.GraphServiceClient) ([]models.Applicationable, error) {
	// Build the initial request configuration with select query parameters and page size
	cfg := &applications.ApplicationsRequestBuilderGetRequestConfiguration{
		QueryParameters: &applications.ApplicationsRequestBuilderGetQueryParameters{
			Select: []string{"id", "appId", "displayName", "passwordCredentials", "keyCredentials"},
			Top:    utils.Ptr(pageSize),
		},
	}

	// Fetch the first page of results
	result, err := gc.Applications().Get(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("fetching applications: %w", err)
	}

	// Use a pager to iterate through all pages of results and collect the applications into a slice
	pager, err := graphcore.NewPageIterator[models.Applicationable](
		result, gc.GetAdapter(), models.CreateApplicationCollectionResponseFromDiscriminatorValue,
	)
	if err != nil {
		return nil, fmt.Errorf("creating applications pager: %w", err)
	}

	// Iterate through the pages and append applications to the output slice
	var out []models.Applicationable
	err = pager.Iterate(ctx, func(a models.Applicationable) bool {
		out = append(out, a)
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("iterating applications: %w", err)
	}

	return out, nil
}

// runAzServicePrincipals fetches all service principals (enterprise applications) in the tenant and returns them as a slice of models.ServicePrincipalable.
func runAzServicePrincipals(ctx context.Context, gc *graphsdk.GraphServiceClient) ([]models.ServicePrincipalable, error) {
	// Build the initial request configuration with select query parameters and page size
	cfg := &serviceprincipals.ServicePrincipalsRequestBuilderGetRequestConfiguration{
		QueryParameters: &serviceprincipals.ServicePrincipalsRequestBuilderGetQueryParameters{
			Select: []string{
				"id", "appId", "displayName", "servicePrincipalType",
				"preferredSingleSignOnMode", "passwordCredentials", "keyCredentials",
			},
			Top: utils.Ptr(pageSize),
		},
	}

	// Fetch the first page of results
	result, err := gc.ServicePrincipals().Get(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("fetching service principals: %w", err)
	}

	// Use a pager to iterate through all pages of results and collect the service principals into a slice
	pager, err := graphcore.NewPageIterator[models.ServicePrincipalable](
		result, gc.GetAdapter(), models.CreateServicePrincipalCollectionResponseFromDiscriminatorValue,
	)
	if err != nil {
		return nil, fmt.Errorf("creating service principals pager: %w", err)
	}

	// Iterate through the pages and append service principals to the output slice
	var out []models.ServicePrincipalable
	err = pager.Iterate(ctx, func(sp models.ServicePrincipalable) bool {
		out = append(out, sp)
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("iterating service principals: %w", err)
	}

	return out, nil
}
