// builders.go contains all of the builder related functionality for
// rawTemplates. Any new builders should be added here.
package ranchr

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	json "github.com/mohae/customjson"
	jww "github.com/spf13/jwalterweatherman"
)

// r.createBuilders takes a raw builder and create the appropriate Packer
// Builders along with a slice of variables for that section builder type.
// Some Settings are in-lined instead of adding them to the variable section.
//
// At this point, all of the settings 
//
// * update BuilderCommon with the ne, as this may be used by any of the Packer
// builders.
// * For each Builder in the template, create it's Packer Template version
//
func (r *rawTemplate) createBuilders() (bldrs []interface{}, vars map[string]interface{}, err error) {
	if r.BuilderTypes == nil || len(r.BuilderTypes) <= 0 {
		err = fmt.Errorf("rawTemplate.createBuilders: no builder types were configured, unable to create builders")
		jww.ERROR.Println(err.Error())
		return nil, nil, err
	}

	var vrbls, tmpVar []string
	var tmpS map[string]interface{}
	var ndx int
	bldrs = make([]interface{}, len(r.BuilderTypes))

	// Set the BuilderCommon settings. Only the builder.Settings field is used
	// for BuilderCommon as everything else is usually builder specific, even
	// if they have common names, e.g. difference between specifying memory 
	// between VMWare and VirtualBox.
//	r.updateBuilderCommon


	jww.TRACE.Println("rawTemplate.createBuilders:\t" + json.MarshalIndentToString(r, "", indent))
	// Generate the builders for each builder type.
	for _, bType := range r.BuilderTypes {
		jww.TRACE.Println(bType)

		// TODO calculate the length of the two longest Settings and VMSettings sections and make it
		// that length. That will prevent a panic should there be more than 50 options. Besides its
		// stupid, on so many levels, to hard code this...which makes me...d'oh!
		tmpVar = make([]string, 50)
		tmpS = make(map[string]interface{})

		switch bType {
		case BuilderVMWareISO:
//			tmpS, tmpVar, err = r.createBuilderVMWareISO()
		case BuilderVMWareOVF:
//			tmpS, tmpVar, err = r.createBuilderVMWareOVF()

		case BuilderVirtualBoxISO:
			tmpS, tmpVar, err = r.createBuilderVirtualBoxISO()

		case BuilderVirtualBoxOVF:
//			tmpS, tmpVar, err = r.createVirtualBoxOVF()

		default:
			err = errors.New("The requested builder, '" + bType + "', is not supported by Rancher")
			jww.ERROR.Println(err.Error())
			return nil, nil, err
		}

		bldrs[ndx] = tmpS
		ndx++
		vrbls = append(vrbls, tmpVar...)
	}

	return bldrs, vars, nil
}

/*
// r.createBuilderVMWareISO generates the settings for a vmware-iso builder.
func (r *rawTemplate) createBuilderVMWareISO() (settings map[string]interface{}, vars []string, err error) {
	// Generate the common Settings and their vars
	if tmpS, tmpVar, err = r.commonVMSettings(bType, r.Builders[BuilderCommon].Settings, r.Builders[bType].Settings); err != nil {
		jww.ERROR.Println(err.Error())
		return nil, nil, err
	}

	tmpS["type"] = bType

	// Generate builder specific section
	tmpvm := make(map[string]string, len(r.Builders[bType].Arrays[VMSettings]))

	for i, v = range r.Builders[bType].Arrays[VMSettings] {
		k, val = parseVar(v)
		val = r.replaceVariables(val)
		tmpvm[k] = val
		tmpS["vmx_data"] = tmpvm
	}
}
*/

