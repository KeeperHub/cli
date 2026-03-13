package protocol

import (
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/keeperhub/cli/internal/output"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewGetCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get <protocol-slug>",
		Short:   "Get a protocol",
		Aliases: []string{"g"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := args[0]

			protocols, err := loadProtocols(f, false, cmd)
			if err != nil {
				return err
			}

			var found *Protocol
			for i := range protocols {
				if protocols[i].Slug == slug {
					found = &protocols[i]
					break
				}
			}

			if found == nil {
				return cmdutil.NotFoundError{Err: fmt.Errorf("protocol %q not found", slug)}
			}

			p := output.NewPrinter(f.IOStreams, cmd)
			return p.PrintData(found, func(_ table.Writer) {
				renderProtocolDetail(f, found)
			})
		},
	}

	return cmd
}

// renderProtocolDetail writes a full reference card for the protocol to stdout.
func renderProtocolDetail(f *cmdutil.Factory, proto *Protocol) {
	fmt.Fprintf(f.IOStreams.Out, "%s\n", proto.Name)
	if proto.Description != "" {
		fmt.Fprintf(f.IOStreams.Out, "%s\n", proto.Description)
	}
	fmt.Fprintln(f.IOStreams.Out)

	for _, action := range proto.Actions {
		fmt.Fprintf(f.IOStreams.Out, "  %s\n", action.Name)
		if action.Description != "" {
			fmt.Fprintf(f.IOStreams.Out, "    %s\n", action.Description)
		}

		if len(action.Fields) > 0 {
			fmt.Fprintln(f.IOStreams.Out)
			printFieldsTable(f, action.Fields)
		}
		fmt.Fprintln(f.IOStreams.Out)
	}
}

// printFieldsTable renders an action's fields as a compact table.
func printFieldsTable(f *cmdutil.Factory, fields []Field) {
	tw := output.NewTable(f.IOStreams.Out, false)
	tw.AppendHeader(table.Row{"NAME", "TYPE", "REQUIRED", "DESCRIPTION"})
	for _, field := range fields {
		req := "no"
		if field.Required {
			req = "yes"
		}
		desc := strings.TrimSpace(field.Description)
		tw.AppendRow(table.Row{field.Name, field.Type, req, desc})
	}
	tw.Render()
}
