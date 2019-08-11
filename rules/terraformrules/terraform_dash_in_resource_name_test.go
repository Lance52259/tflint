package terraformrules

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/configs/configload"
	"github.com/hashicorp/terraform/terraform"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/project"
	"github.com/wata727/tflint/tflint"
)

func Test_TerraformDashInResourceNameRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected issue.Issues
	}{
		{
			Name: "dash in resource name",
			Content: `
resource "aws_eip" "dash-name" {
}`,
			Expected: []*issue.Issue{
				{
					Detector: "terraform_dash_in_resource_name",
					Type:     issue.NOTICE,
					Message:  "`dash-name` resource name has a dash",
					Line:     2,
					File:     "resources.tf",
					Link:     project.ReferenceLink("terraform_dash_in_resource_name"),
				},
			},
		},
		{
			Name: "no dash in resource name",
			Content: `
resource "aws_eip" "no_dash_name" {
}`,
			Expected: []*issue.Issue{},
		},
	}

	dir, err := ioutil.TempDir("", "TerraformDashInResourceName")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(currentDir)

	err = os.Chdir(dir)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range cases {
		loader, err := configload.NewLoader(&configload.Config{})
		if err != nil {
			t.Fatal(err)
		}

		err = ioutil.WriteFile(dir+"/resources.tf", []byte(tc.Content), os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}

		mod, diags := loader.Parser().LoadConfigDir(".")
		if diags.HasErrors() {
			t.Fatal(diags)
		}
		cfg, tfdiags := configs.BuildConfig(mod, configs.DisabledModuleWalker)
		if tfdiags.HasErrors() {
			t.Fatal(tfdiags)
		}

		runner, err := tflint.NewRunner(tflint.EmptyConfig(), map[string]tflint.Annotations{}, cfg, map[string]*terraform.InputValue{})
		if err != nil {
			t.Fatal(err)
		}
		rule := NewTerraformDashInResourceNameRule()

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		if !cmp.Equal(tc.Expected, runner.Issues) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues))
		}
	}
}