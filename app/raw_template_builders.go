package app

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mohae/utilitybelt/deepcopy"
)

// Builder constants
const (
	UnsupportedBuilder Builder = iota
	Common
	Custom
	AmazonEBS
	DigitalOcean
	Docker
	GoogleCompute
	Null
	Openstack
	Parallels
	QEMU
	VirtualBoxISO
	VirtualBoxOVF
	VMWareISO
	VMWareVMX
)

// Builder is a Packer supported builder.
type Builder int

var builders = [...]string{
	"unsupported Builder",
	"common",         // Common is the name of the common builder
	"custom",         // custom is the name of the custom builder
	"amazon-ebs",     // AmazonEBS is the name of the amazon-ebs backed builder
	"digitalocean",   // DigitalOcean is the name of the Digital Ocean builder
	"docker",         // Docker is the name of the Docker builder
	"googlecompute",  // GoogleComputeEngine is the name of the Google Compute Engine builder
	"null",           // Null is the name of the Null
	"openstack",      // Openstack is the name of the Openstack
	"parallels",      // Parallels is the name of the Parallels builder
	"qemu",           // QEMU is the name of the QEMU builder
	"virtualbox-iso", // VirtualBoxISO is the name of the VirtualBox ISO builder
	"virtualbox-ovf", // VirtualBoxOVF is the name of the VirtualBox OVF builder
	"vmware-iso",     // VMWareISO is the name of the VMWare ISO builder
	"vmware-vmx",     // VMWareVMX is the name of the VMWare VMX builder
}

func (b Builder) String() string { return builders[b] }

// BuilderFromString returns the builder constant for the passed string or
// unsupported. All incoming strings are normalized to lowercase.
func BuilderFromString(s string) Builder {
	s = strings.ToLower(s)
	switch s {
	case "common":
		return Common
	case "custom":
		return Custom
	case "amazon-ebs":
		return AmazonEBS
	case "digitalocean":
		return DigitalOcean
	case "docker":
		return Docker
	case "googlecompute":
		return GoogleCompute
	case "null":
		return Null
	case "openstack":
		return Openstack
	case "parallels":
		return Parallels
	case "qemu":
		return QEMU
	case "virtualbox-iso":
		return VirtualBoxISO
	case "virtualbox-ovf":
		return VirtualBoxOVF
	case "vmware-iso":
		return VMWareISO
	case "vmware-vmx":
		return VMWareVMX
	}
	return UnsupportedBuilder
}

// r.createBuilders takes a raw builder and create the appropriate Packer
// Builder
func (r *rawTemplate) createBuilders() (bldrs []interface{}, err error) {
	if r.BuilderTypes == nil || len(r.BuilderTypes) <= 0 {
		err = fmt.Errorf("unable to create builders: none specified")
		return nil, err
	}
	var tmpS map[string]interface{}
	var ndx int
	bldrs = make([]interface{}, len(r.BuilderTypes))
	// Set the CommonBuilder settings. Only the builder.Settings field is used
	// for CommonBuilder as everything else is usually builder specific, even
	// if they have common names, e.g. difference between specifying memory
	// between VMWare and VirtualBox.
	//	r.updateCommonBuilder
	//
	// Generate the builders for each builder type.
	for _, bType := range r.BuilderTypes {
		typ := BuilderFromString(bType)
		switch typ {
		case AmazonEBS:
			tmpS, err = r.createAmazonEBS()
			if err != nil {
				return nil, err
			}
		// AmazonInstance, AmazonChroot:
		// not implemented
		case DigitalOcean:
			tmpS, err = r.createDigitalOcean()
			if err != nil {
				return nil, err
			}
		case Docker:
			tmpS, err = r.createDocker()
			if err != nil {
				return nil, err
			}
		case GoogleCompute:
			tmpS, err = r.createGoogleCompute()
			if err != nil {
				return nil, err
			}
		case Null:
			tmpS, err = r.createNull()
			if err != nil {
				return nil, err
			}
		//	case Openstack:
		//	case ParallelsISO, ParallelsPVM:
		//	case QEMU:
		case VMWareISO:
			tmpS, err = r.createVMWareISO()
			if err != nil {
				return nil, err
			}
		case VMWareVMX:
			tmpS, err = r.createVMWareVMX()
			if err != nil {
				return nil, err
			}
		case VirtualBoxISO:
			tmpS, err = r.createVirtualBoxISO()
			if err != nil {
				return nil, err
			}
		case VirtualBoxOVF:
			tmpS, err = r.createVirtualBoxOVF()
			if err != nil {
				return nil, err
			}
		default:
			err = fmt.Errorf("Builder, %q, is not supported by Rancher", bType)
			return nil, err
		}
		bldrs[ndx] = tmpS
		ndx++
	}
	return bldrs, nil
}

// Go through all of the Settings and convert them to a map.  Each setting is
// parsed into its constituent parts.  The value then goes through variable
// replacement to ensure that the settings are properly resolved.
func (b *builder) settingsToMap(r *rawTemplate) map[string]interface{} {
	var k, v string
	m := make(map[string]interface{})
	for _, s := range b.Settings {
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		m[k] = v
	}
	return m
}

