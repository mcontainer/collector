package packetbeat

import (
	"io/ioutil"
	log "github.com/sirupsen/logrus"
	"github.com/go-yaml/yaml"
	"reflect"
	"os"
	"fmt"
	"path/filepath"
	"github.com/Workiva/go-datastructures/bitarray"
)

type ConfigFile struct {
	Path string
	Yaml map[interface{}]interface{}
}

func NewConfigFile(path string) *ConfigFile {
	p, e := filepath.Abs(path)
	if e != nil {
		log.Fatal(e)
	}

	if _, err := os.Stat(p); os.IsNotExist(err) {
		log.Fatal(err)
		os.Exit(1)
	}

	return &ConfigFile{
		Path: p,
	}
}

func (file *ConfigFile) Read() map[interface{}]interface{} {
	bytes, err := ioutil.ReadFile(file.Path)
	if err != nil {
		log.Fatal(err)
	}
	p := make(map[interface{}]interface{})
	yaml.Unmarshal(bytes, &p)
	return p
}

func (file *ConfigFile) WritePort(data bitarray.BitArray) {
	p := file.GetYaml()
	p["packetbeat.protocols.http"].(map[interface{}]interface{})["ports"] = data.ToNums()
	out, _ := yaml.Marshal(&p)
	f, e := os.Create(file.Path)

	if e != nil {
		log.Fatal(e)
	}

	defer f.Close()
	error := ioutil.WriteFile(f.Name(), out, os.ModePerm)
	if error != nil {
		fmt.Println(error)
	}

	log.WithField("Path", file.Path).Info("Data has been written to config file")
}

func (file *ConfigFile) GetYaml() map[interface{}]interface{} {
	if file.Yaml != nil {
		return file.Yaml
	} else {
		return file.Read()
	}
}

func (file *ConfigFile) GetPortList(protocol string) bitarray.BitArray {
	ba := bitarray.NewBitArray(65535)
	p := file.GetYaml()
	ports := p["packetbeat.protocols."+protocol].(map[interface{}]interface{})["ports"]
	if ports != nil {
		values := reflect.ValueOf(ports)
		size := values.Len()
		for i := 0; i < size; i++ {
			ba.SetBit(uint64(values.Index(i).Interface().(int)))
		}
	}
	return ba
}
