package ini

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// manage all the sections and their key values defined in the .ini file
//
type Ini struct {
	defaultSectionName string
	sections           map[string]*Section
}

func NewIni() *Ini {
	return &Ini{defaultSectionName: "default",
		sections: make(map[string]*Section)}
}

func (ini *Ini) GetDefaultSectionName() string {
	return ini.defaultSectionName
}

func (ini *Ini) SetDefaultSectionName(defSectionName string) {
	ini.defaultSectionName = defSectionName
}

// create a new section if the section with name does not exist
// or return the exist one if the section with name already exists
//
func (ini *Ini) NewSection(name string) *Section {
	if section, ok := ini.sections[name]; ok {
		return section
	}
	section := NewSection(name)
	ini.sections[name] = section
	return section
}

// add a section to the .ini file and overwrite the exist section
// with same name
func (ini *Ini) AddSection(section *Section) {
	ini.sections[section.Name] = section
}

// Get all the section name in the ini
//
// return all the section names
func (ini *Ini) Sections() []*Section {
	r := make([]*Section, 0)
	for _, section := range ini.sections {
		r = append(r, section)
	}
	return r
}

// check if a key exists or not in the Ini
//
// return true if the key in section exists
func (ini *Ini) HasKey(sectionName, key string) bool {
	if section, ok := ini.sections[sectionName]; ok {
		return section.HasKey(key)
	}
	return false
}

// get section by section name
//
// return: section or nil
func (ini *Ini) GetSection(name string) (*Section, error) {
	if section, ok := ini.sections[name]; ok {
		return section, nil
	}
	return nil, noSuchSection(name)
}

// return true if the section with name exists
// return false if the section with name does not exist
func (ini *Ini) HasSection(name string) bool {
	_, err := ini.GetSection(name)
	return err == nil
}

// get the value of key in section
func (ini *Ini) GetValue(sectionName, key string) (string, error) {
	if section, ok := ini.sections[sectionName]; ok {
		return section.GetValue(key)
	}
	return "", noSuchSection(sectionName)
}

// get the value of the key in section
// if the key does not exist, return the defValue
func (ini *Ini) GetValueWithDefault(sectionName, key string, defValue string) string {
	if section, ok := ini.sections[sectionName]; ok {
		return section.GetValueWithDefault(key, defValue)
	}
	return defValue
}

// get the value of key in section as bool.
// return true if the value of the key is one of following(case insensitive):
//  - true
//  - yes
//  - t
//  - y
//  - 1
// return false for all other values
func (ini *Ini) GetBool(sectionName, key string) (bool, error) {
	if section, ok := ini.sections[sectionName]; ok {
		return section.GetBool(key)
	}
	return false, noSuchSection(sectionName)
}

// get the value of key as bool and return the default value if the section in the .ini file
// or key in the section does not exist
func (ini *Ini) GetBoolWithDefault(sectionName, key string, defValue bool) bool {
	if section, ok := ini.sections[sectionName]; ok {
		return section.GetBoolWithDefault(key, defValue)
	}
	return defValue
}

// get the value of key in the section as int
func (ini *Ini) GetInt(sectionName, key string) (int, error) {
	if section, ok := ini.sections[sectionName]; ok {
		return section.GetInt(key)
	}
	return 0, noSuchSection(sectionName)
}

// get the value of key in the section as int and return defValue if the section in the .ini file
// or key in the section does not exist
func (ini *Ini) GetIntWithDefault(sectionName, key string, defValue int) int {
	if section, ok := ini.sections[sectionName]; ok {
		return section.GetIntWithDefault(key, defValue)
	}
	return defValue
}

// get the value of key in the section as uint
func (ini *Ini) GetUint(sectionName, key string) (uint, error) {
	if section, ok := ini.sections[sectionName]; ok {
		return section.GetUint(key)
	}
	return 0, noSuchSection(sectionName)
}

// get the value of key in the section as int and return defValue if the section in the .ini file
// or key in the section does not exist
func (ini *Ini) GetUintWithDefault(sectionName, key string, defValue uint) uint {
	if section, ok := ini.sections[sectionName]; ok {
		return section.GetUintWithDefault(key, defValue)
	}
	return defValue
}

// get the value of key in the section as int64
func (ini *Ini) GetInt64(sectionName, key string) (int64, error) {
	if section, ok := ini.sections[sectionName]; ok {
		return section.GetInt64(key)
	}
	return 0, noSuchSection(sectionName)
}

// get the value of key in the section as int64 and return defValue if the section in the .ini file
// or key in the section does not exist
func (ini *Ini) GetInt64WithDefault(sectionName, key string, defValue int64) int64 {
	if section, ok := ini.sections[sectionName]; ok {
		return section.GetInt64WithDefault(key, defValue)
	}
	return defValue
}

// get the value of key in the section as uint64
func (ini *Ini) GetUint64(sectionName, key string) (uint64, error) {
	if section, ok := ini.sections[sectionName]; ok {
		return section.GetUint64(key)
	}
	return 0, noSuchSection(sectionName)
}

// get the value of key in the section as uint64 and return defValue if the section in the .ini file
// or key in the section does not exist
func (ini *Ini) GetUint64WithDefault(sectionName, key string, defValue uint64) uint64 {
	if section, ok := ini.sections[sectionName]; ok {
		return section.GetUint64WithDefault(key, defValue)
	}
	return defValue
}

// get the value of key in the section as float32
func (ini *Ini) GetFloat32(sectionName, key string) (float32, error) {
	if section, ok := ini.sections[sectionName]; ok {
		return section.GetFloat32(key)
	}
	return 0, noSuchSection(sectionName)
}

// get the value of key in the section as float32 and return defValue if the section in the .ini file
// or key in the section does not exist
func (ini *Ini) GetFloat32WithDefault(sectionName, key string, defValue float32) float32 {
	if section, ok := ini.sections[sectionName]; ok {
		return section.GetFloat32WithDefault(key, defValue)
	}
	return defValue
}

// get the value of key in the section as float64
func (ini *Ini) GetFloat64(sectionName, key string) (float64, error) {
	if section, ok := ini.sections[sectionName]; ok {
		return section.GetFloat64(key)
	}
	return 0, noSuchSection(sectionName)
}

// get the value of key in the section as float64 and return defValue if the section in the .ini file
// or key in the section does not exist
func (ini *Ini) GetFloat64WithDefault(sectionName, key string, defValue float64) float64 {
	if section, ok := ini.sections[sectionName]; ok {
		return section.GetFloat64WithDefault(key, defValue)
	}
	return defValue
}

func noSuchSection(sectionName string) error {
	return fmt.Errorf("no such section:%s", sectionName)
}

func (ini *Ini) String() string {
	buf := bytes.NewBuffer(make([]byte, 0))
	ini.Write(buf)
	return buf.String()
}

// write the content of the .ini in the .ini file format, e.g. in following format:
//
//  [section1]
//  key1 = value1
//  key2 = value2
//  [section2]
//  key3 = value3
//  key4 = value4
func (ini *Ini) Write(writer io.Writer) error {
	for _, section := range ini.sections {
		err := section.Write(writer)
		if err != nil {
			return err
		}
	}
	return nil
}

// Write the conents of ini to a file
func (ini *Ini) WriteToFile(fileName string) error {
	file, err := os.Create(fileName)
	if err == nil {
		defer file.Close()
		return ini.Write(file)
	}
	return err
}
