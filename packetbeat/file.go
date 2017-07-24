package packetbeat

import (
	"io/ioutil"
	"log"
	"github.com/go-yaml/yaml"
	"reflect"
	"os"
	"fmt"
	"github.com/docker/docker/api/types"
	"path/filepath"
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

func (file *ConfigFile) WritePort(data []int) {
	p := file.GetYaml()
	p["packetbeat.protocols.http"].(map[interface{}]interface{})["ports"] = data
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
	log.Println("Data has been written to " + file.Path)
}

func (file *ConfigFile) GetYaml() map[interface{}]interface{} {
	if file.Yaml != nil {
		return file.Yaml
	} else {
		return file.Read()
	}
}

func (file *ConfigFile) GetPortList(protocol string) []int {
	p := file.GetYaml()
	ports := p["packetbeat.protocols."+protocol].(map[interface{}]interface{})["ports"]
	values := reflect.ValueOf(ports)
	size := values.Len()
	result := make([]int, size)
	for i := 0; i < size; i++ {
		result[i] = values.Index(i).Interface().(int)
	}
	return result
}


func (file *ConfigFile) UpdatePort(container types.Container, actualPortList []int) []int {
	for _, p := range container.Ports {
		if shouldBeAppend(int(p.PrivatePort), actualPortList) {
			actualPortList = append(actualPortList, int(p.PrivatePort))
		}
	}
	return actualPortList
}

func shouldBeAppend(p int, ports []int) bool {
	for _, port := range ports {
		if p == port {
			return false
		}
	}
	return true
}
