package cli

import (
	"net/url"
	"strconv"

	"github.com/spf13/cobra"

	"asana-cli/internal/asana"
)

func newListProjectsCommand() *cobra.Command {
	var (
		workspaceGID string
		limit        int
		optFields    string
	)
	cmd := &cobra.Command{
		Use:   "list-projects",
		Short: "List projects in an Asana workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateLimit(limit); err != nil {
				return err
			}
			c, cfg, err := buildClient()
			if err != nil {
				return err
			}
			workspace, err := cfg.ResolveWorkspace(workspaceGID)
			if err != nil {
				return &usageError{err: err}
			}
			ctx, cancel := withTimeout(cmd)
			defer cancel()

			q := url.Values{}
			q.Set("limit", strconv.Itoa(pageSize))
			appendOptFields(q, optFields)
			path := "/workspaces/" + asana.EncodePathSegment(workspace) + "/projects" + querySuffix(q)
			items, err := c.Paginate(ctx, path, limit, maxPages)
			if err != nil {
				return err
			}
			human := humanList(items, summarizeProject, "No projects found.")
			return writeSuccess(cmd.OutOrStdout(), items, opts.human, human)
		},
	}
	cmd.Flags().StringVar(&workspaceGID, "workspace-gid", "", "Asana workspace GID (defaults to ASANA_DEFAULT_WORKSPACE)")
	cmd.Flags().IntVar(&limit, "limit", 20, "maximum items to return (1-100)")
	cmd.Flags().StringVar(&optFields, "opt-fields", "", "comma-separated Asana opt_fields")
	return cmd
}
