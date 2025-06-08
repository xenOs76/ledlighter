package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"testing"
)

var testFiles = map[string]string{
	"wrong_addr":      "testdata/config_test_wrong_addr.yaml",
	"wrong_kind":      "testdata/config_test_wrong_kind.yaml",
	"wrong_yaml":      "testdata/config_test_wrong_yaml.yaml",
	"wrong_yaml2":     "testdata/config_test_wrong_yaml2.yaml",
	"incomplete_yaml": "testdata/config_test_wrong_yaml3.yaml",
	// "wrong_id":        "testdata/config_test_wrong_id.yaml",
}

var dest_test_config_file = "testdata/config.yaml"

func createTestConfigFile(t *testing.T, source_test_config_file string, dest_test_config_file string) (err error) {
	source_file, err := os.Open(source_test_config_file)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer source_file.Close()

	dest_file, err := os.Create(dest_test_config_file)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dest_file.Close()

	if _, err := io.Copy(bufio.NewWriter(dest_file), bufio.NewReader(source_file)); err != nil {
		log.Fatal(err)
	}

	t.Cleanup(func() {
		os.Remove(dest_test_config_file)
	})

	return nil
}

func TestLoadConfig(t *testing.T) {
	for k, v := range testFiles {
		err := createTestConfigFile(t, v, dest_test_config_file)
		if err != nil {
			t.Error(err)
		}
		fmt.Printf("\t ...loading config file for test %v\n", k)

		cfg, errCfg := LoadConfig()
		if errCfg != nil {
			t.Error(errCfg)
		}

		_, errLedMap := GetLedsMap(cfg)
		if errLedMap == nil {
			t.Errorf("error not detected while importing LedsMap from config. Test %v", k)
		} else {
			fmt.Print("\t\t error correctly detected:\n")
			fmt.Printf("\t\t %v\n", errLedMap)
		}
	}
}
