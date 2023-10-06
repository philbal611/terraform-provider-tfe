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
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tfe_workspace_tag.foobar", "tag_name", tagName),
					resource.TestCheckResourceAttrSet("data.tfe_workspace_tag.foobar", "workspace_id"),
				),
			},
		},
	})
}

func testAccTFEWorkspaceTagDataSourceConfig(orgName string, rInt int) string {
	return fmt.Sprintf(`
locals {
    organization_name = "%s"
}

resource "tfe_workspace" "foobar" {
	name         = "workspace-test-%d"
	organization = local.organization_name
}

data "tfe_workspace_tag" "foobar" {
	workspace_id      = resource.tfe_workspace.foobar.id
	tag_name          = "tag-test-%d"
}`, orgName, rInt, rInt)
}
