package main

import (
	hcl "github.com/hashicorp/hcl/v2"
)

const (
	TerragruntCacheDir          = ".terragrunt-cache"
	DefaultTerragruntConfigPath = "terragrunt.hcl"
)

// terragruntConfigFile represents the configuration supported in a Terragrunt configuration file (i.e.
// terragrunt.hcl)
type terragruntConfigFile struct {
	Path      *string
	Terraform *TerragruntTerraformConfig `hcl:"terraform,block"`
	//Inputs                 *cty.Value                 `hcl:"inputs,attr"`
	TerragruntDependencies []Dependency `hcl:"dependency,block"`
	Remain                 hcl.Body     `hcl:",remain"`
}

// type TerraformToplevel struct {
// 	Variables []variable `hcl:"variable,block"`
// 	Remain    hcl.Body   `hcl:",remain"`
// }

type Dependency struct {
	Name       string   `hcl:",label" cty:"name"`
	ConfigPath string   `hcl:"config_path,attr" cty:"config_path"`
	Remain     hcl.Body `hcl:",remain"`
}

type TerragruntTerraformConfig struct {
	Source *string  `hcl:"source,attr"`
	Remain hcl.Body `hcl:",remain"`
}
