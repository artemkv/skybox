package main

import (
	"fmt"
	"log"
)

func main() {
	// load .env
	LoadDotEnv()

	// read config
	folder := GetMandatoryString("SKYBOX_FOLDER")
	bucket := GetMandatoryString("SKYBOX_BUCKET")
	deviceId := GetMandatoryString("SKYBOX_DEVICEID")
	secret := GetMandatoryString("SKYBOX_SECRET")

	// get all encryption related stuff
	// TODO: should be random and be stored on the account
	salt := []byte("saltsaltsaltsalt")
	masterKey, err := DeriveMasterKey(secret, salt)
	if err != nil {
		log.Fatal(err)
	}

	// backup
	fmt.Println("Starting backup")
	objects, err := Backup(folder, bucket, deviceId, masterKey)
	if err != nil {
		fmt.Println("Backup failed")
		log.Fatal(err)
	}
	fmt.Println("Backup completed")

	// report
	fmt.Println("Objects failed to backup:")
	for _, obj := range objects {
		if obj.Error != nil {
			fmt.Printf("'%s': %v\n", obj.Path, obj.Error)
		}
	}
}

/*
func main() {
	// 32-byte symmetric key
	key := []byte("this-is-a-secret-32-byte-key-!!!")

	input, err := os.Open("example.pdf")
	output, err := os.Create("encrypted.dat")

	nonce, err := Encrypt(input, output, key)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Encryption error: %v\n", err)
		return
	}
	output.Close()

	_ = os.WriteFile("nonce.dat", nonce, 0644)

	savedNonce, _ := os.ReadFile("nonce.dat")

	cipherFile, _ := os.Open("encrypted.dat")
	defer cipherFile.Close()

	decFile, _ := os.Create("restored.pdf")
	defer decFile.Close()

	_ = Decrypt(cipherFile, decFile, key, savedNonce)
}
*/