// r.createBuilderVirtualboxISO generates the settings for a vmware-iso builder.
func (r *rawTemplate) createBuilderVirtualBoxISO() (settings map[string]interface{}, vars []string, err error) {
	settings = make(map[string]interface{})

	// Each create function is responsible for setting its own type.
	settings["type"] = BuilderVirtualBoxISO

	jww.TRACE.Println("rawTemplate.createBuilderVirtualBoxISO--common settings:\n" + json.MarshalIndentToString(r.Builders[BuilderCommon].Settings, "", indent))
	jww.TRACE.Println("rawTemplate.createBuilderVirtualBoxISO--settings:\n" + json.MarshalIndentToString(r.Builders[BuilderVirtualBoxISO].Settings, "", indent))
	// Merge the settings between common and this builders.
	mergedSlice := mergeSettingsSlices(r.Builders[BuilderCommon].Settings, r.Builders[BuilderVirtualBoxISO].Settings)
	jww.TRACE.Println("rawTemplate.createBuilderVirtualBoxISO--merged settings:\n" + json.MarshalIndentToString(mergedSlice, "", indent))

	var k, v string

	// Go through each element in the slice, only take the ones that matter
	// to this builder.
	for _, s := range mergedSlice {
		// var tmp interface{}
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "boot_command":
			//If it ends in .command, replace it with the command from the filepath
			var commands []string

			if commands, err = commandsFromFile(v); err != nil {
				jww.ERROR.Println(err.Error())
				return nil, nil, err
			} 

			settings[k] = commands

		case "boot_wait", "export_opts", "floppy_files", "format", "guest_additions_mode",
			"guest_additions_path", "guest_additions_sha256", "guest_additions_url",
			"hard_drive_interface", "http_directory", "ssh_key_path", "ssh_password",
			"ssh_username",	"ssh_wait_timeout", "vboxmanage", "vboxmanage_post",
			"virtualbox_version_file", "vm_name":
			settings[k] = v

		case "guest_os_type":
			if v == "" {
				settings[k] = v
			} else {
				settings[k] = r.osType
			}

		case "headless":
			if strings.ToLower(v) == "true" {
				settings[k] = true
			} else {
				settings[k] = false
			}

		case "iso_checksum_type":
			// First set the ISO info for the desired release, if it's not already set
			if r.osType == "" {
				err = r.ISOInfo(BuilderVirtualBoxISO, mergedSlice)
				if err != nil {
					jww.ERROR.Println(err.Error())
					return nil, nil, err
				}
			}

			switch r.Type {

			case "ubuntu":
				settings["iso_url"] = r.releaseISO.(*ubuntu).isoURL
				settings["iso_checksum"] = r.releaseISO.(*ubuntu).Checksum
				settings["iso_checksum_type"] = r.releaseISO.(*ubuntu).ChecksumType

			case "centos":
				settings["iso_url"] = r.releaseISO.(*centOS).isoURL
				settings["iso_checksum"] = r.releaseISO.(*centOS).Checksum
				settings["iso_checksum_type"] = r.releaseISO.(*centOS).ChecksumType

			case "default":
				err = errors.New("rawTemplate.createBuilderVirtualBoxISO: " + k + " is not a supported builder type")
				jww.ERROR.Println(err.Error())
				return nil, nil, err

			}

		// For the fields of int value, only set if it converts to a valid int. 
		// Otherwise, throw an error
		case "disk_size", "ssh_host_port_min", "ssh_host_port_max", "ssh_port":
			// only add if its an int
			if _, err := strconv.Atoi(v); err != nil {
				return nil, nil, errors.New("rawTemplate.createBuilderVirtualBoxISO: An error occurred while trying to set " + k + "'s value, '" + v + "': " + err.Error())
			}
			settings[k] = v

		case "shutdown_command":
			//If it ends in .command, replace it with the command from the filepath
			var commands []string

			if commands, err = commandsFromFile(v); err != nil {
				jww.ERROR.Println(err.Error())
				return nil, nil, err
			} 

			// Assume it's the first element.
			settings[k] = commands[0]

		}
	}

	// Generate Packer Variables
	// Generate builder specific section
	l, err := getSliceLenFromIface(r.Builders[BuilderVirtualBoxISO].Arrays[VMSettings])
	if err != nil {	
		return nil, nil, err
	}

	tmpVB := make([][]string, l)
	vm_settings := interfaceToStringSlice(r.Builders[BuilderVirtualBoxISO].Arrays[VMSettings])
	for i, v := range vm_settings {
		k, val := parseVar(v)
		val = r.replaceVariables(val)
		tmpVB[i] = make([]string, 4)
		tmpVB[i][0] = "modifyvm"
		tmpVB[i][1] = "{{.Name}}"
		tmpVB[i][2] = "--" + k
		tmpVB[i][3] = val
	}

	settings["vboxmanage"] = tmpVB

	return settings, nil, nil
}

