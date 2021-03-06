package app

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"unsafe"

	"github.com/mohae/contour"
	"github.com/mohae/feedlot/conf"
)

var region = "US"
var country = "CA"
var sponsor = "OSUOSL"
var noSponsor = ""

var testDefaults = &Defaults{
	IODirInf: IODirInf{
		TemplateOutputDir: "packer_templates/:build_name",
		PackerOutputDir:   "packer_boxes/:build_name",
		SourceDir:         "src",
	},
	PackerInf: PackerInf{
		Description:      "Test Default Rancher template",
		MinPackerVersion: "0.4.0",
	},
	BuildInf: BuildInf{
		BaseURL:   "",
		BuildName: "",
		Name:      ":build_name",
	},
	Build: Build{
		BuilderIDs: []string{
			"virtualbox-iso",
		},
		Builders: map[string]BuilderC{
			"common": {
				TemplateSection{
					Type: "common",
					Settings: []string{
						"boot_command = boot_test.command",
						"boot_wait = 5s",
						"disk_size = 20000",
						"guest_os_type = ",
						"headless = true",
						"http_directory = http",
						"iso_checksum_type = sha256",
						"output_directory = :packer_output_dir",
						"shutdown_command = shutdown_test.command",
						"ssh_password = vagrant",
						"ssh_port = 22",
						"ssh_username = vagrant",
						"ssh_wait_timeout = 240m",
					},
				},
			},
			"virtualbox-iso": {
				TemplateSection{
					Type: "virtualbox-iso",
					Settings: []string{
						"guest_additions_path = VBoxGuestAdditions_{{ .Version }}.iso",
						"virtualbox_version_file = .vbox_version",
					},
					Arrays: map[string]interface{}{
						"vboxmanage": []interface{}{
							"cpus=1",
							"memory=1024",
						},
					},
				},
			},
		},
		PostProcessorIDs: []string{
			"vagrant",
		},
		PostProcessors: map[string]PostProcessorC{
			"vagrant": {
				TemplateSection{
					Type: "vagrant",
					Settings: []string{
						"compression_level = 9",
						"keep_input_artifact = false",
						"output = :build_name.box",
					},
				},
			},
		},
		ProvisionerIDs: []string{
			"shell",
		},
		Provisioners: map[string]ProvisionerC{
			"shell": {
				TemplateSection{
					Type: "shell",
					Settings: []string{
						"execute_command = execute_test.command",
					},
					Arrays: map[string]interface{}{
						"scripts": []interface{}{
							"setup_test.sh",
							"vagrant_test.sh",
							"sudoers_test.sh",
							"cleanup_test.sh",
						},
					},
				},
			},
		},
	},
	loaded: true,
}

