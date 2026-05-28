package cli

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "chenhai",
	Short: "Chenhai - 镇海静态博客生成器",
	Long:  "水墨古风、高度自定义的静态博客编译器，由碧蓝航线角色·镇海倾情打造。",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(cleanCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)
	newCmd.AddCommand(themeCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
