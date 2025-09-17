resource "kwgithub_ruleset_allowed_merge_methods" "example" {
  repository = "owner/repo"
  ruleset_id = "12345"
  allowed_merge_methods = ["merge", "squash", "rebase"]
}