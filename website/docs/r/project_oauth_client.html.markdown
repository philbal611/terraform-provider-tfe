---
layout: "tfe"
page_title: "Terraform Enterprise: tfe_project_oauth_client"
description: |-
    Add an oauth client to a project
---

# tfe_project_oauth_client

Adds and removes oauth clients from a project

## Example Usage

Basic usage:

```hcl
resource "tfe_organization" "test" {
  name  = "my-org-name"
  email = "admin@company.com"
}

resource "tfe_project" "test" {
  name         = "my-project-name"
  organization = tfe_organization.test.name
}

resource "tfe_oauth_client" "test" {
  organization     = tfe_organization.test
  api_url          = "https://api.github.com"
  http_url         = "https://github.com"
  oauth_token      = "oauth_token_id"
  service_provider = "github"
}

resource "tfe_project_oauth_client" "test" {
  oauth_client_id = tfe_oauth_client.test.id
  project_id    = tfe_project.test.id
}
```

## Argument Reference

The following arguments are supported:

* `oauth_client_id` - (Required) ID of the oauth client.
* `project_id` - (Required) Project ID to add the oauth client to.

## Attributes Reference

* `id` - The ID of the oauth client attachment. ID format: `<project-id>_<oauth-client-id>`

## Import

Project OAuth Clients can be imported; use `<ORGANIZATION>/<PROJECT ID>/<OAUTH CLIENT NAME>`. For example:

```shell
terraform import tfe_project_oauth_client.test 'my-org-name/project/oauth-client-name'
```