var testSupported = map[string]SupportedDistro{
	"centos": SupportedDistro{
		BuildInf: BuildInf{
			BaseURL: "",
			Region:  &region,
			Country: &country,
		},
		IODirInf: IODirInf{},
		PackerInf: PackerInf{
			MinPackerVersion: "",
			Description:      "Default template config and Rancher options for CentOS",
		},
		Arch: []string{
			"i386",
			"x86_64",
		},
		Image: []string{
			"minimal",
			"netinstall",
		},
		Release: []string{
			"5",
			"6",
		},
		DefImage: []string{
			"release = 6",
			"image = minimal",
			"arch = x86_64",
		},
	},
	"debian": SupportedDistro{
		BuildInf: BuildInf{
			BaseURL: "http://cdimage.debian.org/debian-cd/",
		},
		IODirInf: IODirInf{},
		PackerInf: PackerInf{
			MinPackerVersion: "",
			Description:      "Default template config and Rancher options for Debian",
		},
		Arch: []string{
			"i386",
			"amd64",
		},
		Image: []string{
			"lxde-CD-1",
			"netinst",
			"xfce-CD-1",
		},
		Release: []string{
			"8",
		},
		DefImage: []string{
			"release = 8",
			"image = netinst",
			"arch = amd64",
		},
	},
}
var testSupportedUbuntu = &SupportedDistro{
	BuildInf: BuildInf{
		BaseURL: "http://releases.ubuntu.com/",
	},
	IODirInf: IODirInf{},
	PackerInf: PackerInf{
		MinPackerVersion: "",
		Description:      "Test supported distribution template",
	},
	Arch: []string{
		"i386",
		"amd64",
	},
	Image: []string{
		"server",
	},
	Release: []string{
		"10.04",
		"12.04",
		"12.10",
		"13.04",
		"13.10",
	},
	DefImage: []string{
		"release = 12.04",
		"image = server",
		"arch = amd64",
	},
	Build: Build{
		BuilderIDs: []string{
			"virtualbox-iso",
			"vmware-iso",
		},
		Builders: map[string]BuilderC{
			"common": {
				TemplateSection{
					Settings: []string{
						"boot_command = boot_test.command",
						"shutdown_command = shutdown_test.command",
					},
				},
			},
			"virtualbox-iso": {
				TemplateSection{
					Arrays: map[string]interface{}{
						"vm_settings": []string{"memory=2048"},
					},
				},
			},
			"vmware-iso": {
				TemplateSection{
					Arrays: map[string]interface{}{
						"vm_settings": []string{"memsize=2048"},
					},
				},
			},
		},
		PostProcessorIDs: []string{
			"vagrant",
		},
		PostProcessors: map[string]PostProcessorC{
			"vagrant": {
				TemplateSection{
					Settings: []string{
						"output = out/:build_name-packer.box",
					},
				},
			},
		},
		ProvisionerIDs: []string{
			"shell",
			"file-uploads",
		},
		Provisioners: map[string]ProvisionerC{
			"shell": {
				TemplateSection{
					Settings: []string{
						"execute_command = execute_test.command",
					},
					Arrays: map[string]interface{}{
						"scripts": []string{
							"setup_test.sh",
							"base_test.sh",
							"vagrant_test.sh",
							"sudoers_test.sh",
							"cleanup_test.sh",
						},
					},
				},
			},
			"file-uploads": {
				TemplateSection{
					Settings: []string{
						"source = source/dir",
						"destination = destination/dir",
					},
				},
			},
		},
	},
}

var testSupportedCentOS = &SupportedDistro{
	BuildInf: BuildInf{
		BaseURL: "",
		Region:  &region,
		Country: &country,
		Sponsor: &noSponsor,
	},
	IODirInf: IODirInf{},
	PackerInf: PackerInf{
		MinPackerVersion: "",
		Description:      "Test template config and Rancher options for CentOS",
	},
	Arch: []string{
		"i386",
		"x86_64",
	},
	Image: []string{
		"minimal",
		"netinstall",
	},
	Release: []string{
		"5",
		"6",
	},
	DefImage: []string{
		"release = 6",
		"image = minimal",
		"arch = x86_64",
	},
}

var testBuild = map[string]RawTemplate{
	"1204-amd64": RawTemplate{
		Distro: "ubuntu",
		PackerInf: PackerInf{
			Description: "ubuntu LTS 1204 amd64 server build, minimal install",
		},
		Arch:    "amd64",
		Image:   "server",
		Release: "12.04",
		Build: Build{
			BuilderIDs: []string{
				"virtualbox-iso",
			},
			Builders: map[string]BuilderC{
				"common": {
					TemplateSection{
						Type: "common",
						Settings: []string{
							"ssh_wait_timeout = 300m",
						},
					},
				},
				"virtualbox-iso": {
					TemplateSection{
						Type: "virtualbox-iso",
						Arrays: map[string]interface{}{
							"vboxmanage": []string{
								"memory=4096",
							},
						},
					},
				},
			},
		},
	},
	"centos6": RawTemplate{
		Distro: "centos",
		PackerInf: PackerInf{
			Description: "Centos 6 w virtualbox-iso only",
		},
		Build: Build{
			BuilderIDs: []string{
				"virtualbox-iso",
			},
		},
	},
	"jessie": RawTemplate{
		Distro: "debian",
		PackerInf: PackerInf{
			Description: "debian jessie",
		},
		Arch: "amd64",
		Build: Build{
			BuilderIDs: []string{
				"virtualbox-iso",
			},
			Builders: map[string]BuilderC{
				"virtualbox-iso": {
					TemplateSection{
						Type: "virtualbox-iso",
						Arrays: map[string]interface{}{
							"vboxmanage": []string{
								"--memory=4096",
							},
						},
					},
				},
			},
			PostProcessorIDs: []string{
				"vagrant",
			},
			PostProcessors: map[string]PostProcessorC{
				"vagrant": {
					TemplateSection{
						Type: "vagrant",
						Settings: []string{
							"output = out/:build_name-packer.box",
						},
					},
				},
			},
			ProvisionerIDs: []string{
				"basic-shell",
			},
			Provisioners: map[string]ProvisionerC{
				"basic-shell": {
					TemplateSection{
						Type: "shell",
						Arrays: map[string]interface{}{
							"scripts": []string{
								"setup.sh",
								"sudoers.sh",
								"vagrant.sh",
								"customize.sh",
								"cleanup.sh",
							},
						},
					},
				},
			},
		},
	},
}