// createAmazonEBS creates a map of settings for Packer's amazon-ebs builder.
// Any values that aren't supported by the amazon-ebs builder are ignored.  Any
// required settings that don't exist result in an error and processing of the
// builder is stopped.  For more information, refer to
// https://packer.io/docs/builders/amazon-ebs.html
//
// Required configuration options:
//   access_key                   string
//   ami_name                     string
//   instance_type                string
//   region                       string
//   secret_key                   string
//   source_ami                   string
//   ssh_username                 string
// Optional configuration options:
//   ami_description
//   ami_groups                    array of strings
//   ami_product_codes             array of strings
//   ami_regions                   array of strings
//   ami_users                     array of strings
//   associate_public_ip_address   boolean
//   availability_zone             string
//   enhanced_networking           string
//   iam_instance_profile          string
//   security_group_id             string
//   security_group_ids            array of strings
//   spot_price                    string
//   spot_price_auto_product       string
//   shh_port
//   ssh_private_key_file          string
//   ssh_private_ip                bool
//   ssh_timeout                   string
//   subnet_id                     string
//   temporary_key_pair_name       string
//   token                         string
//   user_data                     string
//   user_data_file                string
//   vpc_id                        string
// Not implemented configuration options:
//   ami_block_device_mappings     array of block device mappings
//   launch_block_device_mappings  array of block device mappings
//   run_tags                      object of key/value strings
//   tags                          object of key/value strings
func (r *rawTemplate) createAmazonEBS() (settings map[string]interface{}, err error) {
	_, ok := r.Builders[AmazonEBS.String()]
	if !ok {
		err = fmt.Errorf("no configuration found for %q", AmazonEBS.String())
		return nil, err
	}
	settings = make(map[string]interface{})
	// Each create function is responsible for setting its own type.
	settings["type"] = AmazonEBS.String()
	// Merge the settings between common and this builders.
	var workSlice []string
	_, ok = r.Builders[Common.String()]
	if ok {
		workSlice, err = mergeSettingsSlices(r.Builders[Common.String()].Settings, r.Builders[AmazonEBS.String()].Settings)
		if err != nil {
			return nil, fmt.Errorf("createAmazonEBS failed: %s", err.Error())
		}

	} else {
		workSlice = r.Builders[AmazonEBS.String()].Settings
	}
	var k, v string
	var hasAccessKey, hasAmiName, hasInstanceType, hasRegion, hasSecretKey, hasSourceAmi, hasSSHUsername bool
	// Go through each element in the slice, only take the ones that matter
	// to this builder.
	for _, s := range workSlice {
		// var tmp interface{}
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "access_key":
			settings[k] = v
			hasAccessKey = true
		case "ami_name":
			settings[k] = v
			hasAmiName = true
		case "instance_type":
			settings[k] = v
			hasInstanceType = true
		case "region":
			settings[k] = v
			hasRegion = true
		case "secret_key":
			settings[k] = v
			hasSecretKey = true
		case "source_ami":
			settings[k] = v
			hasSourceAmi = true
		case "ssh_username":
			settings[k] = v
			hasSSHUsername = true
		case "ami_description", "availability_zone", "iam_instance_profile",
			"security_group_id", "spot_price", "spot_price_auto_product",
			"ssh_private_key_file", "ssh_timeout", "subnet_id",
			"temporary_key_pair_name", "token", "user_data",
			"vpc_id":
			settings[k] = v
		case "ssh_port":
			// only add if its an int
			i, err := strconv.Atoi(v)
			if err != nil {
				err = fmt.Errorf("amazon-ebs builder error while trying to set %q to %q: %s", k, v, err)
				return nil, err
			}
			settings[k] = i
		case "user_data_file":
			src, err := r.findComponentSource(AmazonEBS.String(), v)
			if err != nil {
				return nil, err
			}
			r.files[r.buildOutPath(AmazonEBS.String(), v)] = src
			settings[k] = r.buildTemplateResourcePath(AmazonEBS.String(), v)
		case "associate_public_ip_address", "enhanced_networking", "ssh_private_ip":
			settings[k], _ = strconv.ParseBool(v)
		}
	}
	if !hasAccessKey {
		err := fmt.Errorf("\"access_key\" setting is required for amazon-ebs, not found")
		return nil, err
	}
	if !hasAmiName {
		err := fmt.Errorf("\"ami_name\" setting is required for amazon-ebs, not found")
		return nil, err
	}
	if !hasInstanceType {
		err := fmt.Errorf("\"instance_type\" setting is required for amazon-ebs, not found")
		return nil, err
	}
	if !hasRegion {
		err := fmt.Errorf("\"region\" setting is required for amazon-ebs, not found")
		return nil, err
	}
	if !hasSecretKey {
		err := fmt.Errorf("\"secret_key\" setting is required for amazon-ebs, not found")
		return nil, err
	}
	if !hasSourceAmi {
		err := fmt.Errorf("\"source_ami\" setting is required for amazon-ebs, not found")
		return nil, err
	}
	if !hasSSHUsername {
		err := fmt.Errorf("\"ssh_username\" setting is required for amazon-ebs, not found")
		return nil, err
	}
	// Process the Arrays.
	for name, val := range r.Builders[AmazonEBS.String()].Arrays {
		// if it's not a supported array group, log a warning and move on
		if name == "ami_groups" || name == "ami_product_codes" || name == "ami_regions" || name == "security_group_ids" {
			array := deepcopy.Iface(val)
			if array != nil {
				settings[name] = array
			}
			continue
		}
	}
	return settings, nil
}

// createDigitalOcean creates a map of settings for Packer's digitalocean
// builder.  Any values that aren't supported by the digitalocean builder are
// ignored.  Any required settings that don't exist result in an error and
// processing of the builder is stopped.  For more information, refer to
// https://packer.io/docs/builders/digitalocean.html
//
// NOTE: The deprecated image_id, region_id, and size_id options are not
//       supported.
//
// Required V1 api configuration options:
//   api_key             string
//   client_id           string
// Required V2 api configuration options:
//   api_token           string
// Optional configuration options:
//   api_url             string
//   droplet_name        string
//   image               string
//   image_id            integer
//   private_networking  boolean
//   region              string
//   region_id           integer
//   size                string
//   size_id             integer
//   snapshot_name       string
//   ssh_port            integer
//   ssh_timeout         string
//   ssh_username        string
//   state_timeout       string
func (r *rawTemplate) createDigitalOcean() (settings map[string]interface{}, err error) {
	_, ok := r.Builders[DigitalOcean.String()]
	if !ok {
		err = fmt.Errorf("no configuration found for %q", DigitalOcean.String())
		return nil, err
	}
	settings = make(map[string]interface{})
	// Each create function is responsible for setting its own type.
	settings["type"] = DigitalOcean.String()
	// If a common builder was defined, merge the settings between common and this builders.
	_, ok = r.Builders[Common.String()]
	var workSlice []string
	if ok {
		workSlice, err = mergeSettingsSlices(r.Builders[Common.String()].Settings, r.Builders[DigitalOcean.String()].Settings)
		if err != nil {
			return nil, fmt.Errorf("createDigitalOcean failed: %s", err.Error())
		}
	} else {
		workSlice = r.Builders[DigitalOcean.String()].Settings
	}
	// Go through each element in the slice, only take the ones that matter
	// to this builder.
	// TODO look at snapshot name handling--it should be unique, e.g. timestamp
	var hasApiToken, hasApiKey, hasClientID bool
	for _, s := range workSlice {
		// var tmp interface{}
		k, v := parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "api_key":
			settings[k] = v
			hasApiKey = true
		case "api_token":
			settings[k] = v
			hasApiToken = true
		case "client_id":
			settings[k] = v
			hasClientID = true
		case "api_url", "droplet_name", "image", "region", "size", "snapshot_name", "ssh_timeout", "ssh_username", "state_timeout":
			settings[k] = v
		case "private_networking":
			settings[k], _ = strconv.ParseBool(v)
		case "ssh_port":
			i, err := strconv.Atoi(v)
			if err != nil {
				err = fmt.Errorf("An error occurred while trying to set %s to %s: %s", k, v, err)
				return nil, err
			}
			settings[k] = i
		}
	}
	if hasApiToken {
		return settings, nil
	}
	if hasApiKey && hasClientID {
		return settings, nil
	}
	err = fmt.Errorf("required Digital Ocean API information not set")
	return nil, err
}

