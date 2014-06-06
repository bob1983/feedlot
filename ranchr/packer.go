package ranchr


import (
	_"bufio"
	"encoding/json"
	"errors"
	_"fmt"
	"io"
	"os"
	_"reflect"
	_"strings"
	_"time"
)

type packerer interface {
	mergeSettings([]string)
}

type builderer interface {
	mergeVMSettings([]string)
}

type PackerTemplate struct {
	Description      string                 `json:"description"`
	MinPackerVersion string                 `json:"min_packer_version"`
	Builders         []interface{}          `json:"builders"`
	PostProcessors   []interface{}          `json:"post-processors"`
	Provisioners     []interface{}          `json:"provisioners"`
	Variables        map[string]interface{} `json:"variables"`
}

func (p *PackerTemplate) TemplateToFileJSON(i IODirInf, b BuildInf) error {

	if i.OutDir == "" {
		err := errors.New("ranchr.TemplateToFileJSON: output directory for " + b.BuildName + " not set")
		Log.Error(err.Error())
		return err
	}

	if i.SrcDir == "" {
		err := errors.New("ranchr.TemplateToFileJSON: SrcDir directory for " + b.BuildName + " not set")
		Log.Error(err.Error())
		return err
	}

	if i.ScriptsDir == "" {
		err := errors.New("ranchr.TemplateToFileJSON: ScriptsDir directory for " + b.BuildName + " not set")
		Log.Error(err.Error())
		return err
	}

	// Create the output directory stuff
	if err := os.MkdirAll(i.OutDir + "/scripts", os.FileMode(0744)); err != nil {
		Log.Error(err.Error())
		return err
	}

	if err := os.MkdirAll(i.OutDir + "/http", os.FileMode(0744)); err != nil {
		Log.Error(err.Error())
		return err
	}

	// Write it out as JSON
	tplJSON, err := json.MarshalIndent(p, "", "\t")
	if err != nil {
		Log.Error("Marshalling of the Packer Template failed: " + err.Error())
		return err
	}
	
	f, err := os.Create(i.OutDir + b.Name)
	if err != nil {
		Log.Error(err.Error())
		return err
	}
	defer f.Close()
	
	_, err = io.WriteString(f,  string(tplJSON[:]))
	if err != nil {
		Log.Error(err.Error())
		return err
	}

	return nil
}
