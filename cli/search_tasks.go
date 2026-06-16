package cli

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/dtonair/asana-cli/asana"
)

func newSearchTasksCommand() *cobra.Command {
	var (
		workspaceGID string
		text         string
		assignee     string
		completed    bool
		limit        int
		optFields    string
	)
	cmd := &cobra.Command{
		Use:   "search-tasks",
		Short: "Search tasks in an Asana workspace (may require premium access)",
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
			if v := strings.TrimSpace(text); v != "" {
				q.Set("text", v)
			}
			if v := strings.TrimSpace(assignee); v != "" {
				q.Set("assignee.any", v)
			}
			// Tri-state: only send completed when the flag was explicitly set,
			// matching the extension's typeof === "boolean" check.
			if cmd.Flags().Changed("completed") {
				q.Set("completed", strconv.FormatBool(completed))
			}
			appendOptFields(q, optFields)

			path := "/workspaces/" + asana.EncodePathSegment(workspace) + "/tasks/search?" + q.Encode()
			items, err := c.Paginate(ctx, path, limit, maxPages)
			if err != nil {
				return err
			}
			human := humanList(items, summarizeTask, "No tasks found.")
			return writeSuccess(cmd.OutOrStdout(), items, opts.human, human)
		},
	}
	cmd.Flags().StringVar(&workspaceGID, "workspace-gid", "", "Asana workspace GID (defaults to ASANA_DEFAULT_WORKSPACE)")
	cmd.Flags().StringVar(&text, "text", "", "text search query")
	cmd.Flags().StringVar(&assignee, "assignee", "", "assignee GID, email, or me")
	cmd.Flags().BoolVar(&completed, "completed", false, "filter by completion state (omitted unless set)")
	cmd.Flags().IntVar(&limit, "limit", 20, "maximum items to return (1-100)")
	cmd.Flags().StringVar(&optFields, "opt-fields", "", "comma-separated Asana opt_fields")
	return cmd
}
