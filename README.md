# Terraform Provider KW GitHub

[![Go Reference](https://pkg.go.dev/badge/github.com/knowledge-work/terraform-provider-kw-github.svg)](https://pkg.go.dev/github.com/knowledge-work/terraform-provider-kw-github)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

A Terraform provider for managing GitHub repository rulesets with enhanced support for merge method configurations.

## Features

- **Merge Methods Management**: Configure allowed merge methods for GitHub repository rulesets
- **Auto-Recovery**: Automatically detects and restores merge methods when they are reset by GitHub
- **Dependency Handling**: Properly handles dependencies with other ruleset resources
- **Force Update**: Manual trigger for updates when dependent resources change

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.23.0 (for development)
- GitHub Personal Access Token with appropriate permissions

## Installation

### Using the Provider

To use this provider in your Terraform configuration:

```hcl
terraform {
  required_providers {
    kwgithub = {
      source = "knowledge-work/kwgithub"
      version = "~> 0.0.2"
    }
  }
}
```

### Development Setup

1. Clone the repository:

```bash
git clone https://github.com/knowledge-work/terraform-provider-kw-github.git
cd terraform-provider-kw-github
```

2. Build the provider:

```bash
go build -o terraform-provider-kw-github ./cmd/terraform-provider-kwgithub
```

3. Install the provider locally:

```bash
make install
```

Or use the development override in your `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "knowledge-work/kwgithub" = "/path/to/your/terraform-provider-kw-github"
  }
  direct {}
}
```

## Usage

### Provider Configuration

```hcl
terraform {
  required_providers {
    kwgithub = {
      source = "knowledge-work/kwgithub"
      version = "~> 0.0.2"
    }
  }
}

provider "kwgithub" {
  # GitHub token (optional, can also use GITHUB_TOKEN environment variable)
  token = var.github_token

  # GitHub base URL (optional, defaults to https://api.github.com)
  github_base_url = "https://api.github.com"
}
```

### Authentication

The provider supports two authentication methods:

1. **Provider Configuration** (Recommended):

   ```hcl
   provider "kwgithub" {
     token = var.github_token
   }
   ```

2. **Environment Variable**:

   ```bash
   export GITHUB_TOKEN="your_github_token"
   ```

The provider will use the `token` attribute if provided, otherwise it will fall back to the `GITHUB_TOKEN` environment variable.

## Resources

### `kwgithub_ruleset_allowed_merge_methods`

Manages allowed merge methods for a GitHub repository ruleset.

#### Example Usage

```hcl
resource "kwgithub_ruleset_allowed_merge_methods" "example" {
  repository = "owner/repo"
  ruleset_id = "123456"
  allowed_merge_methods = ["merge", "squash"]

  # Optional: Force update when dependent resources change
  force_update = timestamp()

  # Ensure this runs after other ruleset changes
  depends_on = [
    kwgithub_other_ruleset_resource.example
  ]
}
```

#### Arguments

- `repository` (Required) - Repository in the format "owner/repo"
- `ruleset_id` (Required) - ID of the GitHub ruleset
- `allowed_merge_methods` (Required) - Set of allowed merge methods. Valid values: `"merge"`, `"squash"`, `"rebase"`
- `force_update` (Optional) - Timestamp or value to force update when dependent resources change

#### Attributes

- `id` - Resource identifier in the format "owner/repo:ruleset_id"

#### Import

```bash
terraform import kwgithub_ruleset_allowed_merge_methods.example owner/repo:123456
```

## Why This Provider?

While the official GitHub Terraform provider (`integrations/github`) supports repository rulesets via `github_repository_ruleset`, it has a limitation with merge method configurations:

**The Problem**: When you update other rules in a GitHub ruleset (like pull request rules, status checks, etc.), GitHub's API resets the `allowed_merge_methods` configuration to its default values. This causes Terraform state drift and unexpected behavior.

**Our Solution**: This specialized provider:

- **Detects** when merge methods have been reset by GitHub
- **Automatically restores** the expected configuration during the next Terraform run
- **Provides force update mechanism** to handle dependency changes
- **Works alongside** the official GitHub provider

## Use Cases

1. **Complement Official Provider**: Use `github_repository_ruleset` for main ruleset configuration and `kwgithub_ruleset_allowed_merge_methods` for reliable merge method management
2. **Existing Rulesets**: Manage merge methods for rulesets created outside of Terraform
3. **Complex Dependencies**: Handle scenarios where ruleset updates from other sources affect merge methods

## Examples

```hcl
terraform {
  required_providers {
    github = {
      source  = "integrations/github"
      version = "~> 6.0"
    }
    kwgithub = {
      source = "knowledge-work/kwgithub"
      version = "~> 0.0.2"
    }
  }
}

provider "github" {
  token = var.github_token
}

provider "kwgithub" {
  token = var.github_token
}

# Create ruleset with GitHub official provider
resource "github_repository_ruleset" "main" {
  repository = "myorg/myrepo"
  name       = "Main Branch Protection"
  target     = "branch"
  enforcement = "active"

  conditions {
    ref_name {
      include = ["refs/heads/main"]
      exclude = []
    }
  }

  rules {
    pull_request {
      required_approving_review_count = 2
      dismiss_stale_reviews_on_push   = true
      require_code_owner_review       = true
    }
    required_status_checks {
      required_check {
        context = "ci/tests"
      }
    }
  }
}

# Manage merge methods with our specialized provider
resource "kwgithub_ruleset_allowed_merge_methods" "main" {
  repository = "myorg/myrepo"
  ruleset_id = github_repository_ruleset.main.id
  allowed_merge_methods = ["merge", "squash"]

  # Option 1: Use etag for specific dependency tracking
  force_update = github_repository_ruleset.main.etag

  # Option 2: Use timestamp for general force updates
  # force_update = timestamp()

  depends_on = [
    github_repository_ruleset.main
  ]
}
```

#### Force Update Options

- **`github_repository_ruleset.main.etag`**: Recommended when you want updates triggered only when the specific ruleset changes
- **`timestamp()`**: Use when you want to force updates regardless of specific dependencies

## Development

### Building from Source

```bash
make build
# Or manually:
go build -o terraform-provider-kw-github ./cmd/terraform-provider-kwgithub
```

### Running Tests

```bash
# Unit tests
make test

# Acceptance tests (requires GITHUB_TOKEN)
make test-acc
```

### Code Quality

```bash
# Format and lint code
make lint

# Or run individual commands:
make fmt    # Format code
make vet    # Run go vet
make tidy   # Tidy dependencies
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework) for the provider development framework
- [go-github](https://github.com/google/go-github) for GitHub API client
- The Terraform community for inspiration and best practices
