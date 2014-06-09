package ranchr

import (
	"errors"
	_ "fmt"
	"os"
	_ "reflect"

	"github.com/BurntSushi/toml"
)

/*
type sectioner interface {
	// An interface for Packer template sections-when implemented.
	Load() error
	mergeSettings([]string)
	settingsToMap(IODirInf) map[string]interface{}
}
*/

type build struct {
	// Contains most of the information for Packer templates within a Rancher Build.
	BuilderType []string `toml:"builder_type"`

	Builders map[string]builder `toml:"builders"`

	PostProcessors map[string]postProcessors `toml:"post_processors"`

	Provisioners map[string]provisioners `toml:"provisioners"`
}

type builder struct {
	// Defines a representation of the builder section of a Packer template.
	Settings   []string `toml:"Settings"`
	VMSettings []string `toml:"vm_Settings"`
}

func (b *builder) mergeSettings(new []string) {
	// Merge the settings section of a builder. New values supercede existing ones.
	b.Settings = mergeSettingsSlices(b.Settings, new)
}

func (b *builder) mergeVMSettings(new []string) {
	// Merge the VMSettings section of a builder. New values supercede existing ones.
	b.VMSettings = mergeSettingsSlices(b.VMSettings, new)
}

func (b *builder) settingsToMap(r *RawTemplate) map[string]interface{} {
	// Go through all of the Settings and convert them to a map. If any additional
	// processing needs to be done to a builder setting, it is done here.
	var k, v string

	m := make(map[string]interface{}, len(b.Settings)+len(b.VMSettings))
	//	m := map[string]interface{}
	for _, s := range b.Settings {
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "http_Directory":
			//			v = resolvePathTemplate(io, v)
		}

		m[k] = v
	}

	return m
}

// Type for handling the post-processor section of the configs.
type postProcessors struct {
	Settings []string
}

func (p *postProcessors) mergeSettings(new []string) {
	// Merge the settings section of a post-processor. New values supercede existing ones.
	p.Settings = mergeSettingsSlices(p.Settings, new)
}

func (p *postProcessors) settingsToMap(Type string, r *RawTemplate) map[string]interface{} {
	// Go through all of the Settings and convert them to a map. If any additional
	// processing needs to be done to a post-processor setting, it is done here.
	var k, v string
	m := make(map[string]interface{}, len(p.Settings))

	m["type"] = Type

	for _, s := range p.Settings {
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "output", "vagrantfile_template":
			//			v = resolvePathTemplate(io, v)
		}

		m[k] = v
	}

	return m
}

// Type for handling the provisioners sections of the configs.
type provisioners struct {
	Settings []string `toml:"settings"`
	Scripts  []string `toml:"scripts"`
}

func (p *provisioners) mergeSettings(new []string) {
	// Merge the settings section of a post-processor. New values supercede existing ones.
	p.Settings = mergeSettingsSlices(p.Settings, new)
}

func (p *provisioners) settingsToMap(Type string, r *RawTemplate) map[string]interface{} {
	// Go through all of the Settings and convert them to a map. If any additional
	// processing needs to be done to a provisioners setting, it is done here.
	var k, v string
	m := make(map[string]interface{}, len(p.Settings))

	m["type"] = Type

	for _, s := range p.Settings {
		k, v = parseVar(s)
		v = r.replaceVariables(v)

		switch k {
		case "execute_command":
			if c, err := commandFromFile(v); err != nil {
				v = "Error: " + err.Error()
				err = nil
			} else {
				v = c[0]
			}
		case "environment_vars":
			// TODO--figure out what I was thinking with the above case and comment below--or delete this case
			// do same as scripts except no resolve template path
		}

		m[k] = v
	}

	return m
}

func (p *provisioners) setScripts(new []string) {
	// Scripts are only replaced if it has values, otherwise the existing values are used.
	if len(new) > 0 {
		p.Scripts = new
	}
}

type defaults struct {
	// Defaults is used to store Rancher application level defaults for Packer templates.
	IODirInf
	PackerInf
	BuildInf
	build
}

type BuildInf struct {
	Name      string `toml:"name"`
	BuildName string `toml:"build_name"`
}

type IODirInf struct {
	// IODirInf is used to store information about where Rancher can find and put things.
	CommandsDir string `toml:"commands_dir"`
	HTTPDir		string `toml:"http_dir"`
	OutDir      string `toml:"out_dir"`
	ScriptsDir  string `toml:"scripts_dir"`
	ScriptsSrcDir  string `toml:"scripts_src_dir"`
	SrcDir      string `toml:"src_dir"`
}

type PackerInf struct {
	// PackerInf is used to store information about a Packer Template. In Packer, these fields are optional.
	MinPackerVersion string `toml:"min_packer_Release" json:"min_packer_version"`
	Description      string `toml:"description" json:"description"`
}

// Load the defaults file.
func (d *defaults) Load() error {
	name := os.Getenv(EnvDefaultsFile)
	if name == "" {
		err := errors.New("could not retrieve the default Settings file because the " + EnvDefaultsFile + " ENV variable was not set. Either set it or check your rancher.cfg setting")
		return err
	}
	_, err := toml.DecodeFile(name, &d)

	return err
}

// To add support for a distribution, the information about it must be added to the supported. file, in addition to adding the code to support it to the application.
type Supported struct {
	Distro map[string]distro
}

type distro struct {
	// Struct to hold the details of supported distros. From this information a user should be able to
	// build a Packer template by only executing the following, at minimum:
	//
	//	$ rancher build -distro=ubuntu
	//
	// All settings can be overridden. The information here represents the standard box configuration for
	// its respective distribution.
	IODirInf
	PackerInf
	BuildInf
	BaseURL string `toml:"base_url"`

	// The supported Architectures, which can differ per distro. The labels can also differ, e.g. amd64 and x86_64.
	Arch []string `toml:"Arch"`

	// Supported iso Images, e.g. server, minimal, etc.
	Image []string `toml:"Image"`

	// Supported Releases: the supported Releases are the Releases available for download from that distribution's download page. Archived and unsupported Releases are not used.
	Release []string `toml:"Release"`

	// The default Image configuration for this distribution. This usually consists of things like Release, Architecture, Image type, etc.
	DefImage []string `toml:"default_Image"`

	// The configurations needed to generate the default settings for a build for this distribution.
	build
}

// Load the Supported Distros file
func (s *Supported) Load() error {
	name := os.Getenv(EnvSupportedFile)

	if name == "" {
		err := errors.New("could not retrieve the Supported information because the " + EnvSupportedFile + " Env variable was not set. Either set it or check your rancher.cfg setting")
		return err
	}
	_, err := toml.DecodeFile(name, &s)
	return err
}

// Struct to hold the builds.
type Builds struct {
	Build map[string]RawTemplate
}

func (b *Builds) Load() error {
	name := os.Getenv(EnvBuildsFile)
	if name == "" {
		err := errors.New("could not retrieve the Builds configurations because the " + EnvBuildsFile + "Env variable was not set. Either set it or check your rancher.cfg setting")
		return err
	}
	_, err := toml.DecodeFile(name, &b)

	return err
}

// Contains lists of builds.
type buildLists struct {
	List map[string]list
}

// A list of builds. Each list contains one or more builds.
type list struct {
	Builds []string
}

func (b *buildLists) Load() error {
	// Load the build lists.
	name := os.Getenv(EnvBuildListsFile)
	if name == "" {
		err := errors.New("could not retrieve the BuildLists file because the " + EnvBuildListsFile + " Env variable was not set. Either set it or check your rancher.cfg setting")
		return err
	}
	_, err := toml.DecodeFile(name, &b)

	return err
}
