package main

import (
	"fmt"
	"os"
	"github.com/cloudfoundry/cli/plugin"
	"github.com/krujos/usagereport-plugin/apihelper"
)

//UsageReportCmd the plugin
type UsageReportCmd struct {
	apiHelper apihelper.CFAPIHelper
	cli       plugin.CliConnection
}

type org struct {
	name        string
	memoryQuota int
	memoryUsage int
	spaces      []space
}

type space struct {
	apps []app
	name string
}

type app struct {
	ram       int
	instances int
	running   bool
	name string
	buildpack string
	buildpack_detected string
}

//GetMetadata returns metatada
func (cmd *UsageReportCmd) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "snapshot",
		Version: plugin.VersionType{
			Major: 1,
			Minor: 0,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     "snapshot",
				HelpText: "Report AI and memory usage for orgs and spaces",
				UsageDetails: plugin.Usage{
					Usage: "cf snapshot",
				},
			},
		},
	}
}

//UsageReportCommand doer
func (cmd *UsageReportCmd) UsageReportCommand(args []string) {

	if nil == cmd.cli {
		fmt.Println("ERROR: CLI Connection is nil!")
		os.Exit(1)
	}

	orgs, err := cmd.getOrgs()
	if nil != err {
		fmt.Println(err)
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Windows or Linux (w or l):")
	env, _ := reader.ReadString('\n')
	out_dir := " "
	
	if env.str == "l" or env.str == "L" {
		out_dir = "/home/"
	} else if env.str ="w" or env.str == "W" {
		out_dir = "C:\Users\Public\Desktop"
	} else {
		out_dir = "Error"
	}
	
	totalApps := 0
	totalInstances := 0
	fmt.Printf("App,Space,Org,Memory Usage,Total Instances, Status, Buildpack, Buildpack Detected\n")
	for _, org := range orgs {
		for _, space := range org.spaces {
			consumed := 0
			instances := 0
			runningApps := 0
			stoppedApps := 0
			runningInstances := 0
			var appStatus string
			for _, app := range space.apps {
				if app.running {
					consumed += int(app.instances * app.ram)
					runningApps++
					runningInstances += app.instances
					appStatus = "Running"
				} else {
					stoppedApps++
					appStatus = "Stopped"
				}
				instances += int(app.instances)
				fmt.Printf("%s,%s,%s,%d MB,%d,%s, %s, %s \n",
					app.name, space.name, org.name, consumed, app.instances, appStatus, app.buildpack, app.buildpack_detected )

			}
			totalInstances += instances
			totalApps += len(space.apps)
			totalInstances += instances
			totalApps += len(space.apps)
		}
	}

	
}

func (cmd *UsageReportCmd) getOrgs() ([]org, error) {
	rawOrgs, err := cmd.apiHelper.GetOrgs(cmd.cli)
	if nil != err {
		return nil, err
	}

	var orgs = []org{}

	for _, o := range rawOrgs {
		usage, err := cmd.apiHelper.GetOrgMemoryUsage(cmd.cli, o)
		if nil != err {
			return nil, err
		}
		quota, err := cmd.apiHelper.GetQuotaMemoryLimit(cmd.cli, o.QuotaURL)
		if nil != err {
			return nil, err
		}
		spaces, err := cmd.getSpaces(o.SpacesURL)
		if nil != err {
			return nil, err
		}

		orgs = append(orgs, org{
			name:        o.Name,
			memoryQuota: int(quota),
			memoryUsage: int(usage),
			spaces:      spaces,
		})
	}
	return orgs, nil
}

func (cmd *UsageReportCmd) getSpaces(spaceURL string) ([]space, error) {
	rawSpaces, err := cmd.apiHelper.GetOrgSpaces(cmd.cli, spaceURL)
	if nil != err {
		return nil, err
	}
	var spaces = []space{}
	for _, s := range rawSpaces {
		apps, err := cmd.getApps(s.AppsURL)
		if nil != err {
			return nil, err
		}
		spaces = append(spaces,
			space{
				apps: apps,
				name: s.Name,
			},
		)
	}
	return spaces, nil
}

func (cmd *UsageReportCmd) getApps(appsURL string) ([]app, error) {
	rawApps, err := cmd.apiHelper.GetSpaceApps(cmd.cli, appsURL)
	if nil != err {
		return nil, err
	}
	var apps = []app{}
	for _, a := range rawApps {
		apps = append(apps, app{
			instances: int(a.Instances),
			ram:       int(a.RAM),
			running:   a.Running,
			name:	   a.Name,
			buildpack: a.Buildpack,
			buildpack_detected: a.BuildpackDetected,
		})
	}
	return apps, nil
}

//Run runs the plugin
func (cmd *UsageReportCmd) Run(cli plugin.CliConnection, args []string) {
	if args[0] == "snapshot" {
		cmd.apiHelper = &apihelper.APIHelper{}
		cmd.cli = cli
		cmd.UsageReportCommand(args)
	}
}

func main() {
	plugin.Start(new(UsageReportCmd))
}
