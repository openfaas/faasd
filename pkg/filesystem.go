package pkg

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
)

func MakeFile(filePath, fileContents string) error {
	_, err := os.Stat(filePath)
	if err == nil {
		log.Printf("File exists: %q\n", filePath)
		return nil
	} else if os.IsNotExist(err) {
		log.Printf("Writing to: %q\n", filePath)
		return ioutil.WriteFile(filePath, []byte(fileContents), 0644)
	} else {
		return err
	}
}

func EnsureWorkingDir(folder string) error {
	if _, err := os.Stat(folder); err != nil {
		err = os.MkdirAll(folder, 0600)
		if err != nil {
			return err
		}
	}

	return nil
}

func CopyFile(source, destFolder string) error {
	file, err := os.Open(source)
	if err != nil {
		return err

	}
	defer file.Close()

	out, err := os.Create(path.Join(destFolder, source))
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, file)

	return err
}

func BinaryExists(folder, name string) error {
	findPath := path.Join(folder, name)
	if _, err := os.Stat(findPath); err != nil {
		return fmt.Errorf("unable to stat %s, install this binary before continuing", findPath)
	}
	return nil
}