var testBuildList = map[string]List{
	"ubuntu-all": List{Builds: []string{"1204-amd64-server", "1310-amd64-desktop"}},
}

func TestBuildCopy(t *testing.T) {
	tstTpl := testBuild["jessie"]
	newBuild := tstTpl.Build.Copy()
	msg, ok := EvalStringSlice(newBuild.BuilderIDs, tstTpl.BuilderIDs)
	if !ok {
		t.Errorf("BuilderIDs: %s", msg)
	}

	if (*reflect.SliceHeader)(unsafe.Pointer(&newBuild.Builders)).Data == (*reflect.SliceHeader)(unsafe.Pointer(&tstTpl.Build.Builders)).Data {
		t.Errorf("The pointer for Builders is the same for both newBuild and testBuild: %x, expected them to be different.", (*reflect.SliceHeader)(unsafe.Pointer(&newBuild.Builders)).Data)
		goto buildersEnd
	}
	if len(newBuild.Builders) != len(tstTpl.Builders) {
		t.Errorf("Expected newBuild.Builders to have a length of %d; got %d", len(tstTpl.Builders), len(newBuild.Builders))
		goto buildersEnd
	}
	for k := range tstTpl.Builders {
		_, ok := newBuild.Builders[k]
		if !ok {
			t.Errorf("Expected %s to be a builder in the copy, but it wasn't", k)
		}
	}

buildersEnd:
	msg, ok = EvalStringSlice(newBuild.PostProcessorIDs, tstTpl.PostProcessorIDs)
	if !ok {
		t.Errorf("PostProcessorIDs: %s", msg)
	}

	if (*reflect.SliceHeader)(unsafe.Pointer(&newBuild.PostProcessors)).Data == (*reflect.SliceHeader)(unsafe.Pointer(&tstTpl.Build.PostProcessors)).Data {
		t.Errorf("The pointer for PostProcessors is the same for both newBuild and testBuild: %x, expected them to be different.", (*reflect.SliceHeader)(unsafe.Pointer(&newBuild.PostProcessors)).Data)
		goto postProcessorsEnd
	}
	if len(newBuild.PostProcessors) != len(tstTpl.PostProcessors) {
		t.Errorf("Expected newBuild.PostProcessors to have a length of %d; got %d", len(tstTpl.PostProcessors), len(newBuild.PostProcessors))
		goto postProcessorsEnd
	}
	for k := range tstTpl.PostProcessors {
		_, ok := newBuild.PostProcessors[k]
		if !ok {
			t.Errorf("Expected %s to be a PostProcessor in the copy, but it wasn't", k)
		}
	}

postProcessorsEnd:
	msg, ok = EvalStringSlice(newBuild.ProvisionerIDs, tstTpl.ProvisionerIDs)
	if !ok {
		t.Errorf("ProvisionerIDs: %s", msg)
	}

	if (*reflect.SliceHeader)(unsafe.Pointer(&newBuild.Provisioners)).Data == (*reflect.SliceHeader)(unsafe.Pointer(&tstTpl.Build.Provisioners)).Data {
		t.Errorf("The pointer for Provisioners is the same for both newBuild and testBuild: %x, expected them to be different.", (*reflect.SliceHeader)(unsafe.Pointer(&newBuild.Provisioners)))
	}
	if len(newBuild.Provisioners) != len(tstTpl.Provisioners) {
		t.Errorf("Expected newBuild.Provisioners to have a length of %d; got %d", len(tstTpl.Provisioners), len(newBuild.Provisioners))
		goto provisionersEnd
	}
	for k := range tstTpl.Provisioners {
		_, ok := newBuild.Provisioners[k]
		if !ok {
			t.Errorf("Expected %s to be a Provisioners in the copy, but it wasn't", k)
		}
	}
provisionersEnd:
}

