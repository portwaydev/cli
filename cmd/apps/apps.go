package apps

import (
	"github.com/spf13/cobra"
)

func NewAppsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apps",
		Short: "Apps commands",
		Long:  "Commands for managing apps in Portway",
	}

	cmd.AddCommand(NewListCmd())

	return cmd
}

func NewListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List apps",
		Long:  "List apps in Portway",
		Run: func(cmd *cobra.Command, args []string) {
			// client, err := api.NewViperClientWithResponses()
			// if err != nil {
			// 	pterm.Error.Println(err)
			// 	return
			// }

			// whoami, err := client.GetApiV1WhoamiWithResponse(context.Background())
			// if err != nil {
			// 	pterm.Error.Println(err)
			// 	return
			// }

			// if whoami == nil {
			// 	pterm.Error.Println("Failed to get organization")
			// 	return
			// }
			// if whoami.StatusCode() != 200 {
			// 	pterm.Error.Println("Failed to get organization")
			// 	return
			// }

			// organizationId := whoami.JSON200.Organization.Id

			// apps, err := client.ListAppsV1WithResponse(context.Background(), organizationId)
			// if err != nil {
			// 	pterm.Error.Println(err)
			// 	return
			// }

			// if apps.StatusCode() != 200 {
			// 	pterm.Error.Println("Failed to list apps")
			// 	return
			// }

			// if apps.JSON200 == nil || len(*apps.JSON200) == 0 {
			// 	pterm.Info.Println("No apps found")
			// 	return
			// }

			// tableData := pterm.TableData{
			// 	{"ID", "Name", "Slug", "Created At"},
			// }
			// for _, app := range *apps.JSON200 {
			// 	tableData = append(tableData, []string{
			// 		app.Id.String(),
			// 		app.Name,
			// 		app.Slug,
			// 		app.CreatedAt.Format(time.RFC3339),
			// 	})
			// }
			// pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(tableData).Render()
		},
	}

	return cmd
}
