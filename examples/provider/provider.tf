provider "kwgithub" {
  token = var.github_token
}

# GitHub App authentication example
provider "kwgithub" {
  app_auth {
    id              = var.github_app_id
    installation_id = var.github_app_installation_id
    pem_file        = var.github_app_pem_file
  }
}