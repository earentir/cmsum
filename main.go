package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"cmsmgmt/joomla"
	"cmsmgmt/wordpress"

	"github.com/spf13/cobra"
)

var (
	cmsPath    string
	appVersion = "0.1.21"
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "cmsmgmt",
		Short:   "Content Management System Management",
		Long:    "Content Management System Management - https://github.com/earentir/cmsmgmt",
		Version: appVersion,

		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			if cmsPath != "" {
				if _, err := os.Stat(cmsPath); os.IsNotExist(err) {
					return fmt.Errorf("The specified CMS path does not exist: %s", cmsPath)
				}
			}
			return nil
		},
	}

	rootCmd.PersistentFlags().StringVarP(&cmsPath, "path", "p", "", "Path to the CMS root directory")

	usersCmd := &cobra.Command{
		Use:   "users",
		Short: "User management commands",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List users",
		Run: func(_ *cobra.Command, _ []string) {
			cmsType := detectCMS()
			if cmsType == "" {
				log.Fatal("Unable to detect CMS type. Make sure you're in the correct directory or specify the correct path using the -p flag.")
			}

			var err error
			switch cmsType {
			case "wordpress":
				err = wordpress.ProcessWordPress(cmsPath)
			case "joomla":
				db, cfg, defaultPrefix, err2 := joomla.ProcessJoomla(cmsPath)
				if err2 == nil {
					fmt.Printf("Joomla DB Name: %s\n", cfg.DBName)
					fmt.Printf("Joomla DB User: %s\n", cfg.User)
					fmt.Printf("Identified Joomla table prefixes: %v\n", defaultPrefix)

					users, err3 := joomla.ListUsers(db, defaultPrefix)
					if err3 != nil {
						log.Printf("list users for prefix %s: %v", defaultPrefix, err3)
						fmt.Println(fmt.Errorf("list users for prefix %s: %w", defaultPrefix, err3))
					} else {
						fmt.Printf("\nUsers for prefix '%s':\n", defaultPrefix)
						for _, u := range users {
							fmt.Printf("ID:%d  Username:%s  Name:%s  Email:%s  Roles:%v\n", u.ID, u.Username, u.Name, u.Email, u.Roles)
						}
					}
				}
				err = err2
			}

			if err != nil {
				log.Printf("Error processing %s: %v", cmsType, err)
			}
		},
	}

	userInfoCmd := &cobra.Command{
		Use:   "info",
		Short: "Show user info",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println("User info functionality not implemented yet.")
		},
	}

	editCmd := &cobra.Command{
		Use:   "edit [USERNAME]",
		Short: "Edit user details",
		Args:  cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			username := args[0]
			cmsType := detectCMS()
			if cmsType == "" {
				log.Fatal("Unable to detect CMS type. Make sure you're in the correct directory or specify the correct path using the -p flag.")
			}

			var err error
			switch cmsType {
			case "wordpress":
				err = wordpress.EditUser(cmsPath, username)
			case "joomla":
				db, _, defaultPrefix, err2 := joomla.ProcessJoomla(cmsPath)
				if err2 == nil {
					err = joomla.EditUser(db, defaultPrefix, cmsPath, username)
				} else {
					err = err2
				}
			}

			if err != nil {
				log.Printf("Error editing %s user: %v", cmsType, err)
			}
		},
	}

	usersCmd.AddCommand(listCmd)
	usersCmd.AddCommand(userInfoCmd)
	usersCmd.AddCommand(editCmd)

	infoCmd := &cobra.Command{
		Use:   "info",
		Short: "Show CMS information",
	}

	dbCmd := &cobra.Command{
		Use:   "db",
		Short: "Show db information",
		Run: func(_ *cobra.Command, _ []string) {
			cmsType := detectCMS()
			if cmsType == "" {
				log.Fatal("Unable to detect CMS type. Make sure you're in the correct directory or specify the correct path using the -p flag.")
			}

			var err error
			switch cmsType {
			case "wordpress":
				err = wordpress.ShowInfo(cmsPath)
			case "joomla":
				err = joomla.ShowInfo(cmsPath)
			}

			if err != nil {
				log.Printf("Error showing %s info: %v", cmsType, err)
			}
		},
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show CMS version information",
		Run: func(_ *cobra.Command, _ []string) {
			cmsType := detectCMS()
			if cmsType == "" {
				log.Fatal("Unable to detect CMS type. Make sure you're in the correct directory or specify the correct path using the -p flag.")
			}

			var version, rel string
			var err error
			switch cmsType {
			case "wordpress":
				version, err = wordpress.GetVersion(cmsPath)
			case "joomla":
				version, rel, err = joomla.GetVersion(cmsPath)
			}

			if err != nil {
				log.Printf("Error showing %s version: %v", cmsType, err)
			} else {
				fmt.Printf("%s Version: %s\n", cmsType, version)
				if cmsType == "joomla" {
					fmt.Printf("Release: %s\n", rel)
				}
			}
		},
	}

	infoCmd.AddCommand(dbCmd)
	infoCmd.AddCommand(versionCmd)

	rootCmd.AddCommand(usersCmd)
	rootCmd.AddCommand(infoCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func detectCMS() string {
	wpConfig := filepath.Join(cmsPath, "wp-config.php")
	joomlaConfig := filepath.Join(cmsPath, "configuration.php")

	if _, err := os.Stat(wpConfig); err == nil {
		return "wordpress"
	}
	if _, err := os.Stat(joomlaConfig); err == nil {
		return "joomla"
	}
	return ""
}