func TestTemplateSectionMergeArrays(t *testing.T) {
	ts := &TemplateSection{}
	ts.mergeArrays(nil)
	if ts.Arrays != nil {
		t.Errorf("Expected the merged array to be nil, was not nil: %#v", ts.Arrays)
	}

	old := map[string]interface{}{
		"type":            "shell",
		"execute_command": "echo 'vagrant'|sudo -S sh '{{.Path}}'",
		"override": map[string]interface{}{
			"virtualbox-iso": map[string]interface{}{
				"scripts": []string{
					"base.sh",
					"vagrant.sh",
					"vmware.sh",
					"cleanup.sh",
				},
			},
		},
	}

	nw := map[string]interface{}{
		"type": "shell",
		"override": map[string]interface{}{
			"vmware-iso": map[string]interface{}{
				"scripts": []string{
					"base.sh",
					"vagrant.sh",
					"vmware.sh",
					"cleanup.sh",
				},
			},
		},
	}

	merged := map[string]interface{}{
		"type":            "shell",
		"execute_command": "echo 'vagrant'|sudo -S sh '{{.Path}}'",
		"override": map[string]interface{}{
			"vmware-iso": map[string]interface{}{
				"scripts": []string{
					"base.sh",
					"vagrant.sh",
					"vmware.sh",
					"cleanup.sh",
				},
			},
		},
	}

	ts.Arrays = old
	ts.mergeArrays(nil)
	if ts.Arrays == nil {
		t.Errorf("Expected merged to be not nil, was nil")
	} else {
		if !reflect.DeepEqual(ts.Arrays, old) {
			t.Errorf("Expected %#v, got %#v", old, ts.Arrays)
		}
	}

	ts.Arrays = nil
	ts.mergeArrays(nw)
	if ts.Arrays == nil {
		t.Errorf("Expected merged to be not nil, was nil")
	} else {
		if !reflect.DeepEqual(ts.Arrays, nw) {
			t.Errorf("Expected %#v, got %#v", nw, ts.Arrays)
		}
	}

	ts.Arrays = old
	ts.mergeArrays(nw)
	if ts.Arrays == nil {
		t.Errorf("Expected merged to be not nil, was nil")
	} else {
		if !reflect.DeepEqual(ts.Arrays, merged) {
			t.Errorf("Expected %#v, got %#v", merged, ts.Arrays)
		}
	}
}

func init() {
	var b bool
	b = true
	testDefaults.IncludeComponentString = &b
	testDefaults.TemplateOutputDirIsRelative = &b
	testDefaults.SourceDirIsRelative = &b
}

