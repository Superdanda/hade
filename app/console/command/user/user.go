package user

import (
	"fmt"

	"github.com/Superdanda/hade/framework/cobra"
)

var UserCommand = &cobra.Command{
	Use:   "user",
	Short: "user",
	RunE: func(c *cobra.Command, args []string) error {
		container := c.GetContainer()
		fmt.Println(container)
		return nil
	},
}