// createDocker creates a map of settings for Packer's docker builder. Any
// values that aren't supported by the digitalocean builder are ignored.  Any
// required settings that don't exist result in an error and processing of the
// builder is stopped. For more information, refer to
// https://packer.io/docs/builders/docker.html
//
// Required configuration options:
//   commit         boolean
//   export_path    string
//   image          string
// Optional configuration options:
//   login          boolean
//   login_email    string
//   login_username  string
//   login_password  string
//   login_server    string
//   pull            boolean
//   run_command     array of strings
// Not implemented configuration options:
//   volumes         map of strings to strings
//
// Note: the run+command can either be specified in the settings section or as
// an array of commands in the arrays section.  If it is in the settings
// section, it is expected to be a command file.  The run_command can only
// appear in one section.  Any run_commands specified in the arrays section
// will take precedence; if run_command is also in the settings section, it
// will be ignored.
func (r *rawTemplate) createDocker() (settings map[string]interface{}, err error) {
	_, ok := r.Builders[Docker.String()]
	if !ok {
		err = fmt.Errorf("no configuration found for %q", Docker.String())
		return nil, err
	}
	settings = make(map[string]interface{})
	// Each create function is responsible for setting its own type.
	settings["type"] = Docker.String()
	// Merge the settings between common and this builders.
	var workSlice []string
	_, ok = r.Builders[Common.String()]
	if ok {
		workSlice, err = mergeSettingsSlices(r.Builders[Common.String()].Settings, r.Builders[Docker.String()].Settings)
		if err != nil {
			return nil, fmt.Errorf("createDocker failed: %s", err.Error())
		}
	} else {
		workSlice = r.Builders[Docker.String()].Settings
	}
	// Go through each element in the slice, only take the ones that matter
	// to this builder.
	var hasCommit, hasExportPath, hasImage, hasRunCommandArray bool
	var runCommandFile string
	for _, s := range workSlice {
		k, v := parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "export_path":
			settings[k] = v
			hasExportPath = true
		case "image":
			settings[k] = v
			hasImage = true
		case "login_email", "login_username", "login_password", "login_server":
			settings[k] = v
		case "commit":
			settings[k], _ = strconv.ParseBool(v)
			hasCommit = true
		case "login", "pull":
			settings[k], _ = strconv.ParseBool(v)
		case "run_command":
			// if it's here, cache the value, delay processing until arrays section
			runCommandFile = v
		}
	}
	if !hasCommit {
		err := fmt.Errorf("\"commit\" setting is required for docker, not found")
		return nil, err
	}
	if !hasExportPath {
		err := fmt.Errorf("\"export_path\" setting is required for docker, not found")
		return nil, err
	}
	if !hasImage {
		err := fmt.Errorf("\"image\" setting is required for docker, not found")
		return nil, err
	}
	// Process the Arrays.
	for name, val := range r.Builders[Docker.String()].Arrays {
		if name == "run_command" {
			array := deepcopy.InterfaceToSliceOfStrings(val)
			if array != nil {
				settings[name] = array
			}
			hasRunCommandArray = true
		}
	}
	// if there wasn't an array of run commands, check to see if they should be loaded
	// from a file
	if !hasRunCommandArray {
		if runCommandFile != "" {
			commands, err := r.commandsFromFile(Docker.String(), runCommandFile)
			if err != nil {
				return nil, err
			}
			settings["run_command"] = commands
		}
	}
	return settings, nil
}

// createGoogleCompute creates a map of settings for Packer's googlecompute
// builder.  Any values that aren't supported by the googlecompute builder are
// ignored.  Any required settings that don't exist result in an error and
// processing of the builder is stopped.  For more information, refer to
// https://packer.io/docs/builders/googlecompute.html
//
// Required configuration options:
//   project_id         string
//   source_image       string
//   zone               string
// Optional configuration options:
//   account_file       string
//   disk_size          integer
//   image_name         string
//   image_description  string
//   instance_name      string
//   machine_name       string
//   machine_type       string
//   network            string
//   ssh_port           integer
//   ssh_username       string
//   state_timeout      string
//   tags               array of strings
// Not implemented configuration options:
//   metadata           object of key/value strings
func (r *rawTemplate) createGoogleCompute() (settings map[string]interface{}, err error) {
	_, ok := r.Builders[GoogleCompute.String()]
	if !ok {
		err = fmt.Errorf("no configuration found for %q", GoogleCompute.String())
		return nil, err
	}
	settings = make(map[string]interface{})
	// Each create function is responsible for setting its own type.
	settings["type"] = GoogleCompute.String()
	// Merge the settings between common and this builders.
	var workSlice []string
	_, ok = r.Builders[Common.String()]
	if ok {
		workSlice, err = mergeSettingsSlices(r.Builders[Common.String()].Settings, r.Builders[GoogleCompute.String()].Settings)
		if err != nil {
			return nil, fmt.Errorf("createGoogleCompute failed: %s", err.Error())
		}
	} else {
		workSlice = r.Builders[GoogleCompute.String()].Settings
	}
	// Go through each element in the slice, only take the ones that matter
	// to this builder.
	var hasProjectID, hasSourceImage, hasZone bool
	for _, s := range workSlice {
		k, v := parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "project_id":
			settings[k] = v
			hasProjectID = true
		case "source_image":
			settings[k] = v
			hasSourceImage = true
		case "zone":
			settings[k] = v
			hasZone = true
		case "image_name", "image_description", "instance_name",
			"machine_type", "network", "ssh_timeout", "ssh_username", "state_timeout":
			settings[k] = v
		case "account_file":
			src, err := r.findComponentSource(GoogleCompute.String(), v)
			if err != nil {
				return nil, err
			}
			r.files[r.buildOutPath(GoogleCompute.String(), v)] = src
			settings[k] = r.buildTemplateResourcePath(GoogleCompute.String(), v)
		case "disk_size", "ssh_port":
			i, err := strconv.Atoi(v)
			if err != nil {
				err = fmt.Errorf("An error occurred while trying to set %s to %s: %s", k, v, err)
				return nil, err
			}
			settings[k] = i
		}
	}
	if !hasProjectID {
		err := fmt.Errorf("\"project_id\" setting is required for googlecompute, not found")
		return nil, err
	}
	if !hasSourceImage {
		err := fmt.Errorf("\"source_image\" setting is required for googlecompute, not found")
		return nil, err
	}
	if !hasZone {
		err := fmt.Errorf("\"zone\" setting is required for googlecompute, not found")
		return nil, err
	}
	// Process the Arrays.
	for name, val := range r.Builders[GoogleCompute.String()].Arrays {
		if name == "tags" {
			array := deepcopy.InterfaceToSliceOfStrings(val)
			if array != nil {
				settings[name] = array
			}
		}
	}
	return settings, nil
}

// createNull creates a map of settings for Packer's null builder. Any  values
// that aren't supported by the null builder are ignored.  Any required
// settings that don't exist result in an error and processing of the builder
// is stopped.  For more information, refer to
// https://packer.io/docs/builders/null.html
//
// Required configuration options:
//   host string
//   ssh_password string
//   ssh_privateKey_file string
//   ssh_username string
// Optional configuration options:
//   port            integer
func (r *rawTemplate) createNull() (settings map[string]interface{}, err error) {
	_, ok := r.Builders[Null.String()]
	if !ok {
		err = fmt.Errorf("no configuration found for %q", Null.String())
		return nil, err
	}
	settings = make(map[string]interface{})
	// Each create function is responsible for setting its own type.
	settings["type"] = Null.String()
	// Merge the settings between common and this builders.
	var workSlice []string
	_, ok = r.Builders[Common.String()]
	if ok {
		workSlice, err = mergeSettingsSlices(r.Builders[Common.String()].Settings, r.Builders[Null.String()].Settings)
		if err != nil {
			return nil, fmt.Errorf("createNull failed: %s", err.Error())
		}
	} else {
		workSlice = r.Builders[Null.String()].Settings
	}
	// Go through each element in the slice, only take the ones that matter
	// to this builder.
	for _, s := range workSlice {
		k, v := parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "host", "ssh_password", "ssh_private_key_file", "ssh_username":
			settings[k] = v
		case "port":
			// only add if its an int
			i, err := strconv.Atoi(v)
			if err != nil {
				err = fmt.Errorf("%s builder error while processing setting %q: %s", Null.String(), k, err)
				return nil, err
			}
			settings[k] = i
		}
	}
	return settings, nil
}

