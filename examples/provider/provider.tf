provider "kwgithub" {
  token = var.github_token
  owner = "knowledge-work"
}

# GitHub App authentication example
provider "kwgithub" {
  owner = "knowledge-work"
  app_auth {}
}