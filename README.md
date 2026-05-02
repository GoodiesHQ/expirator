# expirator

A CLI tool for reporting expiring (and expired) credentials across an Azure tenant. It queries Microsoft Graph to enumerate client secrets, app certificates, and SAML signing certificates for both **App Registrations** and **Enterprise Applications**, then emits a sorted report.

Read-only — never reads secret values, only metadata.

---

## Azure Setup

Create an App Registration in Azure AD and grant it the following API permission:

- **Microsoft Graph** → **Application Permissions** → `Application.Read.All` (admin consent required)

Remove all other permissions (including delegated permissions), then create a client secret and note the values below.

### `.env` / environment variables

```env
AZ_TENANT_ID=<your-tenant-id>
AZ_CLIENT_ID=<your-client-id>
AZ_CLIENT_SECRET=<your-client-secret>
```

Place a `.env` file in the same directory as the binary, or export the variables in your shell. Flags also accept these values directly (see [Flags](#flags)).

---

## Usage

```
expirator [flags]
```

### Flags

| Flag | Alias | Environment Variable | Default | Description |
|------|-------|---------|---------|-------------|
| `--az-tenant-id` | | `AZ_TENANT_ID` | *(required)* | Azure tenant ID |
| `--az-client-id` | | `AZ_CLIENT_ID` | *(required)* | Azure application client ID |
| `--az-client-secret` | | `AZ_CLIENT_SECRET` | *(required)* | Azure client secret |
| `--expiring-within-days` | `-e` | `EXPIRING_WITHIN_DAYS` | `0` (all) | Only show credentials expiring within N days |
| `--include-expired` | `-x` | `INCLUDE_EXPIRED` | `false` | Include already-expired credentials in output |
| `--group-sources` | `-g` | `GROUP_SOURCES` | `false` | Group output by source app instead of sorting by expiry date |
| `--format` | `-f` | `OUTPUT_FORMAT` | `table` | Output format: `table`, `csv`, or `json` |
| `--output-file` | `-o` | `OUTPUT_FILE` | stdout | Write output to a file instead of stdout |

---

## Examples

### Show all credentials expiring within 90 days (table)

```
expirator -e 90
```

### Include already-expired credentials in the report

```
expirator -x
```

### Export everything to JSON

```
expirator -f json -o report.json
```

### Export credentials expiring within 30 days (or already exampled) to CSV, grouped by source app

```
expirator -e 30 -x -g -f csv -o expiring.csv
```

---

## Output

### Table (default)

Sorted by expiry date. Expired credentials are highlighted in red.

```
+-----------------+---------------------+-----------------------------+-----------------+------------------------+
| CRED SOURCE     | APPLICATION         | CRED NAME                   | CRED TYPE       | EXPIRY                 |
+-----------------+---------------------+-----------------------------+-----------------+------------------------+
| AppRegistration | My API Gateway      | gateway-secret              | ClientSecret    | EXPIRED on 2024-11-30  |
| EnterpriseApp   | Acme SSO            | CN=Microsoft Azure Fed SSO  | SAMLSigningCert | EXPIRED on 2025-02-14  |
| AppRegistration | Monitoring Agent    | agent-cert                  | AppCertificate  | expires on 2025-08-01  |
| EnterpriseApp   | Corporate VPN       | CN=VPN Root CA              | SPCertificate   | expires on 2025-12-15  |
| AppRegistration | Notification Relay  | relay-secret                | ClientSecret    | expires on 2026-03-22  |
| AppRegistration | Backup Service      | backup-key                  | ClientSecret    | expires on 2027-04-10  |
+-----------------+---------------------+-----------------------------+-----------------+------------------------+
```

### CSV

Includes all fields. Suitable for importing into a spreadsheet or downstream tooling.

```csv
owner_name,owner_app_id,owner_object_id,source_kind,cred_kind,cred_key_id,cred_name,date_start,date_expiry
My API Gateway,11111111-0000-0000-0000-000000000001,22222222-0000-0000-0000-000000000001,AppRegistration,ClientSecret,33333333-0000-0000-0000-000000000001,gateway-secret,2023-11-30T00:00:00Z,2024-11-30T00:00:00Z
Acme SSO,11111111-0000-0000-0000-000000000002,22222222-0000-0000-0000-000000000002,EnterpriseApp,SAMLSigningCert,33333333-0000-0000-0000-000000000002,CN=Microsoft Azure Fed SSO,2024-02-14T00:00:00Z,2025-02-14T00:00:00Z
```

### JSON

Includes all fields. Suitable for piping into `jq` or other tooling.

```json
[
  {
    "owner_name": "My API Gateway",
    "owner_app_id": "11111111-0000-0000-0000-000000000001",
    "owner_obj_id": "22222222-0000-0000-0000-000000000001",
    "source_kind": "AppRegistration",
    "cred_kind": "ClientSecret",
    "cred_key_id": "33333333-0000-0000-0000-000000000001",
    "cred_name": "gateway-secret",
    "date_start": "2023-11-30T00:00:00Z",
    "date_expiry": "2024-11-30T00:00:00Z"
  },
  {
    "owner_name": "Acme SSO",
    "owner_app_id": "11111111-0000-0000-0000-000000000002",
    "owner_obj_id": "22222222-0000-0000-0000-000000000002",
    "source_kind": "EnterpriseApp",
    "cred_kind": "SAMLSigningCert",
    "cred_key_id": "33333333-0000-0000-0000-000000000002",
    "cred_name": "CN=Microsoft Azure Fed SSO",
    "date_start": "2024-02-14T00:00:00Z",
    "date_expiry": "2025-02-14T00:00:00Z"
  }
]
```

---

## Credential types

| Type | Source | Description |
|------|--------|-------------|
| `ClientSecret` | AppRegistration | Password credential on an app registration |
| `AppCertificate` | AppRegistration | Certificate credential on an app registration |
| `SPSecret` | EnterpriseApp | Password credential on a service principal |
| `SPCertificate` | EnterpriseApp | Certificate credential on a service principal |
| `SAMLSigningCert` | EnterpriseApp | SAML signing certificate on an enterprise app |