/*
vmx
*/
// r.createBuilderVMWareISO generates the settings for a vmware-iso builder.
func (r *rawTemplate) createBuilderVMWareISO() (settings map[string]interface{}, vars []string, err error) {
	// Each create function is responsible for setting its own type.
	settings["type"] = BuilderVMWareISO

	// Merge the settings between common and this builders.
	mergedSlice := mergeSettingsSlices(r.Builders[BuilderCommon].Settings, r.Builders[BuilderVMWareISO].Settings)

	// Go through each element in the slice, only take the ones that matter
	// to this builder.
	for _, s := range mergedSlice {
		// var tmp interface{}
		k, v := parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "boot_command":
			//If it ends in .command, replace it with the command from the filepath
			var commands []string

			if commands, err = commandsFromFile(v); err != nil {
				jww.ERROR.Println(err.Error())
				return nil, nil, err
			} 

			settings[k] = commands

		case "boot_wait", "disk_size_id", "floppy_files", "fusion_app_path", "http_directory",
			"iso_urls", "output_directory", "remote_datastore", "remote_host", "remote_password",
			"remote_type", "remote_username", "shutdown_timeout", "ssh_host", "ssh_key_path",
			"ssh_password", "ssh_username", "ssh_wait_timeout", "tools_upload_flavor",
			"tools_upload_path", "vm_name", "vmdk_name", "vmx_data", "vmx_data_post", 
			"vmx_template_path":
			settings[k] = v

		case "guest_os_type":
			if v == "" {
				settings[k] = v
			} else {
				settings[k] = r.osType
			}

		case "headless", "skip_compaction", "ssh_skip_request_pty":
			if strings.ToLower(v) == "true" {
				settings[k] = true
			} else {
				settings[k] = false
			}

		case "iso_checksum_type":
			// First set the ISO info for the desired release, if it's not already set
			if r.osType == "" {
				err = r.ISOInfo(BuilderVMWareISO, mergedSlice)
				if err != nil {
					jww.ERROR.Println(err.Error())
					return nil, nil, err
				}
			}

			switch r.Type {

			case "ubuntu":
				settings["iso_url"] = r.releaseISO.(*ubuntu).isoURL
				settings["iso_checksum"] = r.releaseISO.(*ubuntu).Checksum
				settings["iso_checksum_type"] = r.releaseISO.(*ubuntu).ChecksumType

			case "centos":
				settings["iso_url"] = r.releaseISO.(*centOS).isoURL
				settings["iso_checksum"] = r.releaseISO.(*centOS).Checksum
				settings["iso_checksum_type"] = r.releaseISO.(*centOS).ChecksumType

			case "default":
				err = errors.New("rawTemplate.createBuilderVirtualBoxISO: " + k + " is not a supported builder type")
				jww.ERROR.Println(err.Error())
				return nil, nil, err

			}

		// For the fields of int value, only set if it converts to a valid int. 
		// Otherwise, throw an error
		case "disk_size", "http_port_min", "http_port_max", "ssh_host_port_min", "ssh_host_port_max",
			"ssh_port", "vnc_port_min", "vnc_port_max":
			// only add if its an int
			if _, err := strconv.Atoi(v); err != nil {
				return nil, nil, errors.New("rawTemplate.createBuilderVirtualBoxISO: An error occurred while trying to set " + k + "'s value, '" + v + "': " + err.Error())
			}
			settings[k] = v

		case "shutdown_command":
			//If it ends in .command, replace it with the command from the filepath
			var commands []string

			if commands, err = commandsFromFile(v); err != nil {
				jww.ERROR.Println(err.Error())
				return nil, nil, err
			} 

			// Assume it's the first element.
			settings[k] = commands[0]

		}
	}
	return settings, nil, nil
}

