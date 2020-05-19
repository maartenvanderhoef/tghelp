package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	//	gohcl2 "github.com/hashicorp/hcl2/gohcl"
	gohcl2 "github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	hcl2write "github.com/hashicorp/hcl/v2/hclwrite"

	configs "github.com/hashicorp/terraform/configs"

	//#"github.com/hashicorp/hcl2/hclparse"

	"github.com/maartenvanderhoef/tghelp/utils"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var defaultLogger *log.Logger

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tghelp",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// startCmd represents the start command
var terragruntCmd = &cobra.Command{
	Use:   "parse",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		path, err := os.Getwd()
		utils.CheckAndExit(err)
		process(path)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	//log.SetOutput(ioutil.Discard)

	defaultLogger = log.New(os.Stdout,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	defaultLogger.SetOutput(ioutil.Discard)
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(terragruntCmd)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.tghelp.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".tghelp" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".tghelp")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func process(folder string) {
	path := filepath.ToSlash(filepath.Join(folder, DefaultTerragruntConfigPath))
	root := parseterragrunthclfile(&path)

	for _, dep := range root.TerragruntDependencies {
		path := filepath.ToSlash(filepath.Join(dep.ConfigPath, DefaultTerragruntConfigPath))
		t := parseterragrunthclfile(&path)
		terraformDownloadFolder, err := downloadSource(t)
		utils.CheckAndExit(err)
		processTerraformFilesDependencies(terraformDownloadFolder, dep.Name)
	}
	terraformDownloadFolder, err := downloadSource(root)
	utils.CheckAndExit(err)

	processTerraformFiles(terraformDownloadFolder)
}

func parseterragrunthclfile(path *string) *terragruntConfigFile {
	defaultLogger.Printf("Processing ..%s file", *path)
	parser := hclparse.NewParser()
	f, parseDiags := parser.ParseHCLFile(*path)
	if parseDiags.HasErrors() {
		log.Fatal(parseDiags.Error())
	}
	terragruntConfig := &terragruntConfigFile{}

	decodeDiags := gohcl2.DecodeBody(f.Body, nil, terragruntConfig)
	if decodeDiags.HasErrors() {
		log.Fatal(decodeDiags.Error())
	}
	terragruntConfig.Path = path
	return terragruntConfig
}

func processTerraformFilesDependencies(folder string, module string) error {
	// Read whole Terraform Config folder
	parser := configs.NewParser(nil)
	mod, _ := parser.LoadConfigDir(folder)
	fmt.Println("## Dependencies Module " + module)

	for name, o := range mod.Outputs {
		fmt.Println("# " + o.Description)
		fmt.Println("# dependency." + module + ".outputs." + name)
		fmt.Println("")
	}
	return nil
}

func processTerraformFiles(folder string) error {
	// Read whole Terraform Config folder
	parser := configs.NewParser(nil)
	mod, _ := parser.LoadConfigDir(folder)

	for name, myvar := range mod.Variables {
		hf := hcl2write.NewEmptyFile()
		hf.Body().SetAttributeValue(name, myvar.Default)
		//fmt.Printf("Default :\n%s\n", string(hf.Bytes()))
		fmt.Println("# name: " + name + " -- desc: " + myvar.Description + " -- type:" + myvar.Type.FriendlyName())
		if !myvar.Default.IsNull() {
			fmt.Print("# ")
		}
		fmt.Print(string(hf.Bytes()))
		fmt.Println("")
	}
	return nil
}
