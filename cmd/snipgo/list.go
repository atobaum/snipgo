package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all snippets",
	Long:  "Lists all snippets in a table format",
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	snippets := manager.GetAll()

	if len(snippets) == 0 {
		fmt.Println("No snippets found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tTitle\tTags\tLanguage\tFavorite")
	fmt.Fprintln(w, "---\t-----\t----\t--------\t--------")

	for _, snippet := range snippets {
		idShort := snippet.ID[:8]
		tags := strings.Join(snippet.Tags, ", ")
		if tags == "" {
			tags = "-"
		}
		language := snippet.Language
		if language == "" {
			language = "-"
		}
		favorite := "No"
		if snippet.IsFavorite {
			favorite = "Yes"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			idShort, snippet.Title, tags, language, favorite)
	}

	return w.Flush()
}

