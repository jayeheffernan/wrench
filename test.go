package main

import (
	"github.com/nightrune/wrench/logging"
	"os"
	"path/filepath"
	"strings"
)

const TEST_FILE_PATTERN = "*.test.nut"

var cmdTest = &Command{
	UsageLine: "test",
	Short:     "Runs all tests it can find",
	Long:      "Recursively looks for files with the .test.nut type within the directory",
}

func init() {
	cmdTest.Run = TestMe
}

func FindTestFiles() ([]string, error) {
	var files []string
	err := filepath.Walk("./", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			location := path + "\\" + TEST_FILE_PATTERN
			matches, err := filepath.Glob(location)
			if err != nil {
				return err
			}
			if len(matches) > 0 {
				logging.Debug("Found these files: %s", matches)
				files = append(files, matches...)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

func TestMe(cmd *Command, args []string) {
	logging.Info("Attempting to find test scripts...")
	test_files, err := FindTestFiles()
	if err != nil {
		os.Exit(1)
	}

	for _, input_file := range test_files {
		split_file := strings.Split(input_file, ".")
		// Remove test, and add .o
		split_file = append(split_file[:len(split_file)-2], split_file[len(split_file)-1], "o")
		output_file := strings.Join(split_file, ".")
		logging.Info("Processing target: %s as output: %s", input_file, output_file)
		err = PreProcessFile(output_file, input_file, cmd.settings.LibraryDirs)
		if err != nil {
			logging.Warn("Could not processes output file %s, got error: ", output_file, err.Error())
			continue
		}
		logging.Info("Running target: %s", output_file)
		err = ExecuteSqrl(output_file)
		if err != nil {
			logging.Warn("Could not run output_file %s, got error: ", output_file, err.Error())
			continue
		}
	}

	logging.Debug("All Files Found: %s", test_files)
	logging.Info("Script finished")
}
