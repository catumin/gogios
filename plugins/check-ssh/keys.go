package main

import (
	"io/ioutil"
	"log"

	"golang.org/x/crypto/ssh"
)

// KeyFile - Use non encrypted private key to auth
func KeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		log.Println("Failed to open key file, error: ", err)
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		log.Println("Failed to parse key file, error: ", err)
	}

	return ssh.PublicKeys(key)
}

// EncryptedKeyFile - Use encrypted private key to auth
func EncryptedKeyFile(file, password string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		log.Println("Failed to open key file, error: ", err)
	}

	key, err := ssh.ParsePrivateKeyWithPassphrase(buffer, []byte(password))
	if err != nil {
		log.Println("Failed to parse key file, error: ", err)
	}

	return ssh.PublicKeys(key)
}
