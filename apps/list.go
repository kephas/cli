package apps

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/errgo.v1"
	"github.com/Scalingo/cli/config"
	"github.com/Scalingo/cli/io"
)

func List() error {
	c := config.ScalingoClient()
	apps, err := c.AppsList()
	if err != nil {
		return errgo.Mask(err, errgo.Any)
	}

	if len(apps) == 0 {
		fmt.Println(io.Indent("\nYou haven't created any app yet, create your first application using:\n→ scalingo create <app_name>\n", 2))
		return nil
	}

	t := tablewriter.NewWriter(os.Stdout)
	t.SetHeader([]string{"Name", "Role", "Owner"})

	for _, app := range apps {
		if app.Owner.Email == config.AuthenticatedUser.Email {
			t.Append([]string{app.Name, "owner", "-"})
		} else {
			t.Append([]string{app.Name, "collaborator", fmt.Sprintf("%s <%s>", app.Owner.Username, app.Owner.Email)})
		}
	}
	t.Render()

	return nil
}
