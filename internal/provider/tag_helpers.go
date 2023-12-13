// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
)

// fetchWorkspaceTag returns the tag association in a workspace by name
func fetchWorkspaceTag(name, workspace, organization string, client *tfe.Client) (*tfe.Tag, error) {
	ws, err := client.Workspaces.Read(ctx, organization, workspace)
	if err != nil {
		return nil, fmt.Errorf("Error reading configuration of workspace %s in organization %s: %w", workspace, organization, err)
	}

	options := &tfe.WorkspaceTagListOptions{}
	// add queried tag name
	for {
		wsTags, err := client.Workspaces.ListTags(ctx, ws.ID, options)
		if err != nil {
			return nil, fmt.Errorf("Error retrieving workspace tags: %w", err)
		}
		for _, wsTag := range wsTags.Items {
			if wsTag != nil && wsTag.Name == name {
				return wsTag, nil
			}
		}

		// Exit the loop when we've seen all pages.
		if wsTags.CurrentPage >= wsTags.TotalPages {
			break
		}

		// Update the page number to get the next page.
		options.PageNumber = wsTags.NextPage
	}

	return nil, fmt.Errorf("could not find tag %s for workspace %s in organization %s", name, workspace, organization)
}
