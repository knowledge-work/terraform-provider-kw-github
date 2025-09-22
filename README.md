# Terraform Provider KW GitHub

A Terraform provider for managing GitHub repository rulesets with enhanced support for merge method configurations.

## Installation

```hcl
terraform {
  required_providers {
    kwgithub = {
      source = "knowledge-work/kw-github"
      // version = "0.0.5"
    }
  }
}
```

## Authentication

### Personal Access Token

```hcl
provider "kwgithub" {
  token = var.github_token
  owner = "knowledge-work"
}
```

### GitHub App Installation

```hcl
provider "kwgithub" {
  owner = "knowledge-work"
  app_auth {}
}
```

Set environment variables:

- `GITHUB_APP_ID`
- `GITHUB_APP_INSTALLATION_ID`
- `GITHUB_APP_PEM_FILE`

## Usage

```hcl
resource "kwgithub_ruleset_allowed_merge_methods" "example" {
  repository = "repo"
  ruleset_id = "123456"
  allowed_merge_methods = ["merge", "squash"]
}
```

## Why This Provider?

The official GitHub provider resets `allowed_merge_methods` when updating other ruleset rules. This provider automatically detects and restores the expected configuration.

## License

Apache License 2.0
