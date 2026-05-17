package commands

import (
	"encoding/json"
	"os"
	"sort"

	"github.com/spf13/cobra"
)

type agentContext struct {
	Commands []agentCommand `json:"commands"`
}

type agentCommand struct {
	Name        string            `json:"name"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Subcommands []agentCommand    `json:"subcommands,omitempty"`
}

func addAgentContext(root *cobra.Command) {
	cmd := &cobra.Command{
		Use:         "agent-context",
		Short:       "Emit structured JSON describing this CLI for agents",
		Annotations: map[string]string{"mcp:read-only": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return json.NewEncoder(os.Stdout).Encode(agentContext{Commands: collectAgentCommands(root)})
		},
	}
	root.AddCommand(cmd)
}

func collectAgentCommands(cmd *cobra.Command) []agentCommand {
	children := cmd.Commands()
	sort.Slice(children, func(i, j int) bool { return children[i].Name() < children[j].Name() })
	out := make([]agentCommand, 0, len(children))
	for _, child := range children {
		if child.Name() == "agent-context" {
			continue
		}
		entry := agentCommand{Name: child.Name()}
		if len(child.Annotations) > 0 {
			entry.Annotations = child.Annotations
		}
		if child.HasSubCommands() {
			entry.Subcommands = collectAgentCommands(child)
		}
		out = append(out, entry)
	}
	return out
}
