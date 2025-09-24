# Terraform Provider KW GitHub

A Terraform provider for managing GitHub repository rulesets with enhanced support for merge method configurations.

## Installation

```hcl
terraform {
  required_providers {
    kwgithub = {
      source = "knowledge-work/kw-github"
    }
  }
}
```

## Authentication

### Personal Access Token

```hcl
provider "kwgithub" {
  owner = "knowledge-work"
  token = var.github_token
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
resource "github_repository_ruleset" "example" {
  name        = "example-ruleset"
  repository  = "repo"
  target      = "branch"
  enforcement = "active"

  conditions {
    ref_name {
      include = ["~DEFAULT_BRANCH"]
      exclude = []
    }
  }

  rules {
    pull_request {
      required_approving_review_count   = 1
      dismiss_stale_reviews_on_push     = true
      require_code_owner_review         = true
      require_last_push_approval        = false
      required_review_thread_resolution = false
    }
  }
}

resource "kwgithub_ruleset_allowed_merge_methods" "example" {
  repository = "repo"
  ruleset_id = github_repository_ruleset.example.ruleset_id
  allowed_merge_methods = ["merge", "squash"]

  # Recommended: Update only when ruleset configuration changes
  force_update = sha256(jsonencode({
    name        = github_repository_ruleset.example.name
    target      = github_repository_ruleset.example.target
    enforcement = github_repository_ruleset.example.enforcement
    conditions  = github_repository_ruleset.example.conditions
    rules       = github_repository_ruleset.example.rules
  }))

  depends_on = [github_repository_ruleset.example]
}
```

### ⚠️ Important: Force Update Recommendation

It is strongly recommended to include a `force_update` parameter in your resource configuration. This ensures the resource is updated when the ruleset configuration changes, which is necessary because GitHub's API specification causes `allowed_merge_methods` to be reset whenever `github_repository_ruleset` is updated.

Use a hash of the ruleset configuration to trigger updates only when the ruleset actually changes:

```hcl
force_update = sha256(jsonencode({
  name        = github_repository_ruleset.example.name
  target      = github_repository_ruleset.example.target
  enforcement = github_repository_ruleset.example.enforcement
  conditions  = github_repository_ruleset.example.conditions
  rules       = github_repository_ruleset.example.rules
}))
```

This approach avoids unnecessary updates while ensuring merge method configuration is restored when needed.

## Why This Provider?

The official GitHub provider resets `allowed_merge_methods` when updating other ruleset rules. This provider automatically detects and restores the expected configuration.

## License

Apache License 2.0