/*
func TestBuilderMergeSettings(t *testing.T) {
	b := builder{}
	key1 := "key1=value1"
	key2 := "key2=value2"
	key3 := "key3=value3"

	b.Settings = []string{key1, key2, key3}
	b.mergeSettings(nil)
	if !stringSliceContains(b.Settings, key1) {
		t.Errorf("expected %s in slice: not found", key1)
	}
	if !stringSliceContains(b.Settings, key2) {
		t.Errorf("expected %s in slice: not found", key2)
	}
	if !stringSliceContains(b.Settings, key3) {
		t.Errorf("expected %s in slice: not found", key3)
	}

	key4 := "key4=value4"
	key2update := "key2=value22"
	newSettings := []string{key4, key2update}
	b.mergeSettings(newSettings)
	if !stringSliceContains(b.Settings, key1) {
		t.Errorf("expected %s in slice: not found", key1)
	}
	if !stringSliceContains(b.Settings, key2update) {
		t.Errorf("expected %s in slice: not found", key2update)
	}
	if !stringSliceContains(b.Settings, key3) {
		t.Errorf("expected %s in slice: not found", key3)
	}
	if !stringSliceContains(b.Settings, key3) {
		t.Errorf("expected %s in slice: not found", key4)
	}
	if stringSliceContains(b.Settings, key2) {
		t.Errorf("did not expect %s in slice: was found", key2)
	}
}

func TestPostProcessorMergeSettings(t *testing.T) {
	pp := postProcessor{}
	pp.Settings = []string{"key1=value1", "key2=value2"}
	pp.mergeSettings(nil)
	if !stringSliceContains(pp.Settings, "key1=value1") {
		t.Errorf("expected %s in slice: not found", "key1=value1")
	}
	if !stringSliceContains(pp.Settings, "key2=value2") {
		t.Errorf("expected %s in slice: not found", "key2=value2")
	}

	newSettings := []string{"key1=value1", "key2=value22", "key3=value3"}
	pp.mergeSettings(newSettings)
	if !stringSliceContains(pp.Settings, "key1=value1") {
		t.Errorf("expected %s in slice: not found", "key1=value1")
	}
	if !stringSliceContains(pp.Settings, "key2=value22") {
		t.Errorf("expected %s in slice: not found", "key2=value22")
	}
	if !stringSliceContains(pp.Settings, "key3=value3") {
		t.Errorf("expected %s in slice: not found", "key3=value3")
	}
	if stringSliceContains(pp.Settings, "key2=value2") {
		t.Errorf("expected %s in slice: not found", "key2=value2")
	}

	post := postProcessor{}
	post.mergeSettings(newSettings)
	if !stringSliceContains(pp.Settings, "key1=value1") {
		t.Errorf("expected %s in slice: not found", "key1=value1")
	}
	if !stringSliceContains(pp.Settings, "key2=value22") {
		t.Errorf("expected %s in slice: not found", "key2=value22")
	}
	if !stringSliceContains(pp.Settings, "key3=value3") {
		t.Errorf("expected %s in slice: not found", "key3=value3")
	}
}

func TestProvisionerMergeSettings(t *testing.T) {
	p := provisioner{}
	p.Settings = []string{"key1=value1", "key2=value2"}
	p.mergeSettings(nil)
	if !stringSliceContains(p.Settings, "key1=value1") {
		t.Errorf("expected %s in slice: not found", "key1=value1")
	}
	if !stringSliceContains(p.Settings, "key2=value2") {
		t.Errorf("expected %s in slice: not found", "key2=value2")
	}

	newSettings := []string{"key1=value1", "key2=value22", "key3=value3"}
	p.mergeSettings(newSettings)
	if !stringSliceContains(p.Settings, "key1=value1") {
		t.Errorf("expected %s in slice: not found", "key1=value1")
	}
	if !stringSliceContains(p.Settings, "key2=value22") {
		t.Errorf("expected %s in slice: not found", "key2=value22")
	}
	if !stringSliceContains(p.Settings, "key3=value3") {
		t.Errorf("expected %s in slice: not found", "key3=value3")
	}
	if stringSliceContains(p.Settings, "key2=value2") {
		t.Errorf("expected %s in slice: not found", "key2=value2")
	}

	pr := provisioner{}
	pr.mergeSettings(newSettings)
	if !stringSliceContains(pr.Settings, "key1=value1") {
		t.Errorf("expected %s in slice: not found", "key1=value1")
	}
	if !stringSliceContains(pr.Settings, "key2=value22") {
		t.Errorf("expected %s in slice: not found", "key2=value22")
	}
	if !stringSliceContains(pr.Settings, "key3=value3") {
		t.Errorf("expected %s in slice: not found", "key3=value3")
	}
}
*/
func TestDefaults(t *testing.T) {
	tests := []struct {
		format      string
		expectedErr string
	}{
		{"", "load defaults: : unsupported conf format"},
		{"yaml", "load defaults: yaml: unsupported conf format"},
		{"toml", ""},
		{"json", ""},
	}

	contour.UpdateString(conf.Dir, "../test_files/conf")
	for i, test := range tests {
		contour.UpdateString(conf.Format, test.format)
		d := Defaults{}
		err := d.Load("")
		if err != nil {
			if err.Error() != test.expectedErr {
				t.Errorf("%d: expected %q, got %q", i, test.expectedErr, err)
			}
			continue
		}
		if test.expectedErr != "" {
			t.Errorf("%d: expepcted an error: %q, got none", i, test.expectedErr)
			continue
		}
		got, _ := json.MarshalIndent(d, "", "\t")
		want, _ := json.MarshalIndent(testDefaults, "", "\t")
		if string(got) != string(want) {
			t.Errorf("%d: expected %#v, got %#v", i, string(want), string(got))
		}
	}
}

