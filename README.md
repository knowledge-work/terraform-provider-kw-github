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

  # Recommended: GitHub resets allowed_merge_methods when ruleset is updated
  force_update = timestamp()

  depends_on = [github_repository_ruleset.example]
}
```

### ⚠️Important: Force Update Recommendation

It is strongly recommended to include `force_update = timestamp()` in your resource configuration. This ensures the resource is updated on every Terraform run, which is necessary because GitHub's API specification causes `allowed_merge_methods` to be reset whenever `github_repository_ruleset` is updated. Without `force_update`, your merge method configuration may be unexpectedly lost when other ruleset changes are applied.

## Why This Provider?

The official GitHub provider resets `allowed_merge_methods` when updating other ruleset rules. This provider automatically detects and restores the expected configuration.

## License

Apache License 2.0
