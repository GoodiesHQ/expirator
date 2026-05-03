package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/goodieshq/expirator/config"
	"github.com/goodieshq/expirator/run"
	"github.com/goodieshq/expirator/utils"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v3"
)

var Version string = "dev"

func init() {
	godotenv.Load()
}

func main() {
	app := &cli.Command{
		Name:    "expirator",
		Version: utils.GetVersion(),
		Usage:   "Report client secret and SAML certificate expirations across an Azure tenant",
		Description: "Authenticates with client credentials against Microsoft Graph and reports\n" +
			"expiration metadata for app registration secrets/certs and enterprise app\n" +
			"SAML signing certs. Read-only — never reads secret values, only metadata.\n\n" +
			"Required env vars: AZ_TENANT_ID, AZ_CLIENT_ID, AZ_CLIENT_SECRET\n" +
			"Required Graph permission: Application.Read.All (admin consent required)",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "expiring-within-days",
				Aliases: []string{"e"},
				Usage:   "Only show credentials expiring within N days (0 = show all)",
				Value:   0,
				Sources: cli.EnvVars("EXPIRING_WITHIN_DAYS"),
			},
			&cli.BoolFlag{
				Name:    "include-expired",
				Aliases: []string{"x"},
				Usage:   "Include already expired credentials in the report (by default, only unexpired credentials are shown, but marked as EXPIRED if past their expiry date)",
				Value:   false,
				Sources: cli.EnvVars("INCLUDE_EXPIRED"),
			},
			&cli.BoolFlag{
				Name:    "group-sources",
				Aliases: []string{"g"},
				Usage:   "Group credentials by their source application. Otherwise, all credentials are sorted by expiry date",
				Value:   false,
				Sources: cli.EnvVars("GROUP_SOURCES"),
			},
			&cli.StringFlag{
				Name:    "output-file",
				Aliases: []string{"o"},
				Usage:   "Output file (defaults to stdout)",
				Value:   "",
				Sources: cli.EnvVars("OUTPUT_FILE"),
			},
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"f"},
				Usage:   "Output format: table, csv, json",
				Value:   "table",
				Sources: cli.EnvVars("OUTPUT_FORMAT"),
				Validator: func(s string) error {
					switch s {
					case config.FORMAT_TABLE, config.FORMAT_CSV, config.FORMAT_JSON:
						return nil
					}
					return fmt.Errorf("unknown output format (must be %q, %q, or %q)", config.FORMAT_TABLE, config.FORMAT_CSV, config.FORMAT_JSON)
				},
			},
			&cli.StringFlag{
				Name:    "webhook-url",
				Aliases: []string{"u"},
				Usage:   "If set, a POST request will be sent with the report as JSON to this URL. ",
				Value:   "",
				Sources: cli.EnvVars("WEBHOOK_URL"),
			},
			&cli.StringFlag{
				Name:     "az-tenant-id",
				Usage:    "Azure tenant ID",
				Sources:  cli.EnvVars("AZ_TENANT_ID"),
				Required: true,
			},
			&cli.StringFlag{
				Name:     "az-client-id",
				Usage:    "Azure application client ID",
				Sources:  cli.EnvVars("AZ_CLIENT_ID"),
				Required: true,
			},
			&cli.StringFlag{
				Name:     "az-client-secret",
				Usage:    "Azure client secret (env var suggested)",
				Sources:  cli.EnvVars("AZ_CLIENT_SECRET"),
				Required: true,
			},
		},
		Action: do,
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func do(ctx context.Context, c *cli.Command) error {
	cfg, err := config.FromCmd(c)
	if err != nil {
		return err
	}

	// Open output file (or default to stdout)
	var f *os.File = os.Stdout
	if cfg.OutputFile != "" && cfg.OutputFile != "-" && cfg.Format != config.FORMAT_NONE {
		f, err = os.Create(cfg.OutputFile)
		if err != nil {
			return fmt.Errorf("creating output file: %w", err)
		}
		defer f.Close()
	}

	// Run the expirator program
	rep, err := run.Expirator(ctx, cfg)
	if err != nil {
		return err
	}

	// Send a report to the webhook URL if configured
	if cfg.WebhookURL != "" {
		if err := utils.SendWebhook(ctx, cfg.WebhookURL, rep.Entries); err != nil {
			// Don't fail the whole program if the webhook fails; just log the error and continue
			log.Printf("sending webhook: %v", err)
		}
	}

	// Output results in the requested format
	switch cfg.Format {
	case config.FORMAT_NONE:
		// No output
	case config.FORMAT_TABLE:
		// Table format uses a subset of the fields for brevity

		// Use go-pretty to render a nice table in the terminal
		tw := table.NewWriter()
		tw.SetOutputMirror(f)

		// Header row
		tw.AppendHeader(table.Row{
			"Cred Source", "Application", "Cred Name", "Cred Type", "Expiry",
		})

		// Data rows
		for _, entry := range rep.Entries {
			tw.AppendRow(table.Row{
				entry.SourceKind, entry.OwnerName, entry.CredName, entry.CredKind, utils.FormatPrettyExpiry(entry.DateExpiry),
			})
		}

		// Render the table to the output
		tw.Render()
	case config.FORMAT_CSV:
		// CSV format includes all fields for maximum fidelity and machine-readability
		columns := []string{
			"owner_name", "owner_app_id", "owner_object_id", "source_kind",
			"cred_kind", "cred_key_id", "cred_name", "date_start", "date_expiry",
		}

		// Create a stdlib CSV writer
		cw := csv.NewWriter(f)

		// Write header row and flush
		cw.Write(columns)
		cw.Flush()
		for _, entry := range rep.Entries {
			if err := cw.Write([]string{
				entry.OwnerName, entry.OwnerAppID, entry.OwnerObjID, string(entry.SourceKind),
				string(entry.CredKind), entry.CredKeyID.String(), entry.CredName,
				entry.DateStart.Format(time.RFC3339), entry.DateExpiry.Format(time.RFC3339),
			}); err != nil {
				return fmt.Errorf("writing CSV record: %w", err)
			}
			cw.Flush()
		}
	case config.FORMAT_JSON:
		// JSON format includes all fields for maximum fidelity and machine-readability
		jsonBytes, err := json.Marshal(rep.Entries)
		if err != nil {
			return fmt.Errorf("marshaling JSON: %w", err)
		}

		// Write JSON to output
		if _, err := fmt.Fprintf(f, "%s\n", string(jsonBytes)); err != nil {
			return fmt.Errorf("writing JSON output: %w", err)
		}
	}

	return nil
}