// createVirtualBoxISO creates a map of settings for Packer's virtualbox-iso
// builder.  Any values that aren't supported by the virtualbox-iso builder are
// ignored.  Any required settings that doesn't exist result in an error and
// processing of the builder is stopped. For more information, refer to
// https://packer.io/docs/builders/virtualbox-iso.html
//
// Required configuration options:
//   iso_checksum             string
//   iso_checksum_type        string
//   iso_url                  string
//   ssh_username             string
// Optional configuration options:
//   boot_command             array of strings*
//   boot_wait                string
//   disk_size                integer
//   export_opts              array of strings
//   floppy_files             array of strings
//   format                   string; "ovf" or "ova"
//   guest_additions_mode     string
//   guest_additions_path     string
//   guest_additions_sha256   string
//   guest_additions_url      string
//   guest_os_type            string; if empty, generated by rancher
//   hard_drive_interface     string
//   headless                 boolean
//   http_directory           string
//   http_port_min            integer
//   http_port_max            integer
//   iso_interface            string
//   iso_urls                 array_of_strings
//   output_directory         string
//   shutdown_command         string
//   shutdown_timeout         string
//   ssh_host_port_min        integer
//   ssh_host_port_max        integer
//   ssh_key_path             string
//   ssh_password             string
//   ssh_port                 integer
//   ssh_wait_timeout         string
//   vboxmanage               array of array of strings
//   vboxmanage_post          array of array of strings
//   virtualbox_version_file  string
//   vm_name                  string
func (r *rawTemplate) createVirtualBoxISO() (settings map[string]interface{}, err error) {
	_, ok := r.Builders[VirtualBoxISO.String()]
	if !ok {
		err = fmt.Errorf("no configuration found for %q", VirtualBoxISO.String())
		return nil, err
	}
	settings = make(map[string]interface{})
	// Each create function is responsible for setting its own type.
	settings["type"] = VirtualBoxISO.String()
	// Merge the settings between common and this builders.
	var workSlice []string
	_, ok = r.Builders[Common.String()]
	if ok {
		workSlice, err = mergeSettingsSlices(r.Builders[Common.String()].Settings, r.Builders[VirtualBoxISO.String()].Settings)
		if err != nil {
			return nil, fmt.Errorf("createVirtualBoxISO failed: %s", err.Error())
		}
	} else {
		workSlice = r.Builders[VirtualBoxISO.String()].Settings
	}
	var k, v string
	var bootCmdProcessed, hasSSHUsername bool
	var tmpISOChecksumType, tmpISOChecksum, tmpISOUrl, tmpGuestOSType string
	// Go through each element in the slice, only take the ones that matter
	// to this builder.
	for _, s := range workSlice {
		// var tmp interface{}
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "boot_command":
			// if the boot_command exists in the Settings section, it should
			// reference a file. This boot_command takes precedence over any
			// boot_command in the array defined in the Arrays section.
			if strings.HasSuffix(v, ".command") {
				var commands []string
				commands, err = r.commandsFromFile("", v)
				if err != nil {
					return nil, err
				}
				if len(commands) == 0 {
					err = fmt.Errorf("%s: error getting %s from %s file, no commands were found", VirtualBoxISO.String(), k, v)
					return nil, err
				}
				settings[k] = commands
				bootCmdProcessed = true
			}
		case "shutdown_command":
			//If it ends in .command, replace it with the command from the filepath
			if strings.HasSuffix(v, ".command") {
				var commands []string
				commands, err = r.commandsFromFile("", v)
				if err != nil {
					return nil, err
				}
				if len(commands) == 0 {
					err = fmt.Errorf("%s: error getting %s from %s file, no commands were found", VirtualBoxISO.String(), k, v)
					return nil, err
				}
				// Assume it's the first element.
				settings[k] = commands[0]
			} else {
				settings[k] = v // the value is the command
			}
		case "boot_wait", "format", "guest_additions_mode", "guest_additions_path",
			"guest_additions_sha256", "guest_additions_url", "hard_drive_interface",
			"http_directory", "iso_interface", "output_directory", "shutdown_timeout",
			"ssh_key_path", "ssh_password", "ssh_wait_timeout", "virtualbox_version_file",
			"vm_name":
			settings[k] = v
		case "guest_os_type":
			tmpGuestOSType = v
		case "ssh_username":
			settings[k] = v
			hasSSHUsername = true
		case "headless":
			settings[k], _ = strconv.ParseBool(v)
		case "iso_checksum_type":
			settings[k] = v
			tmpISOChecksumType = v
		case "iso_checksum":
			settings[k] = v
			tmpISOChecksum = v
		case "iso_url":
			settings[k] = v
			tmpISOUrl = v
		case "disk_size", "http_port_min", "http_port_max", "ssh_host_port_min",
			"ssh_host_port_max", "ssh_port":
			// only add if its an int
			i, err := strconv.Atoi(v)
			if err != nil {
				err = fmt.Errorf("VirtualBoxISO: An error occurred while trying to set %q to %q: %s ", k, v, err)
				return nil, err
			}
			settings[k] = i
		}
	}

	// Only check to see if the required ssh_username field was set. The required iso info is checked after Array processing
	if !hasSSHUsername {
		err = fmt.Errorf("\"ssh_username\" is a required setting for virtualbox-iso; not found")
		return nil, err
	}
	// Process arrays, iso_urls is only valid if iso_url is not set so we first
	// check to see if it has been set, and if not, if it's in this array prior
	// to ranging through the rest of the elements. The range ignores iso_url
	if tmpISOUrl != "" {
		isoURL, ok := r.Builders[VirtualBoxISO.String()].Arrays["iso_url"].(string)
		if !ok {
			goto noISOURL
		}
		tmpISOUrl = isoURL
		if tmpISOChecksum == "" {
			err = fmt.Errorf("\"iso_url\" found for virtualbox-iso but no \"iso_checksum\" information was found")
			return nil, err
		}
		if tmpISOChecksumType == "" {
			err = fmt.Errorf("\"iso_url\" found for virtualbox-iso but no \"iso_checksum_type\" information was found")
			return nil, err
		}
		settings["iso_url"] = isoURL
	}

	// make sure http_directory is set and add to dir list
	err = r.setHTTP(VirtualBoxISO.String(), settings)
	if err != nil {
		return nil, err
	}

