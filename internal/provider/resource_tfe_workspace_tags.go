// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"log"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFEWorkspaceTags() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFEWorkspaceTagsCreate,
		Read:   resourceTFEWorkspaceTagsRead,
		Update: resourceTFEWorkspaceTagsUpdate,
		Delete: resourceTFEWorkspaceTagsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTFEWorkspaceTagsImporter,
		},

		CustomizeDiff: func(c context.Context, d *schema.ResourceDiff, meta interface{}) error {
			if err := validateTagNames(c, d); err != nil {
				return err
			}

			return nil
		},

		Schema: map[string]*schema.Schema{
			"tag_names": {
				Description: "The names of the tags to attach to the Workspace.",

				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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

func resourceTFEWorkspaceTagsCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	tagNames := d.Get("tag_names").(*schema.Set).List()
	workspaceID := d.Get("workspace_id").(string)

	ws, err := config.Client.Workspaces.ReadByID(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf(
			"Error retrieving workspace %s: %w", workspaceID, err)
	}

	options := tfe.WorkspaceAddTagsOptions{}
	for _, tagName := range tagNames {
		name := tagName.(string)
		options.Tags = append(options.Tags, &tfe.Tag{Name: name})
	}

	log.Printf("[DEBUG] Create tags %s in workspace %s", tagNames, ws.ID)
	err = config.Client.Workspaces.AddTags(ctx, ws.ID, options)
	if err != nil {
		return fmt.Errorf("Error creating tags %s on workspace %s: %w", tagNames, ws.ID, err)
	}

	return resourceTFEWorkspaceTagsRead(d, meta)
}

func resourceTFEWorkspaceTagsDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	tagNames := d.Get("tag_names").(*schema.Set).List()
	workspaceID := d.Get("workspace_id").(string)

	log.Printf("[DEBUG] Delete tags %s in workspace %s", tagNames, workspaceID)

	options := tfe.WorkspaceRemoveTagsOptions{}
	for _, tagName := range tagNames {
		name := tagName.(string)
		options.Tags = append(options.Tags, &tfe.Tag{Name: name})
	}

	err := config.Client.Workspaces.RemoveTags(ctx, workspaceID, options)
	if err != nil && !isErrResourceNotFound(err) {
		return fmt.Errorf("Error deleting tags %s from workspace %s: %w", tagNames, workspaceID, err)
	}

	return nil
}

func resourceTFEWorkspaceTagsRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)

	tagNames := d.Get("tag_names").(*schema.Set).List()
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
		return fmt.Errorf("Error retrieving tags from workspace %s: %v", workspaceID, err)
	}

	var rId string = ""
	var foundTagNames []string
	// Iterate tags to ensure they're found on the workspace
	for _, tagName := range tagNames {
		for _, wsTag := range wsTags.Items {
			if wsTag.Name == tagName {
				log.Printf("[DEBUG] Found tag %s on workspace %s", tagName, workspaceID)
				rId += "|" + wsTag.ID
				foundTagNames = append(foundTagNames, wsTag.Name)
				break
			}
		}
	}
	d.Set("tag_names", foundTagNames)
	d.SetId(strings.Trim(rId, "|"))

	return nil
}

func resourceTFEWorkspaceTagsUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(ConfiguredClient)
	workspaceID := d.Get("workspace_id").(string)

	if d.HasChange("tag_names") {
		oldTagNameValues, newTagNameValues := d.GetChange("tag_names")
		newTagNamesSet := newTagNameValues.(*schema.Set)
		oldTagNamesSet := oldTagNameValues.(*schema.Set)

		newTagNames := newTagNamesSet.Difference(oldTagNamesSet)
		oldTagNames := oldTagNamesSet.Difference(newTagNamesSet)

		// First add the new tags
		if newTagNames.Len() > 0 {
			var addTags []*tfe.Tag

			for _, tagName := range newTagNames.List() {
				name := tagName.(string)
				addTags = append(addTags, &tfe.Tag{Name: name})
			}

			log.Printf("[DEBUG] Adding tags to workspace: %s", workspaceID)
			err := config.Client.Workspaces.AddTags(ctx, workspaceID, tfe.WorkspaceAddTagsOptions{Tags: addTags})
			if err != nil {
				return fmt.Errorf("Error adding tags to workspace %s: %w", workspaceID, err)
			}
		}

		// Then remove all the old tags
		if oldTagNames.Len() > 0 {
			var removeTags []*tfe.Tag

			for _, tagName := range oldTagNames.List() {
				removeTags = append(removeTags, &tfe.Tag{Name: tagName.(string)})
			}

			log.Printf("[DEBUG] Removing tags from workspace: %s", workspaceID)
			err := config.Client.Workspaces.RemoveTags(ctx, workspaceID, tfe.WorkspaceRemoveTagsOptions{Tags: removeTags})
			if err != nil {
				return fmt.Errorf("Error removing tags from workspace %s: %w", workspaceID, err)
			}
		}
	}

	return resourceTFEWorkspaceTagsRead(d, meta)
}

func resourceTFEWorkspaceTagsImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(ConfiguredClient)

	// Split individual tag strings
	tagList := strings.Split(d.Id(), "|")
	var finalTagId string
	var finalTagNames []string
	for _, tagListItem := range tagList {
		// Split current tag string
		var tagItem = strings.Split(tagListItem, "/")
		if len(tagItem) != 3 {
			return nil, fmt.Errorf(
				"invalid tag input format: %s (expected <ORGANIZATION>/<WORKSPACE NAME>/<TAG NAME>|...)",
				d.Id(),
			)
		}

		wsTag, err := fetchWorkspaceTag(tagItem[2], tagItem[1], tagItem[0], config.Client)
		if err != nil {
			return nil, err
		}
		finalTagNames = append(finalTagNames, wsTag.Name)
		finalTagId += "|" + wsTag.ID
	}

	var org = strings.Split(tagList[0], "/")[0]
	var wsName = strings.Split(tagList[0], "/")[1]
	ws, err := config.Client.Workspaces.Read(ctx, org, wsName)
	if err != nil {
		return nil, fmt.Errorf(
			"Error retrieving workspace %s: %w", wsName, err)
	}

	d.Set("workspace_id", ws.ID)
	d.Set("tag_names", finalTagNames)
	d.SetId(strings.Trim(finalTagId, "|"))

	return []*schema.ResourceData{d}, nil
}
