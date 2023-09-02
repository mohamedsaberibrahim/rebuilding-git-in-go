/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package command

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mohamedsaberibrahim/gity/database"
	"github.com/mohamedsaberibrahim/gity/reference"
	"github.com/spf13/cobra"
)

var message string

// commitCmd represents the commit command
var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "record changes to a gity repository.",
	Long: `Gity commits are a way to "record changes to a repository."
	A Gity repository is the collection of files tracked in the .gity folder of a project.
	In simple terms, a commit is a snapshot of your local repository. 

	It can be helpful to think of a commit as a checkpoint or savepoint for your project.
	In many video games, checkpoints are reached and your progress is saved after completing a specific action
	or challenge. Similarly, a Gity commit is usually performed after a significant contribution 
	is made to your project and you want to save your progress.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("commit called")

		dir, err := os.Getwd()
		git_path := strings.Join([]string{dir, database.METADATA_DIR}, string(os.PathSeparator))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to read the current directory - %v\n", err)
		}

		workspace := database.Workspace{}
		db := database.Database{}
		workspace.New(dir)
		db.New(strings.Join([]string{git_path, database.DATABASE_DIR}, string(os.PathSeparator)))
		entries := []database.Entry{}
		files_name := workspace.ListFiles("")
		for _, file_name := range files_name {

			content, err := workspace.ReadFile(file_name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to read the current directory - %v\n", err)
			}
			blob := database.Blob{}
			entry := database.Entry{}
			blob.New(content)
			err = db.Store(&blob)
			if err != nil {
				fmt.Print(err)
			}
			stat, err := workspace.StatFile(file_name)
			if err != nil {
				fmt.Print(err)
			}
			entry.New(file_name, blob.GetOid(), stat.Mode())
			entries = append(entries, entry)
		}

		root := database.Tree{}.Build(entries)
		root.Traverse(db.Store)
		author := database.Author{}
		author.New(os.Getenv("GIT_AUTHOR_NAME"), os.Getenv("GIT_AUTHOR_EMAIL"), time.Now())

		r := reference.Ref{}
		r.New(git_path)
		parent, _ := r.ReadHead()
		commit := database.Commit{}
		commit.New(parent, root.GetOid(), author, message)
		db.Store(&commit)

		err = r.UpdateHead(commit.GetOid())
		if err != nil {
			fmt.Print(err)
		}
		root_commit := ""
		if parent == nil {
			root_commit = "(root-commit) "
		}
		fmt.Printf("[%s%x] %s\n", root_commit, commit.GetOid(), commit.GetMessage())
	},
}

func init() {
	rootCmd.AddCommand(commitCmd)
	commitCmd.Flags().StringVarP(&message, "message", "m", "", "Use the given <msg> as the commit message.")
}
