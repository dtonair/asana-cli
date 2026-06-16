package cli

import (
	"net/url"
	"strconv"

	"github.com/spf13/cobra"
)

func newListWorkspacesCommand() *cobra.Command {
	var (
		limit     int
		optFields string
	)
	cmd := &cobra.Command{
		Use:   "list-workspaces",
		Short: "List workspaces visible to the authenticated user",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateLimit(limit); err != nil {
				return err
			}
			c, _, err := buildClient()
			if err != nil {
				return err
			}
			ctx, cancel := withTimeout(cmd)
			defer cancel()

			q := url.Values{}
			q.Set("limit", strconv.Itoa(pageSize))
			appendOptFields(q, optFields)
			items, err := c.Paginate(ctx, "/workspaces"+querySuffix(q), limit, maxPages)
			if err != nil {
				return err
			}
			human := humanList(items, summarizeWorkspace, "No workspaces found.")
			return writeSuccess(cmd.OutOrStdout(), items, opts.human, human)
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 20, "maximum items to return (1-100)")
	cmd.Flags().StringVar(&optFields, "opt-fields", "", "comma-separated Asana opt_fields")
	return cmd
}