noISOURL:

	for name, val := range r.Builders[VirtualBoxISO.String()].Arrays {
		switch name {
		case "boot_command":
			if bootCmdProcessed {
				continue // if the boot command was already set, don't use this array
			}
			settings[name] = val
		case "export_opts", "floppy_files":
			settings[name] = val
		case "iso_url":
			continue // skip as it was processed before the range
		case "iso_urls":
			// these are only added if iso_url isn't set
			if tmpISOUrl == "" {
				if tmpISOChecksum == "" {
					err = fmt.Errorf("\"iso_urls\" found for virtualbox-iso but no \"iso_checksum\" information was found")
					return nil, err
				}
				if tmpISOChecksumType == "" {
					err = fmt.Errorf("\"iso_urls\" found for virtualbox-iso but no \"iso_checksum_type\" information was found")
					return nil, err
				}
				settings[name] = val
			}
		case "vboxmanage", "vboxmanage_post":
			settings[name] = r.createVBoxManage(val)
		}
	}

	if r.osType == "" { // if the os type hasn't been set, the ISO info hasn't been retrieved
		err = r.ISOInfo(VirtualBoxISO, workSlice)
		if err != nil {
			return nil, err
		}
	}
	// set the guest_os_type
	if tmpGuestOSType == "" {
		tmpGuestOSType = r.osType
	}
	settings["guest_os_type"] = tmpGuestOSType

	// If the iso info wasn't set from the Settings, get it from the distro's release
	if tmpISOUrl == "" {
		//handle iso lookup vs set in file
		switch r.Distro {
		case CentOS.String():
			settings["iso_url"] = r.releaseISO.(*centOS).isoURL
			settings["iso_checksum"] = r.releaseISO.(*centOS).Checksum
			settings["iso_checksum_type"] = r.releaseISO.(*centOS).ChecksumType
		case Debian.String():
			settings["iso_url"] = r.releaseISO.(*debian).isoURL
			settings["iso_checksum"] = r.releaseISO.(*debian).Checksum
			settings["iso_checksum_type"] = r.releaseISO.(*debian).ChecksumType
		case Ubuntu.String():
			settings["iso_url"] = r.releaseISO.(*ubuntu).isoURL
			settings["iso_checksum"] = r.releaseISO.(*ubuntu).Checksum
			settings["iso_checksum_type"] = r.releaseISO.(*ubuntu).ChecksumType
		default:
			err = fmt.Errorf("%q is not a supported Distro", r.Distro)
			return nil, err
		}
		return settings, nil
	}
	if tmpISOChecksumType == "" {
		err = fmt.Errorf("\"iso_url\" information was set for virtualbox-iso but the \"iso_checksum_type\" was not")
		return nil, err
	}
	if tmpISOChecksum == "" {
		err = fmt.Errorf("\"iso_url\" information was set for virtualbox-iso but the \"iso_checksum\" was not")
		return nil, err
	}
	return settings, nil
}

// createVirtualBoxOVF creates a map of settings for Packer's virtualbox-ovf
// builder.  Any values that aren't supported by the virtualbox-ovf builder are
// ignored.  Any required settings that don't exist result in an error and
// processing of the builder is stopped. For more information, refer to
// https://packer.io/docs/builders/virtualbox-ovf.html
//
// Required configuration options:
//   source_path              string
//   ssh_username             string
// Optional configuration options:
//   boot_command             array of strings*
//   boot_wait                string
//   export_opts              array of strings
//   floppy_files             array of strings
//   format                   string
//   guest_additions_mode     string
//   guest_additions_path     string
//   guest_additions_sha256	   string
//   guest_additions_url      string
//   headless                 boolean
//   http_directory           string
//   http_port_min            integer
//   http_port_max            integer
//   import_flags             array of strings
//   import_opts              string
//   output_directory         string
//   shutdown_command         string
//   shutdown_timeout         string
//   ssh_host_port_min        integer
//   ssh_host_port_max        integer
//   ssh_key_path             string
//   ssh_password             string
//   ssh_port                 integer
//   ssh_wait_timeout         string
//   vboxmanage               array of strings
//   vboxmanage_post          array of strings
//   virtualbox_version_file  string
//   vm_name                  string
func (r *rawTemplate) createVirtualBoxOVF() (settings map[string]interface{}, err error) {
	_, ok := r.Builders[VirtualBoxOVF.String()]
	if !ok {
		err = fmt.Errorf("no configuration found for %q", VirtualBoxOVF.String())
		return nil, err
	}
	settings = make(map[string]interface{})
	// Each create function is responsible for setting its own type.
	settings["type"] = VirtualBoxOVF.String()
	// Merge the settings between common and this builders.
	var workSlice []string
	_, ok = r.Builders[Common.String()]
	if ok {
		workSlice, err = mergeSettingsSlices(r.Builders[Common.String()].Settings, r.Builders[VirtualBoxOVF.String()].Settings)
		if err != nil {
			return nil, fmt.Errorf("createVirtualBoxOVF failed: %s", err.Error())
		}
	} else {
		workSlice = r.Builders[VirtualBoxOVF.String()].Settings
	}
	// Go through each element in the slice, only take the ones that matter
	// to this builder.
	var hasSourcePath, hasSSHUsername, bootCmdProcessed bool
	for _, s := range workSlice {
		// var tmp interface{}
		k, v := parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "boot_command":
			// if the boot_command exists in the Settings section, it should
			// reference a file. This boot_command takes precedence over any
			// boot_command in the array defined in the Arrays section.
			if strings.HasSuffix(v, ".command") {
				var commands []string
				commands, err = r.commandsFromFile("", v)
				if err != nil {
					return nil, err
				}
				if len(commands) == 0 {
					err = fmt.Errorf("%s: error getting %s from %s file, no commands were found", VirtualBoxOVF.String(), k, v)
					return nil, err
				}
				settings[k] = commands
				bootCmdProcessed = true
			}
		case "source_path":
			src, err := r.findComponentSource(VirtualBoxOVF.String(), v)
			if err != nil {
				return nil, err
			}
			settings[k] = r.buildTemplateResourcePath(VirtualBoxOVF.String(), v)
			r.files[r.buildOutPath(VirtualBoxOVF.String(), v)] = src
			hasSourcePath = true
		case "ssh_username":
			settings[k] = v
			hasSSHUsername = true
		case "boot_wait", "format", "guest_additions_mode", "guest_additions_path",
			"guest_additions_sha256", "guest_additions_url", "http_directory",
			"import_opts", "output_directory", "shutdown_timeout", "ssh_key_path",
			"ssh_password", "ssh_wait_timeout", "virtualbox_version_file", "vm_name":
			settings[k] = v
		case "headless":
			settings[k], _ = strconv.ParseBool(v)
		// For the fields of int value, only set if it converts to a valid int.
		// Otherwise, throw an error
		case "http_port_min", "http_port_max", "ssh_host_port_min", "ssh_host_port_max", "ssh_port":
			// only add if its an int
			i, err := strconv.Atoi(v)
			if err != nil {
				err = fmt.Errorf("VirtualBoxOVF error while trying to set %q to %q: %s", k, v, err)
				return nil, err
			}
			settings[k] = i
		case "shutdown_command":
			if strings.HasSuffix(v, ".command") {
				//If it ends in .command, replace it with the command from the filepath
				var commands []string
				commands, err = r.commandsFromFile("", v)
				if err != nil {
					return nil, err
				}
				if len(commands) == 0 {
					err = fmt.Errorf("%s: error getting %s from %s file, no commands were found", VirtualBoxOVF.String(), k, v)
					return nil, err
				}
				// Assume it's the first element.
				settings[k] = commands[0]
			} else {
				settings[k] = v
			}
		}
	}
	// Check to see if the required info was processed.
	if !hasSSHUsername {
		err = fmt.Errorf("\"ssh_username\" is a required setting for virtualbox-ovf; not found")
		return nil, err
	}
	if !hasSourcePath {
		err = fmt.Errorf("\"source_path\" is a required setting for virtualbox-ovf; not found")
		return nil, err
	}

	// make sure http_directory is set and add to dir list
	err = r.setHTTP(VirtualBoxOVF.String(), settings)
	if err != nil {
		return nil, err
	}

	// Generate Packer Variables
	// Generate builder specific section
	for name, val := range r.Builders[VirtualBoxOVF.String()].Arrays {
		switch name {
		case "boot_command":
			if bootCmdProcessed {
				continue // if the boot command was already set, don't use this array
			}
			settings[name] = val
		case "export_opts", "floppy_files":
			settings[name] = val
		case "vboxmanage", "vboxmanage_post":
			settings[name] = r.createVBoxManage(val)
		}
	}
	return settings, nil
}

