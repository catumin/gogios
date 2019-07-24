package main

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestAppendStringToFile(t *testing.T) {
	emptyFile, err := os.Create("./testa")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(emptyFile)
	emptyFile.Close()

	err = AppendStringToFile("./testa", "String")
	if err != nil {
		t.Errorf("String writing failed, got error: %s", err)
	}

	content, err := ioutil.ReadFile("./testa")
	if err != nil {
		t.Errorf("String writing failed, got error: %s", err)
	}

	if string(content) != "\nString" {
		t.Errorf("String writing failed, got: %s, want: %s.", content, "String")
	}
}

func TestCopy(t *testing.T) {
	err := Copy("./testa", "./testb")
	if err != nil {
		t.Errorf("Copy function failed, got error: %s", err)
	}

	content, err := ioutil.ReadFile("./testb")
	if err != nil {
		t.Errorf("Copy process failed, got error: %s", err)
	}

	if string(content) != "\nString" {
		t.Errorf("Result of copy is wrong, got: %s, want: %s.", content, "String")
	}

	os.Remove("./testa")
	os.Remove("./testb")
}
