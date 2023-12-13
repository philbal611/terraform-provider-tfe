// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTFEWorkspaceTagDataSource_basic(t *testing.T) {
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	tagName := fmt.Sprintf("tag-test-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTFEWorkspaceTagDataSourceConfig(rInt, tagName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tfe_workspace_tag.foo", "tag_name", tagName),
					resource.TestCheckResourceAttr("data.tfe_workspace_tag.bar", "tag_name", tagName),
					resource.TestCheckResourceAttrSet("data.tfe_workspace_tag.foo", "id"),
					resource.TestCheckResourceAttrSet("data.tfe_workspace_tag.bar", "id"),
				),
			},
		},
	})
}

func testAccTFEWorkspaceTagDataSourceConfig(rInt int, tagName string) string {
	return fmt.Sprintf(`
locals {
	tag_name = "%s"
}

resource "tfe_organization" "foobar" {
	name  = "org-%d"
	email = "admin@company.com"
}

resource "tfe_workspace" "foo" {
	name         = "workspace-test-foo-%d"
	organization = tfe_organization.foobar.name

	tag_names = [local.tag_name]
}

resource "tfe_workspace" "bar" {
	name         = "workspace-test-bar-%d"
	organization = tfe_organization.foobar.name
}

resource "tfe_workspace_tag" "bar" {
	workspace_id      = tfe_workspace.bar.id
	tag_name          = local.tag_name
}

data "tfe_workspace_tag" "foo" {
	workspace_id = tfe_workspace.foo.id
	tag_name     = local.tag_name
}

data "tfe_workspace_tag" "bar" {
	workspace_id = tfe_workspace.bar.id
	tag_name     = local.tag_name
	depends_on   = [tfe_workspace_tag.bar]
}`, tagName, rInt, rInt, rInt)
}