// createVBoxManage creates the vboxmanage and vboxmanage_post arrays from the
// received interface.
func (r *rawTemplate) createVBoxManage(v interface{}) [][]string {
	vms := deepcopy.InterfaceToSliceOfStrings(v)
	tmp := make([][]string, len(vms))
	for i, v := range vms {
		k, vv := parseVar(v)
		// ensure that the key starts with --. A naive concatonation is done.
		if !strings.HasPrefix(k, "--") {
			k = "--" + k
		}
		vv = r.replaceVariables(vv)
		tmp[i] = make([]string, 4)
		tmp[i][0] = "modifyvm"
		tmp[i][1] = "{{.Name}}"
		tmp[i][2] = k
		tmp[i][3] = vv
	}
	return tmp
}

// createVMWareISO creates a map of settings for Packer's vmware-iso builder.
// Any values that aren't supported by the vmware-iso builder are ignored.  Any
// required settings that don't exist result in an error and processing of the
// builder is stopped. For more information, refer to
// https://packer.io/docs/builders/vmware-iso.html
//
// Required configuration options:
//   iso_checksum            string
//	 iso_checksum_type       string
//	 iso_url                 string
//   ssh_username            string
// Optional configuration options
//   boot_command            array of strings*
//   boot_wait               string
//   disk_size               integer
//   disk_type_id            string
//   floppy_files            array of strings
//   fusion_app_path         string
//   guest_os_type           string; if not set, will be generated
//   headless                boolean
//   http_directory          string
//   http_port_min           integer
//   http_port_max           integer
//   iso_urls                array of strings
//   output_directory        string
//   remote_cache_datastore  string
//   remote_cache_directory	  string
//   remote_datastore        string
//   remote_host             string
//   remote_password         string
//   remote_type             string
//   remote_username         string
//   shutdown_command        string
//   shutdown_timeout        string
//   skip_compaction         boolean
//   ssh_host                string
//   ssh_key_path            string
//   ssh_password            string
//   ssh_port                integer
//   ssh_skip_request_pty    boolean
//   ssh_wait_timeout        string
//   tools_upload_flavor     string
//   tools_upload_path       string
//   version                 string
//   vm_name                 string
//   vmdk_name               string
//   vmx_data                object of key/value strings
//   vmx_data_post           object of key/value strings
//   vmx_template_path       string
//   vnc_port_min            integer
//   vnc_port_max            integer
func (r *rawTemplate) createVMWareISO() (settings map[string]interface{}, err error) {
	_, ok := r.Builders[VMWareISO.String()]
	if !ok {
		err = fmt.Errorf("no configuration found for %q", VMWareISO.String())
		return nil, err
	}
	settings = make(map[string]interface{})
	// Each create function is responsible for setting its own type.
	settings["type"] = VMWareISO.String()
	// Merge the settings between common and this builders.
	var workSlice []string
	_, ok = r.Builders[Common.String()]
	if ok {
		workSlice, err = mergeSettingsSlices(r.Builders[Common.String()].Settings, r.Builders[VMWareISO.String()].Settings)
		if err != nil {
			return nil, fmt.Errorf("createVMWareISO failed: %s", err.Error())
		}
	} else {
		workSlice = r.Builders[VMWareISO.String()].Settings
	}
	// Go through each element in the slice, only take the ones that matter
	// to this builder.
	var bootCmdProcessed, hasSSHUsername bool
	var tmpISOChecksum, tmpISOChecksumType, tmpISOUrl, tmpGuestOSType string
	for _, s := range workSlice {
		// var tmp interface{}
		k, v := parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "boot_command":
			// if the boot_command exists in the Settings section, it should
			// reference a file. This boot_command takes precedence over any
			// boot_command in the array defined in the Arrays section.
			if strings.HasSuffix(v, ".command") {
				var commands []string
				commands, err = r.commandsFromFile("", v)
				if err != nil {
					return nil, err
				}
				if len(commands) == 0 {
					err = fmt.Errorf("%s: error getting %s from %s file, no commands were found", VMWareISO.String(), k, v)
					return nil, err
				}
				settings[k] = commands
				bootCmdProcessed = true
			}
		case "shutdown_command":
			//If it ends in .command, replace it with the command from the filepath
			if strings.HasSuffix(v, ".command") {
				var commands []string
				commands, err = r.commandsFromFile("", v)
				if err != nil {
					return nil, err
				}
				if len(commands) == 0 {
					err = fmt.Errorf("%s: error getting %s from %s file, no commands were found", VMWareISO.String(), k, v)
					return nil, err
				}
				// Assume it's the first element.
				settings[k] = commands[0]
			} else {
				settings[k] = v // the value is the command
			}
		case "boot_wait", "disk_type_id", "fusion_app_path", "http_directory",
			"output_directory", "remote_cache_datastore", "remote_cache_directory",
			"remote_datastore", "remote_host", "remote_password", "remote_type",
			"remote_username", "shutdown_timeout", "ssh_host", "ssh_key_path",
			"ssh_password", "ssh_wait_timeout", "tools_upload_flavor", "tools_upload_path",
			"vm_name", "vmdk_name", "vmx_template_path":
			settings[k] = v
		case "guest_os_type":
			tmpGuestOSType = v
		case "ssh_username":
			settings[k] = v
			hasSSHUsername = true
		case "headless":
			settings[k], _ = strconv.ParseBool(v)
		case "iso_checksum_type":
			settings[k] = v
			tmpISOChecksumType = v
		case "iso_checksum":
			settings[k] = v
			tmpISOChecksum = v
		case "iso_url":
			settings[k] = v
			tmpISOUrl = v
		case "disk_size", "http_port_min", "http_port_max", "ssh_port", "vnc_port_min",
			"vnc_port_max":
			// only add if its an int
			i, err := strconv.Atoi(v)
			if err != nil {
				err = fmt.Errorf("vmware-iso: An error occurred while trying to set %q to %q: %s ", k, v, err)
				return nil, err
			}
			settings[k] = i
		}
	}

	// Only check to see if the required ssh_username field was set. The required iso info is checked after Array processing
	if !hasSSHUsername {
		err = fmt.Errorf("\"ssh_username\" is a required setting for virtualbox-iso; not found")
		return nil, err
	}

	// make sure http_directory is set and add to dir list
	err = r.setHTTP(VMWareISO.String(), settings)
	if err != nil {
		return nil, err
	}

	// Process arrays, iso_urls is only valid if iso_url is not set
	for name, val := range r.Builders[VMWareISO.String()].Arrays {
		switch name {
		case "boot_command":
			if bootCmdProcessed {
				continue // if the boot command was already set, don't use this array
			}
			settings[name] = val
		case "floppy_files":
			settings[name] = val
		case "iso_urls":
			// these are only added if iso_url isn't set
			if tmpISOUrl == "" {
				if tmpISOChecksum == "" {
					err = fmt.Errorf("\"iso_urls\" found for vmware-iso but no \"iso_checksum\" information was found")
					return nil, err
				}
				if tmpISOChecksumType == "" {
					err = fmt.Errorf("\"iso_urls\" found for vmware-iso but no \"iso_checksum_type\" information was found")
					return nil, err
				}
				settings[name] = val
			}
		case "vmx_data", "vmx_data_post":
			settings[name] = r.createVMXData(val)
		}
	}

	if r.osType == "" { // if the os type hasn't been set, the ISO info hasn't been retrieved
		err = r.ISOInfo(VirtualBoxISO, workSlice)
		if err != nil {
			return nil, err
		}
	}
	// set the guest_os_type
	if tmpGuestOSType == "" {
		tmpGuestOSType = r.osType
	}
	settings["guest_os_type"] = tmpGuestOSType

	// If the iso info wasn't set from the Settings, get it from the distro's release
	if tmpISOUrl == "" {
		//handle iso lookup vs set in file
		switch r.Distro {
		case CentOS.String():
			settings["iso_url"] = r.releaseISO.(*centOS).isoURL
			settings["iso_checksum"] = r.releaseISO.(*centOS).Checksum
			settings["iso_checksum_type"] = r.releaseISO.(*centOS).ChecksumType
		case Debian.String():
			err = fmt.Errorf("automatic resolution of iso information for %q is not supported, the \"iso_url\" and \"iso_checksum\" settings must exist", r.Distro)
			return nil, err
		case Ubuntu.String():
			settings["iso_url"] = r.releaseISO.(*ubuntu).isoURL
			settings["iso_checksum"] = r.releaseISO.(*ubuntu).Checksum
			settings["iso_checksum_type"] = r.releaseISO.(*ubuntu).ChecksumType
		default:
			err = fmt.Errorf("%q is not a supported Distro", r.Distro)
			return nil, err
		}
		return settings, nil
	}
	if tmpISOChecksumType == "" {
		err = fmt.Errorf("\"iso_url\" information was set for vmware-iso but the \"iso_checksum_type\" was not")
		return nil, err
	}
	if tmpISOChecksum == "" {
		err = fmt.Errorf("\"iso_url\" information was set for vmware-iso but the \"iso_checksum\" was not")
		return nil, err
	}
	return settings, nil
}

