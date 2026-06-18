package main

import (
	"fmt"
	"log"
	"os"
)

// TODO: I need proper modes with arguments
const (
	ModeUndefined = iota
	ModeBackup
	ModeRestore
)

// TODO: make this actually usable
func readArgs() int {
	args := os.Args[1:]
	if len(args) == 0 {
		return ModeBackup
	}

	if args[0] == "backup" {
		return ModeBackup
	}
	if args[0] == "restore" {
		return ModeRestore
	}

	log.Fatalln("Wrong arguments")
	return ModeUndefined
}

func main() {
	// detect mode
	mode := readArgs()

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

	if mode == ModeBackup {
		// backup
		fmt.Println("Starting backup")
		local, cloud, err := Backup(folder, bucket, deviceId, masterKey)
		if err != nil {
			fmt.Println("Backup failed")
			log.Fatal(err)
		}
		fmt.Println("Backup completed")

		// report
		fmt.Println("Objects failed to backup to cloud:")
		for _, obj := range local {
			if obj.Error != nil {
				fmt.Printf("'%s': %v\n", obj.Path, obj.Error)
			}
		}
		fmt.Println("Objects failed to remove in cloud:")
		for _, obj := range cloud {
			if obj.Error != nil {
				fmt.Printf("'%s': %v\n", obj.Path, obj.Error)
			}
		}

		return
	}

	if mode == ModeRestore {
		fmt.Println("Starting restore")
		err := Restore(folder, bucket, deviceId, masterKey)
		if err != nil {
			fmt.Println("Restore failed")
			log.Fatal(err)
		}
		fmt.Println("Restore completed")
		return
	}

	log.Fatalln("Unexpected mode")
}