// rawTemplate.updateBuilders updates the rawTemplate's builders with the
// passed new builder.
// 
// Builder Update rules:
// 	* If r's old builder does not have a matching builder in the new
// 	  builder map, new, nothing is done.
//	* If the builder exists in both r and new, the new builder updates r's
//	  builder.
//	* If the new builder does not have a matching builder in r, the new
//	  builder is added to r's builder map.
// 
// Settings update rules:
//
//	* If the setting exists in r's builder but not in new, nothing is done.
//	  This means that deletion of settings via not having them exist in the
//	  new builder is not supported. This is to simplify overriding
//	  templates in the configuration files.
//	* If the setting exists in both r's builder and new, r's builder is 
//	  updated with new's value.
//	* If the setting exists in new, but not r's builder, new's setting is
//	  added to r's builder.
//	* To unset a setting, specify the key, without a value:
//	      `"key="`
//	  In most situations, Rancher will interprete an key without a value as
//	  a deletion of that key. There are exceptions:
//
//	  	* `guest_os_type`: This is generally set at Packer Template 
//		  generation time by Rancher.	
func (r *rawTemplate) updateBuilders(new map[string]*builder) {
	// If there is nothing new, old equals merged.
	if len(new) <= 0 || new == nil {
		return
	}
	jww.TRACE.Println("rawTemplate.updateBuilders-new:\n" + json.MarshalIndentToString(new, "", indent))

	// Convert to an interface.
	var ifaceOld map[string]interface{} = make(map[string]interface{}, len(r.Builders))
	for i, o := range r.Builders {
		ifaceOld[i] = o
	}
	// Convert to an interface.
	var ifaceNew map[string]interface{} = make(map[string]interface{}, len(new))
	for i, n := range new {
		ifaceNew[i] = n
	}

	// Get all the keys from map.
	var keys []string
	keys = keysFromMaps(ifaceOld, ifaceNew)

	bM := map[string]builder{}
	var vm_settings []string

	// If there's a builder with the key BuilderCommon, merge them
	if _, ok := new[BuilderCommon]; ok {
		r.updateBuilderCommon(new[BuilderCommon])	
	}

	jww.TRACE.Println("rawTemplate.updateBuilders-postCommon:\t" + json.MarshalIndentToString(r.Builders, "", indent))

	for _, v := range keys {
		b := &builder{}
		b = r.Builders[v]

		// If the element for this key doesn't exist, skip it.
		if _, ok := new[v]; !ok {
			continue
		}

		vm_settings = interfaceToStringSlice(new[v].Arrays[VMSettings])
		// If there is anything to merge, do so

		if vm_settings != nil {
			jww.TRACE.Println("rawTemplate.updateBuilders-build-preMerge:\t" + json.MarshalIndentToString(b, "", indent))			
			b.mergeVMSettings(vm_settings)
			jww.TRACE.Println("rawTemplate.updateBuilders-build-preMerge:\t" + json.MarshalIndentToString(b, "", indent))			
			bM[v] = *b
		}

	}

	jww.TRACE.Println("rawTemplate.updateBuilders-r.Builders-postMerge:\t" + json.MarshalIndentToString(r.Builders, "", indent))
	return 
}

// r.updateBuilderCommonSettings updates rawTemplate's BuilderCommon settings
// Update rules:
//	* When both the existing BuilderCommon, r, and the new one, b, have the
//	  same setting, b's value replaces r's; the new setting value replaces
//        the existing.
//	* When the setting in b is new, it is added to r: new settings are 
//	  inserted into r's BuilderCommon setting list.
//	* When r has a setting that does not exist in b, nothing is done. This
//	  method does not delete any settings that already exist in R.
func (r *rawTemplate) updateBuilderCommon(new *builder) {
	// If the existing builder doesn't have a BuilderCommon section, just add it
	if _, ok := r.Builders[BuilderCommon]; !ok {
		r.Builders[BuilderCommon] = new
		return
	}

	// Otherwise merge the two
	r.Builders[BuilderCommon].mergeSettings(new.Settings)

	return
}