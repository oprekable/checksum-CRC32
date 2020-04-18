package logic

import (
	"bufio"
	"encoding/base64"
	"errors"
	"hash/crc32"
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/spf13/afero"
)

func ReadFile(fileName string, fs afero.Fs) (returnData afero.File, err error){
	returnData, err = fs.Open(fileName)
	if err != nil {
		return
	}

	return
}

func ReaderToStringArray(handle io.Reader) (returnData []string) {
	scanner := bufio.NewScanner(handle)
	for scanner.Scan() {
		d := scanner.Text()
		if d != "" {
			returnData = append(returnData, scanner.Text())
		}
	}
	return
}

func ReaderToBase64(handle io.Reader) (returnData string, err error) {
	content, err := ioutil.ReadAll(handle)
	if err != nil {
		return
	}
	returnData = base64.StdEncoding.EncodeToString(content)
	return
}

func StringArrayToPathAbs(handle []string) (returnData []string) {
	for _, v := range handle {
		f, err := filepath.Abs(v)
		if err == nil {
			returnData = append(returnData, f)
		}
	}
	return
}

func FileToPathAbsArray(fileName string, fs afero.Fs) (returnData []string, err error) {
	file, err := ReadFile(fileName, fs)
	if err != nil {
		return
	}

	defer func() {
		_ = file.Close()
	}()

	s := ReaderToStringArray(file)
	returnData = StringArrayToPathAbs(s)
	return
}

func CheckSumCRC32FromReader(handle io.Reader) (returnData uint32, err error){
	b64, err := ReaderToBase64(handle)
	if err != nil {
		return
	}

	// abp := *(*[]byte)(unsafe.Pointer(&b64))
	abp := []byte(b64)
	table := crc32.MakeTable(crc32.IEEE)
	returnData = crc32.Checksum(abp, table)
	return
}

func CheckSumCRC32FromFilePath(fileName string, fs afero.Fs) (returnData uint32, err error){
	file, err := ReadFile(fileName, fs)
	if err != nil {
		return
	}

	return CheckSumCRC32FromReader(file)
}

func CheckSumCRC32FromFileCSVPath(fileCSVName string, fs afero.Fs) (returnData map[string]uint32, err error){
	fileFromReadPathAbsArray, err := FileToPathAbsArray(fileCSVName, fs)
	if err != nil {
		return
	}

	if len(fileFromReadPathAbsArray) == 0 {
		err = errors.New("empty data")
		return
	}

	m := make(map[string]uint32, len(fileFromReadPathAbsArray))

	for _, s := range fileFromReadPathAbsArray {
		c, err := CheckSumCRC32FromFilePath(s, fs)
		if err != nil {
			continue
		}

		m[s] = c
	}

	returnData = m

	return
}
