package suggestion

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/texttheater/golang-levenshtein/levenshtein"
)

// SubcommandsRequiredWithSuggestions will ensure we have a subcommand provided by the user and augments it with
// suggestion for commands, alias and help on root command.
func SubcommandsRequiredWithSuggestions(cmd *cobra.Command, args []string) error {
	requireMsg := "unknown command %q for %q"
	typedName := ""
	// This will be triggered if cobra didn't find any subcommands.
	// Find some suggestions.
	var suggestions []string

	if len(args) != 0 && !cmd.DisableSuggestions {
		typedName += args[0]
		if cmd.SuggestionsMinimumDistance <= 0 {
			cmd.SuggestionsMinimumDistance = 2
		}
		// subcommand suggestions
		suggestions = cmd.SuggestionsFor(args[0])

		// subcommand alias suggestions (with distance, not exact)
		for _, c := range cmd.Commands() {
			if !c.IsAvailableCommand() {
				continue
			}

			candidate := suggestsByPrefixOrLd(typedName, c.Name(), cmd.SuggestionsMinimumDistance)
			if candidate == "" {
				continue
			}
			_, found := Find(suggestions, candidate)
			if !found {
				suggestions = append(suggestions, candidate)
			}
		}

		// help for root command
		if !cmd.HasParent() {
			candidate := suggestsByPrefixOrLd(typedName, "help", cmd.SuggestionsMinimumDistance)
			if candidate != "" {
				suggestions = append(suggestions, candidate)
			}
		}
	}

	var suggestionsMsg string
	if len(suggestions) > 0 {
		suggestionsMsg += "\nDid you mean this?\n"
		for _, s := range suggestions {
			suggestionsMsg += fmt.Sprintf("\t%v\n", s)
		}
	}

	if suggestionsMsg != "" {
		requireMsg = fmt.Sprintf("%s\n%s", requireMsg, suggestionsMsg)
		return fmt.Errorf(requireMsg, typedName, cmd.CommandPath())
	}

	return cmd.Help()
}

// suggestsByPrefixOrLd suggests a command by levenshtein distance or by prefix.
// It returns an empty string if nothing was found
func suggestsByPrefixOrLd(typedName, candidate string, minDistance int) string {
	levenshteinVariable := levenshtein.DistanceForStrings([]rune(typedName), []rune(candidate), levenshtein.DefaultOptions)
	suggestByLevenshtein := levenshteinVariable <= minDistance
	suggestByPrefix := strings.HasPrefix(strings.ToLower(candidate), strings.ToLower(typedName))
	if !suggestByLevenshtein && !suggestByPrefix {
		return ""
	}
	return candidate
}

// Find takes a slice and looks for an element in it. If found it will
// return it's key, otherwise it will return -1 and a bool of false.
func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}
