package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/afero"

	"github.com/oprekable/checksum-CRC32/logic"
)

func main() {
	var fileCSV string
	fileCSVUsage := fmt.Sprintf("Usage\t: %s -csv \"list_file.csv\"\n", os.Args[0])
	flag.StringVar(&fileCSV, "csv", "", fileCSVUsage)

	var fileImage string
	fileImageUsage := fmt.Sprintf("Usage\t: %s -f \"image.jpg\"\n", os.Args[0])

	flag.StringVar(&fileImage, "f", "", fileImageUsage)
	flag.Parse()

	if (fileCSV == "" && fileImage == "") || (fileCSV != "" && fileImage != "") {
		flag.Usage()
		os.Exit(1)
	}

	appfs := afero.NewOsFs()

	if fileCSV != "" {
		fileCSVHandler(fileCSV, appfs)
	}

	if fileImage != "" {
		fileImageHandler(fileImage, appfs)
	}
}

func fileCSVHandler(fileCSV string, fs afero.Fs) {
	mapData, err := logic.CheckSumCRC32FromFileCSVPath(fileCSV, fs)
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}

	for k, v := range mapData {
		fmt.Printf("%v\t%v", k, v)
	}
}

func fileImageHandler(fileImage string, fs afero.Fs) {
	fileImage, err := filepath.Abs(fileImage)
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}

	crc32, err := logic.CheckSumCRC32FromFilePath(fileImage, fs)
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}
	fmt.Printf("%v\t%v", fileImage, crc32)
}
