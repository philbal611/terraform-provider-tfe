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
					resource.TestCheckResourceAttr("data.tfe_workspace_tag.foobar", "tag_name", tagName),
					resource.TestCheckResourceAttrSet("data.tfe_workspace_tag.foobar", "id"),
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

resource "tfe_workspace" "foobar" {
	name         = "workspace-test-%d"
	organization = tfe_organization.foobar.name

	tag_names = [local.tag_name]
}

data "tfe_workspace_tag" "foobar" {
	workspace_id      = tfe_workspace.foobar.id
	tag_name          = local.tag_name
}`, tagName, rInt, rInt)
}
