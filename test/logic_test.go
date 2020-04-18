package test

import (
	"bytes"
	"crypto/rand"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"gotest.tools/assert"

	"github.com/oprekable/checksum-CRC32/logic"
	"github.com/oprekable/checksum-CRC32/test/helper"
)

func init() {
	GenBlob()
}

type errReader int

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("test error")
}

func generateRandomBytes(n int) (returnData []byte, err error) {
	returnData = make([]byte, n)
	_, err = rand.Read(returnData)
	if err != nil {
		return
	}

	return
}

func generateRandomString(n int) (returnData string, err error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	bytes, err := generateRandomBytes(n)
	if err != nil {
		returnData = ""
		return
	}
	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}
	returnData = string(bytes)
	return
}

func setFile(fileName string, nFile int)(returnFs afero.Fs, returnFileContent string, err error){
	returnFs = afero.NewMemMapFs()
	file, err := returnFs.Create(fileName)

	if err != nil {
		return
	}

	defer func() {
		_ = file.Close()
	}()

	if file == nil {
		err = errors.New("failed create file")
		return
	}

	var fileContent strings.Builder

	n := 1
	for n <= nFile {
		fileName, err := generateRandomString(10)

		if err != nil {
			continue
		}

		fileName += ".jpg"
		fileNameAbs, _ := filepath.Abs(fileName)

		fileContent.WriteString(fileNameAbs)
		fileContent.WriteString("\n")

		fileImage, err := returnFs.Create(fileNameAbs)

		if err != nil {
			continue
		}
		_, _ = fileImage.WriteString("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-")
		fileImage.Close()
		n++
	}

	if nFile > 0 {
		fileContent.WriteString("not_exists_file.jpg")
		fileContent.WriteString("\n")
	}

	_, _ = file.WriteString(fileContent.String())

	returnFileContent = fileContent.String()
	return
}

func TestReadFileSuccess(t *testing.T) {
	fileName := "file_list.csv"
	appFS, fileContent, err := setFile(fileName,5)

	if err != nil {
		t.Error(err)
	}

	fileFromRead, err := logic.ReadFile(fileName, appFS)

	if err != nil {
		t.Error(err)
	}

	fileSetupStringArray := logic.ReaderToStringArray(strings.NewReader(fileContent))
	fileFromReadStringArray := logic.ReaderToStringArray(fileFromRead)


	assert.DeepEqual(t, fileSetupStringArray, fileFromReadStringArray)
}

func TestReadFileError(t *testing.T) {
	fileName := "file_list.csv"
	fs := afero.NewMemMapFs()
	_, err := logic.ReadFile(fileName, fs)

	assert.Error(t, err, "open " + fileName +": file does not exist")

}

func TestFileToPathAbsArraySuccess(t *testing.T) {
	fileName := "file_list.csv"
	appFS, fileContent, err := setFile(fileName, 5)

	if err != nil {
		t.Error(err)
	}

	fileFromReadPathAbsArray, err := logic.FileToPathAbsArray(fileName, appFS)

	if err != nil {
		t.Error(err)
	}

	fileSetupStringArray := logic.ReaderToStringArray(strings.NewReader(fileContent))
	fileSetupPathAbsArray := logic.StringArrayToPathAbs(fileSetupStringArray)


	assert.DeepEqual(t, fileSetupPathAbsArray, fileFromReadPathAbsArray)
}

func TestFileToPathAbsArrayError(t *testing.T) {
	fileName := "file_list.csv"
	fs := afero.NewMemMapFs()
	_, err := logic.FileToPathAbsArray(fileName, fs)

	assert.Error(t, err, "open " + fileName +": file does not exist")
}

func TestReaderToBase64Success(t *testing.T) {
	imageFileInByte := helper.Box.Get("/arsya.jpg")
	imageFileInReader := bytes.NewReader(imageFileInByte)
	b64,err := logic.ReaderToBase64(imageFileInReader)

	if err != nil {
		t.Error(err)
	}
	
	imageFileBase64InByte := string(helper.Box.Get("/arsya_base64.txt"))
	assert.Equal(t, imageFileBase64InByte, b64)
}

func TestReaderToBase64Error(t *testing.T) {
	_,err := logic.ReaderToBase64(errReader(0))
	assert.Error(t, err, "test error")
}

func TestCheckSumCRC32FromReaderSuccess(t *testing.T) {
	imageFileInByte := helper.Box.Get("/arsya.jpg")
	imageFileInReader := bytes.NewReader(imageFileInByte)
	checkSum, err := logic.CheckSumCRC32FromReader(imageFileInReader)

	if err != nil {
		t.Error(err)
	}

	expectedCheckSum := uint32(1352499051)

	assert.Equal(t, expectedCheckSum, checkSum)
}

func TestCheckSumCRC32FromReaderError(t *testing.T) {
	_, err := logic.CheckSumCRC32FromReader(errReader(0))
	assert.Error(t, err, "test error")
}

func TestCheckSumCRC32FromFilePathSuccess(t *testing.T) {
	fileName := "file_list.csv"
	appFS, _, err := setFile(fileName, 5)

	if err != nil {
		t.Error(err)
	}

	fileFromReadPathAbsArray, err := logic.FileToPathAbsArray(fileName, appFS)

	if err != nil {
		t.Error(err)
	}

	expectedCheckSum := uint32(3182332081)

	for _, s := range fileFromReadPathAbsArray {
		crc32, err := logic.CheckSumCRC32FromFilePath(s, appFS)

		if err != nil {
			if err.Error() == "open " + s + ": file does not exist" {
				continue
			}

			t.Error(err)
		}

		assert.Equal(t, expectedCheckSum, crc32)
	}
}

func TestCheckSumCRC32FromFilePathError(t *testing.T) {
	fileName := "image.png"
	fs := afero.NewMemMapFs()

	_, err := logic.CheckSumCRC32FromFilePath(fileName, fs)
	assert.Error(t, err, "open " + fileName +": file does not exist")
}

func TestCheckSumCRC32FromFileCSVPathSuccess(t *testing.T) {
	fileName := "file_list.csv"
	appFS, _, err := setFile(fileName, 5)

	if err != nil {
		t.Error(err)
	}

	dataMap, err := logic.CheckSumCRC32FromFileCSVPath(fileName, appFS)

	if err != nil {
		t.Error(err)
	}

	expectedCheckSum := uint32(3182332081)

	for _, v := range dataMap {
		assert.Equal(t, expectedCheckSum, v)
	}
}

func TestCheckSumCRC32FromFileCSVPathError(t *testing.T) {
	fileName := "file_list.csv"
	fs := afero.NewMemMapFs()

	_, err := logic.CheckSumCRC32FromFileCSVPath(fileName, fs)
	assert.Error(t, err, "open " + fileName +": file does not exist")
}

func TestCheckSumCRC32FromFileCSVPathError2(t *testing.T) {
	fileName := "file_list.csv"
	appFS, _, err := setFile(fileName, 0)

	if err != nil {
		t.Error(err)
	}

	_, err = logic.CheckSumCRC32FromFileCSVPath(fileName, appFS)
	assert.Error(t, err, "empty data")
}

