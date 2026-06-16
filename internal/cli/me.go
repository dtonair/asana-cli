package cli

import (
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
)

func newMeCommand() *cobra.Command {
	var optFields string
	cmd := &cobra.Command{
		Use:   "me",
		Short: "Get the current authenticated Asana user",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, _, err := buildClient()
			if err != nil {
				return err
			}
			ctx, cancel := withTimeout(cmd)
			defer cancel()

			q := url.Values{}
			appendOptFields(q, optFields)
			data, err := requestData(ctx, c, http.MethodGet, "/users/me"+querySuffix(q), nil)
			if err != nil {
				return err
			}
			return writeSuccess(cmd.OutOrStdout(), data, opts.human, summarizeUser(data))
		},
	}
	cmd.Flags().StringVar(&optFields, "opt-fields", "", "comma-separated Asana opt_fields")
	return cmd
}
