# KW GitHub Provider

The KW GitHub provider is used to interact with GitHub resources specific to Knowledge Work organization.

## Authentication

The provider supports two authentication methods:

### Personal Access Token

```terraform
provider "kwgithub" {
  token = var.github_token
}
```

### GitHub App Installation

```terraform
provider "kwgithub" {
  app_auth {
    id              = var.github_app_id
    installation_id = var.github_app_installation_id
    pem_file        = var.github_app_pem_file
  }
}
```

## Configuration

### Arguments

* `token` - (Optional) GitHub personal access token. Can also be set via `GITHUB_TOKEN` environment variable.
* `github_base_url` - (Optional) GitHub base URL. Defaults to https://api.github.com. Can also be set via `GITHUB_BASE_URL` environment variable.
* `app_auth` - (Optional) Configuration block to use GitHub App installation token. Conflicts with `token`.
  * `id` - (Required) GitHub App ID. Can also be set via `GITHUB_APP_ID` environment variable.
  * `installation_id` - (Required) GitHub App installation ID. Can also be set via `GITHUB_APP_INSTALLATION_ID` environment variable.
  * `pem_file` - (Required) GitHub App private key PEM file contents. Can also be set via `GITHUB_APP_PEM_FILE` environment variable.

### Environment Variables

When using environment variables, an empty `app_auth` block is required to allow provider configurations from environment variables to be specified:

```terraform
provider "kwgithub" {
  app_auth {}
}
```

Set the following environment variables:
- `GITHUB_APP_ID`
- `GITHUB_APP_INSTALLATION_ID`
- `GITHUB_APP_PEM_FILE`

Note: If you have a PEM file on disk, you can pass it in via `pem_file = file("path/to/file.pem")`.