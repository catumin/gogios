package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

// AppendStringToFile - Add string to the bottom of a file
func AppendStringToFile(path, text string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString("\n" + text)
	if err != nil {
		return err
	}
	return nil
}

// WriteStringToFile - Replace contents of the file
func WriteStringToFile(path, text string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		CreateFile(path)
	}

	err := ioutil.WriteFile(path, []byte(text), 0644)
	if err != nil {
		return err
	}

	return nil
}

// Copy makes a replica of a file in a new location
func Copy(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

// CreateFile makes a blank file at the specified path
func CreateFile(dst string) error {
	emptyFile, err := os.Create(dst)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(emptyFile)
	emptyFile.Close()

	return err
}

// ReadFile returns the cotents of a file as a string
func ReadFile(src string) (string, error) {
	content, err := ioutil.ReadFile(src)
	if err != nil {
		log.Fatal(err)
	}

	return string(content), err
}