// createVMWareVMX creates a map of settings for Packer's vmware-vmx builder.
// Any values that aren't supported by the vmware-vmx builder are ignored.  Any
// required settings that don't exist result in an error and processing of the
// builder is stopped.  For more information, refer to
// https://packer.io/docs/builders/vmware-vmx.html
//
// Required configuration options:
//   source_name				// string
//   ssh_username				// string
// Optional configuration options
//   boot_command				// array of strings*
//   boot_wait					// string
//   floppy_files				// array of strings
//   fusion_app_path			// string
//   headless					// boolean
//   http_directory				// string
//   http_port_min				// integer
//   http_port_max				// integer
//   output_directory			// string
//   shutdown_command			// string
//   shutdown_timeout			// string
//   skip_compaction			// boolean
//   ssh_key_path				// string
//   ssh_password				// string
//   ssh_port					// integer
//   ssh_skip_request_pty		// boolean
//   ssh_wait_timeout			// string
//   vm_name					// string
//   vmx_data					// object of key/value strings
//   vmx_data_post				// object of key/value strings
//   vnc_port_min				// integer
//   vnc_port_max				// integer
func (r *rawTemplate) createVMWareVMX() (settings map[string]interface{}, err error) {
	_, ok := r.Builders[VMWareVMX.String()]
	if !ok {
		err = fmt.Errorf("no configuration found for %q", VMWareVMX.String())
		return nil, err
	}
	settings = make(map[string]interface{})
	// Each create function is responsible for setting its own type.
	settings["type"] = VMWareVMX.String()
	// Merge the settings between common and this builders.
	var workSlice []string
	_, ok = r.Builders[Common.String()]
	if ok {
		workSlice, err = mergeSettingsSlices(r.Builders[Common.String()].Settings, r.Builders[VMWareVMX.String()].Settings)
		if err != nil {
			return nil, fmt.Errorf("createVMWareVMX failed: %s", err.Error())
		}
	} else {
		workSlice = r.Builders[VMWareVMX.String()].Settings
	}
	var hasSourcePath, hasSSHUsername, bootCmdProcessed bool
	// Go through each element in the slice, only take the ones that matter
	// to this builder.
	for _, s := range workSlice {
		// var tmp interface{}
		k, v := parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "boot_command":
			// if the boot_command exists in the Settings section, it should
			// reference a file. This boot_command takes precedence over any
			// boot_command in the array defined in the Arrays section.
			if strings.HasSuffix(v, ".command") {
				var commands []string
				commands, err = r.commandsFromFile("", v)
				if err != nil {
					return nil, err
				}
				if len(commands) == 0 {
					err = fmt.Errorf("%s: error getting %s from %s file, no commands were found", VMWareVMX.String(), k, v)
					return nil, err
				}
				settings[k] = commands
				bootCmdProcessed = true
			}
		case "shutdown_command":
			//If it ends in .command, replace it with the command from the filepath
			if strings.HasSuffix(v, ".command") {
				var commands []string
				commands, err = r.commandsFromFile("", v)
				if err != nil {
					return nil, err
				}
				if len(commands) == 0 {
					err = fmt.Errorf("%s: error getting %s from %s file, no commands were found", VMWareVMX.String(), k, v)
					return nil, err
				}
				// Assume it's the first element.
				settings[k] = commands[0]
			} else {
				settings[k] = v // the value is the command
			}
		case "source_path":
			src, err := r.findComponentSource(VMWareVMX.String(), v)
			if err != nil {
				return nil, err
			}
			r.files[r.buildOutPath(VMWareVMX.String(), v)] = src
			settings[k] = r.buildTemplateResourcePath(VMWareVMX.String(), v)
			hasSourcePath = true
		case "ssh_username":
			settings[k] = v
			hasSSHUsername = true
		case "boot_wait", "fusion_app_path", "http_directory", "output_directory", "shutdown_timeout",
			"ssh_key_path", "ssh_password", "ssh_wait_timeout", "vm_name":
			settings[k] = v
		case "headless", "skip_compaction", "ssh_skip_request_pty":
			settings[k], _ = strconv.ParseBool(v)
		// For the fields of int value, only set if it converts to a valid int.
		// Otherwise, throw an error
		case "http_port_max", "http_port_min", "ssh_port", "vnc_port_max", "vnc_port_min":
			// only add if its an int
			i, err := strconv.Atoi(v)
			if err != nil {
				err = fmt.Errorf("vmware-vmx error while trying to set %q to %q: %s", k, v, err)
				return nil, err
			}
			settings[k] = i
		}
	}
	// Check if required fields were processed
	if !hasSSHUsername {
		err = fmt.Errorf("\"ssh_username\" is a required setting for vmware-vmx; not found")
		return nil, err
	}
	if !hasSourcePath {
		err = fmt.Errorf("\"source_path\" is a required setting for vmware-vmx; not found")
		return nil, err
	}
	// make sure http_directory is set and add to dir list
	err = r.setHTTP(VMWareVMX.String(), settings)
	if err != nil {
		return nil, err
	}

	// Process arrays, iso_urls is only valid if iso_url is not set
	for name, val := range r.Builders[VMWareVMX.String()].Arrays {
		switch name {
		case "boot_command":
			if bootCmdProcessed {
				continue // if the boot command was already set, don't use this array
			}
			settings[name] = val
		case "floppy_files":
			settings[name] = val
		case "vmx_data", "vmx_data_post":
			settings[name] = r.createVMXData(val)
		}
	}
	return settings, nil
}

