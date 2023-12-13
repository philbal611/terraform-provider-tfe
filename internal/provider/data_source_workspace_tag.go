// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Read context to implement cancellation
//

package provider

import (
	"context"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTFEWorkspaceTag() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTFEWorkspaceTagRead,

		Schema: map[string]*schema.Schema{
			"tag_name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceTFEWorkspaceTagRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ConfiguredClient)

	tagName := d.Get("tag_name").(string)
	workspaceID := d.Get("workspace_id").(string)

	_, err := config.Client.Workspaces.ReadByID(ctx, workspaceID)
	if err != nil {
		return diag.Errorf(
			"Error retrieving workspace %s: %v", workspaceID, err)
	}

	// Create an options struct.
	options := &tfe.WorkspaceTagListOptions{}

	l, err := config.Client.Workspaces.ListTags(ctx, workspaceID, options)
	if err != nil {
		return diag.Errorf("Error retrieving tags on workspace %s: %v", workspaceID, err)
	}

	for _, tag := range l.Items {
		// Case-insensitive uniqueness is enforced in TFC
		if strings.EqualFold(tag.Name, tagName) {
			// Found tag, set id.
			d.SetId(tag.ID)
			return nil
		}
	}
	return diag.Errorf("could not find tag %s on workspace %s", tagName, workspaceID)
}
