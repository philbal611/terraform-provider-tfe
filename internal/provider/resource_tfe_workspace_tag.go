// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	// TODO: determine which imports are still required
	"context"
	"fmt"
	"log"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFEWorkspaceTag() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEWorkspaceTagCreate,
		Read:   resourceTFEWorkspaceTagRead,
		Delete: resourceTFEWorkspaceTagDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFEWorkspaceTagImporter,
		},

		CustomizeDiff: func(c context.Context, d *schema.ResourceDiff, meta interface{}) error {
			if err := validateTagNames(c, d); err != nil {
				return err
			}

			return nil
		},

		Schema: map[string]*schema.Schema{
			"tag_name": {
				Description: "The name of the tag to attach to the Workspace.",

				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},

			"workspace_id": {
				Description: "The id of the workspace to attach the tag to.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
			},
		},
	}
}

func resourceTFEWorkspaceTagCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	tagName := d.Get("tag_name").(string)
	workspaceID := d.Get("workspace_id").(string)

	ws, err := config.Client.Workspaces.ReadByID(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf(
			"Error retrieving workspace %s: %w", workspaceID, err)
	}

	options := tfe.WorkspaceAddTagsOptions{}
	options.Tags = append(options.Tags, &tfe.Tag{Name: tagName})

	log.Printf("[DEBUG] Create tag %s in workspace %s", tagName, ws.ID)
	err = config.Client.Workspaces.AddTags(ctx, ws.ID, options)
	if err != nil {
		return fmt.Errorf("Error creating tag %s in workspace %s: %w", tagName, ws.ID, err)
	}

	return resourceTFEWorkspaceTagRead(d, meta)
}

func resourceTFEWorkspaceTagDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	tagName := d.Get("tag_name").(string)
	workspaceID := d.Get("workspace_id").(string)

	log.Printf("[DEBUG] Delete tag %s in workspace %s", d.Id(), workspaceID)

	options := tfe.WorkspaceRemoveTagsOptions{}
	options.Tags = append(options.Tags, &tfe.Tag{Name: tagName})

	err := config.Client.Workspaces.RemoveTags(ctx, workspaceID, options)
	if err != nil && !isErrResourceNotFound(err) {
		return fmt.Errorf("Error deleting tag %s in workspace %s: %w", d.Id(), workspaceID, err)
	}

	return nil
}

func resourceTFEWorkspaceTagRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	tagName := d.Get("tag_name").(string)
	workspaceID := d.Get("workspace_id").(string)

	_, err := config.Client.Workspaces.ReadByID(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf(
			"Error retrieving workspace %s: %w", workspaceID, err)
	}

	// Get workspace tags
	// Create an options struct.
	options := &tfe.WorkspaceTagListOptions{}

	wsTags, err := config.Client.Workspaces.ListTags(ctx, workspaceID, options)
	if err != nil {
		return fmt.Errorf("Error retrieving tags on workspace %s: %v", workspaceID, err)
	}

	// Iterate tags to find tagName
	for _, tag := range wsTags.Items {
		if tag.Name == tagName {
			// Found tag, set id.
			d.SetId(tag.ID)
			return nil
		}
	}

	return fmt.Errorf("could not find tag %s on workspace %s", tagName, workspaceID)
}

func resourceTFEWorkspaceTagImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(ConfiguredClient)

	s := strings.Split(d.Id(), "/")
	if len(s) != 3 {
		return nil, fmt.Errorf(
			"invalid tag input format: %s (expected <ORGANIZATION>/<WORKSPACE NAME>/<TAG NAME>)",
			d.Id(),
		)
	}

	wsTag, err := fetchWorkspaceTag(s[2], s[1], s[0], config.Client)
	if err != nil {
		return nil, err
	}

	ws, err := config.Client.Workspaces.Read(ctx, s[0], s[1])
	if err != nil {
		return nil, fmt.Errorf(
			"Error retrieving workspace %s: %w", s[1], err)
	}

	d.Set("workspace_id", ws.ID)
	d.Set("tag_name", wsTag.Name)
	d.SetId(wsTag.ID)

	return []*schema.ResourceData{d}, nil
}