func TestSupportedDistros(t *testing.T) {
	tests := []struct {
		format      string
		p           string
		expectedErr string
	}{
		{"", "", "load supported: : unsupported conf format"},
		{"yaml", "", "load supported: yaml: unsupported conf format"},
		{"toml", "../test_files", ""},
		{"json", "../test_files", ""},
	}
	for i, test := range tests {
		contour.UpdateString(conf.Format, test.format)
		s := SupportedDistros{}
		err := s.Load(test.p)
		if err != nil {
			if err.Error() != test.expectedErr {
				t.Errorf("%d: expected %q, got %q", i, test.expectedErr, err)
			}
			continue
		}
		if test.expectedErr != "" {
			t.Errorf("%d: expected an error: %q, got none", i, test.expectedErr)
			continue
		}
		got, _ := json.MarshalIndent(s.Distros, "", "\t")
		want, _ := json.MarshalIndent(testSupported, "", "\t")
		if string(got) != string(want) {
			t.Errorf("%d: expected %s, got %s", i, string(want), string(got))
		}
	}
}

func TestBuildStuff(t *testing.T) {
	tests := []struct {
		filename    string
		format      string
		expectedErr string
	}{
		{"", "", "load build: no build name specified"},
		{"", "yaml", "load build: no build name specified"},
		{"", "toml", "load build: no build name specified"},
		{"", "json", "load build: no build name specified"},
		{"../test_files/conf/build2.yaml", "yaml", "load build ../test_files/conf/build2.yaml: unsupported format"},
		{"../test_files/conf/build2.toml", "toml", ""},
		{"../test_files/conf/build2.json", "json", ""},
	}
	contour.UpdateString(conf.Dir, "../test_files/conf")
	for i, test := range tests {
		contour.UpdateString(conf.Format, test.format)
		b := Builds{}
		err := b.Load(test.filename)
		if err != nil {
			if err.Error() != test.expectedErr {
				t.Errorf("%d: expected %q, got %q", i, test.expectedErr, err)
			}
			continue
		}
		if test.expectedErr != "" {
			t.Errorf("%d: expepcted an error: %q, got none", i, test.expectedErr)
			continue
		}
		got, _ := json.MarshalIndent(b.Templates, "", "\t")
		want, _ := json.MarshalIndent(testBuild, "", "\t")
		if string(got) != string(want) {
			t.Errorf("%d: expected %s, got %s", i, string(want), string(got))
		}
	}
}

func TestBuildListStuff(t *testing.T) {
	tests := []struct {
		format      string
		expectedErr string
	}{
		{"", "load build list: : : unsupported conf format"},
		{"yaml", "load build list: : yaml: unsupported conf format"},
		{"toml", ""},
		{"json", ""},
	}
	contour.UpdateString(conf.Dir, "conf")
	for i, test := range tests {
		contour.UpdateString(conf.Format, test.format)
		b := &BuildLists{Lists: map[string]List{}}
		err := b.Load("../test_files")
		if err != nil {
			if err.Error() != test.expectedErr {
				t.Errorf("%d: expected %q, got %q", i, test.expectedErr, err)
			}
			continue
		}
		if test.expectedErr != "" {
			t.Errorf("%d: expected an error: %q, got none", i, test.expectedErr)
			continue
		}
		if !reflect.DeepEqual(b.Lists, testBuildList) {
			t.Errorf("%d: expected %#v, got %#v", i, testBuildList, b.Lists)
		}
	}
}