func (r *rawTemplate) createVMXData(v interface{}) map[string]string {
	vms := deepcopy.InterfaceToSliceOfStrings(v)
	tmp := make(map[string]string, len(vms))
	for _, v := range vms {
		k, val := parseVar(v)
		val = r.replaceVariables(val)
		tmp[k] = val
	}
	return tmp
}

// updateBuilders updates the rawTemplate's builders with the passed new
// builder.
// Builder Update rules:
//   * If r's old builder does not have a matching builder in the new builder
//     map, new, nothing is done.
//   * If the builder exists in both r and new, the new builder updates r's
//     builder.
//   * If the new builder does not have a matching builder in r, the new
//     builder is added to r's builder map.
//
// Settings update rules:
//   * If the setting exists in r's builder but not in new, nothing is done.
//     This means that deletion of settings via not having them exist in the
//     new builder is not supported. This is to simplify overriding templates
//     in the configuration files.
//   * If the setting exists in both r's builder and new, r's builder is
//     updated with new's value.
//   * If the setting exists in new, but not r's builder, new's setting is
//     added to r's builder.
//   * To unset a setting, specify the key, without a value: `"key="`.  In most
//     situations, Rancher will interpret an key without a value as a deletion
//     of that key. There is an exception:
//   * `guest_os_type`: This is generally set at Packer Template generation
//     time by Rancher.
func (r *rawTemplate) updateBuilders(newB map[string]builder) error {
	// If there is nothing new, old equals merged.
	if len(newB) == 0 || newB == nil {
		return nil
	}
	// Convert the existing Builders to interfaces.
	var ifaceOld = make(map[string]interface{}, len(r.Builders))
	ifaceOld = DeepCopyMapStringBuilder(r.Builders)
	// Convert the new Builders to interfaces.
	var ifaceNew = make(map[string]interface{}, len(newB))
	ifaceNew = DeepCopyMapStringBuilder(newB)
	// Make the slice as long as the slices in both builders, odds are its shorter, but this is the worst case.
	var keys []string
	// Convert the keys to a map
	keys = mergedKeysFromMaps(ifaceOld, ifaceNew)
	// If there's a builder with the key CommonBuilder, merge them. This is a special case for builders only.
	_, ok := newB[Common.String()]
	if ok {
		r.updateCommon(newB[Common.String()])
	}
	// Copy: if the key exists in the new builder only.
	// Ignore: if the key does not exist in the new builder.
	// Merge: if the key exists in both the new and old builder.
	for _, v := range keys {
		// If it doesn't exist in the old builder, add it.
		b, ok := r.Builders[v]
		if !ok {
			bb, _ := newB[v]
			r.Builders[v] = bb.DeepCopy()
			continue
		}
		// If the element for this key doesn't exist, skip it.
		bb, ok := newB[v]
		if !ok {
			continue
		}
		err := b.mergeSettings(bb.Settings)
		if err != nil {
			return fmt.Errorf("merge builder settings failed: %s", err.Error())
		}
		b.mergeArrays(bb.Arrays)
		r.Builders[v] = b
	}
	return nil
}

// updateCommon updates rawTemplate's common builder settings.
// Update rules:
//   * When both the existing common builder, r, and the new one, b, have the
//     same setting, b's value replaces r's; the new value replaces the
//     existing value.
//   * When the setting in b is new, it is added to r: new settings are
//     inserted into r's CommonBuilder setting list.
//   * When r has a setting that does not exist in b, nothing is done.  This
//     method does not delete any settings that already exist in r.
func (r *rawTemplate) updateCommon(newB builder) error {
	if r.Builders == nil {
		r.Builders = map[string]builder{}
	}
	// If the existing builder doesn't have a CommonBuilder section, just add it
	b, ok := r.Builders[Common.String()]
	if !ok {
		r.Builders[Common.String()] = builder{templateSection: templateSection{Settings: newB.Settings, Arrays: newB.Arrays}}
		return nil
	}
	// Otherwise merge the two
	err := b.mergeSettings(b.Settings)
	if err != nil {
		return fmt.Errorf("ipdate common settings failed: %s", err.Error())
	}
	r.Builders[Common.String()] = b
	return nil
}

// setHTTP ensures that http setting is set and adds it to the dirs info so that its
// contents can be copied. If it is not set, http is assumed.
func (r *rawTemplate) setHTTP(component string, m map[string]interface{}) error {
	v, ok := m["http_directory"]
	if !ok {
		v = "http"
	}
	src, err := r.findComponentSource(component, v.(string))
	if err != nil {
		return fmt.Errorf("setHTTP error for %s: %s", component, err)
	}
	r.dirs[r.buildOutPath(component, v.(string))] = src
	m["http_directory"] = r.buildTemplateResourcePath(component, v.(string))
	return nil
}

// DeepCopyMapStringBuilder makes a deep copy of each builder passed and
// returns the copy map[string]builder as a map[string]interface{}
func DeepCopyMapStringBuilder(b map[string]builder) map[string]interface{} {
	c := map[string]interface{}{}
	for k, v := range b {
		tmpB := builder{}
		tmpB = v.DeepCopy()
		c[k] = tmpB
	}
	return c
}
