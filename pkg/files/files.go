package files

import (
	"encoding/json"
	"fmt"
	log4 "github.com/alecthomas/log4go"
	"io/ioutil"
	"os"
)

// LoadConfig
func LoadConfig(fileName string, ins interface{}) error {
	data, err := readFile(fileName)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, ins)
	if err != nil {
		return fmt.Errorf("json.Unmarshal TestConfig:%s error:%s", data, err)
	}
	return nil
}

// LoadParams
func LoadParams(dir, fileName string, data interface{}) error {
	fullPath := dir + string(os.PathSeparator) + fileName
	bz, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return err
	}
	return json.Unmarshal(bz, data)
}

func readFile(fileName string) ([]byte, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0666)
	if err != nil {
		return nil, fmt.Errorf("OpenFile %s error %s", fileName, err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log4.Error("File %s close error %s", fileName, err)
		}
	}()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("ioutil.ReadAll %s error %s", fileName, err)
	}
	return data, nil
}