func TestIODirInfUpdate(t *testing.T) {
	oldIODirInf := IODirInf{TemplateOutputDir: "old TemplateOutputDir", PackerOutputDir: "old PackerOutputDir", SourceDir: "old SrcDir"}
	newIODirInf := IODirInf{}
	oldIODirInf.update(newIODirInf)
	if oldIODirInf.TemplateOutputDir != "old TemplateOutputDir/" {
		t.Errorf("Expected \"old TemplateOutputDir/\", got %q", oldIODirInf.TemplateOutputDir)
	}
	if oldIODirInf.PackerOutputDir != "old PackerOutputDir/" {
		t.Errorf("Expected \"old PackerOutputDir/\", got %q", oldIODirInf.PackerOutputDir)
	}
	if oldIODirInf.SourceDir != "old SrcDir/" {
		t.Errorf("Expected \"old SrcDir/\", got %q", oldIODirInf.SourceDir)
	}

	oldIODirInf = IODirInf{TemplateOutputDir: "old TemplateOutputDir", PackerOutputDir: "old PackerOutputDir", SourceDir: "old SrcDir"}
	newIODirInf = IODirInf{TemplateOutputDir: "new TemplateOutputDir", PackerOutputDir: "new PackerOutputDir", SourceDir: "new SrcDir"}
	oldIODirInf.update(newIODirInf)
	if oldIODirInf.TemplateOutputDir != "new TemplateOutputDir/" {
		t.Errorf("Expected \"new TemplateOutputDir/\", got %q", oldIODirInf.TemplateOutputDir)
	}
	if oldIODirInf.PackerOutputDir != "new PackerOutputDir/" {
		t.Errorf("Expected \"new PackerOutputDir/\", got %q", oldIODirInf.PackerOutputDir)
	}
	if oldIODirInf.SourceDir != "new SrcDir/" {
		t.Errorf("Expected \"new SrcDir/\", got %q", oldIODirInf.SourceDir)
	}

	oldIODirInf = IODirInf{TemplateOutputDir: "old TemplateOutputDir", PackerOutputDir: "old PackerOutputDir", SourceDir: "old SrcDir"}
	newIODirInf = IODirInf{TemplateOutputDir: "TemplateOutputDir"}
	oldIODirInf.update(newIODirInf)
	if oldIODirInf.TemplateOutputDir != "TemplateOutputDir/" {
		t.Errorf("Expected \"TemplateOutputDir/\", got %q", oldIODirInf.TemplateOutputDir)
	}
	if oldIODirInf.PackerOutputDir != "old PackerOutputDir/" {
		t.Errorf("Expected \"old PackerOutputDir/\", got %q", oldIODirInf.PackerOutputDir)
	}
	if oldIODirInf.SourceDir != "old SrcDir/" {
		t.Errorf("Expected \"old SrcDir/\", got %q", oldIODirInf.SourceDir)
	}
}

func TestFindConfigFile(t *testing.T) {
	tests := []struct {
		fName          string
		findName       string
		cfgFormat      string
		expectedName   string
		expectedFormat conf.ConfFormat
		expectedErr    string
	}{
		{"test.json", "test.json", "json", "test.cjsn", conf.JSON, ""},
		{"test.json", "test.json", "json", "test.cjon", conf.JSON, ""},
		{"test.cjson", "test.cjson", "json", "test.json", conf.JSON, ""},
		{"test.tml", "test.tml", "toml", "test.toml", conf.TOML, ""},
		{"atest.toml", "test.toml", "toml", "", conf.TOML, "stat test.toml: no such file or directory"},
		{"test.yaml", "test.yaml", "yaml", "", conf.UnsupportedConfFormat, ""},
	}

	tmpDir, err := ioutil.TempDir("", "feedlotCfgTest")
	if err != nil {
		t.Errorf("unable to create tmp dir; testing of FindconfigFile failed: %s", err)
		return
	}
	b := []byte("this is a test")
	contour.RegisterString(conf.Format, "")
	for i, test := range tests {
		// create file
		err := ioutil.WriteFile(filepath.Join(tmpDir, test.fName), b, 0777)
		if err != nil {
			t.Errorf("an error occurred while creating temp file %s: %s", test.fName, err)
			continue
		}
		contour.UpdateString(conf.Format, test.cfgFormat)
		name, format, err := conf.ConfFilename(filepath.Join(tmpDir, test.findName))
		if err != nil {
			if err.Error() != test.expectedErr {
				t.Errorf("%d: expected error to be %q got %q", i, test.expectedErr, err)
				continue
			}
			if test.expectedFormat != format {
				t.Errorf("%d: expected CfgFormat to be %s, got %s", i, test.expectedFormat, format)
			}
			if filepath.Join(tmpDir, test.expectedName) != name {
				t.Errorf("%d: expected filename to be %s, got %s", i, filepath.Join(tmpDir, test.expectedName), name)
			}
		}

	}
	os.RemoveAll(tmpDir)

}
