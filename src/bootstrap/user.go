package bootstrap

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
	"math/big"
	"os"

	"github.com/syndtr/goleveldb/leveldb"
	"golang.org/x/crypto/bcrypt"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `json:"username" gorm:"uniqueIndex,not null"`
	Password string `json:"password" gorm:"not null"`
	Role     string `json:"role" gorm:"not null"`
	Email    string `json:"email"`
}

func MigrateLevelDBToSqlite(oldDBFile, newDBFile string) {
	oldDB, err := leveldb.OpenFile(oldDBFile, nil)
	if err != nil {
		log.Printf("Error opening old leveldb: %s", err)
		panic(err)
	}
	defer oldDB.Close()

	newDB, err := gorm.Open(sqlite.Open(newDBFile), nil)
	if err != nil {
		log.Printf("Error open sqlite and gorm: %s", err)
		panic(err)
	}
	defer func() {
		db, err2 := newDB.DB()
		if err2 != nil {
			log.Printf("Error getting real DB from gorm: %s", err2)
		}
		if db != nil {
			err2 = db.Close()
			if err2 != nil {
				log.Printf("Error closing real DB of gorm: %s", err2)
				panic(err2)
			}
		}
	}()

	err = newDB.AutoMigrate(&User{})
	if err != nil {
		log.Printf("Error autoMigrating sqlite database with user: %s", err)
		panic(err)
	}

	oldUserData, err := oldDB.Get([]byte("httpauth::userdata"), nil)
	if err != nil {
		log.Printf("Error getting `httpauth::userdata` from leveldb: %s", err)
		panic(err)
	}

	var migrationData map[string]struct {
		Username string
		Email    string
		Hash     string
		Role     string
	}
	err = json.Unmarshal(oldUserData, &migrationData)
	if err != nil {
		log.Printf("Error unmarshalling old user data: %s", err)
		panic(err)
	}

	for _, datum := range migrationData {
		// check if password is "factorio", which was the default password in the old system
		decodedHash, err := base64.StdEncoding.DecodeString(datum.Hash)
		if err != nil {
			log.Printf("Error decoding base64 hash: %s", err)
			panic(err)
		}

		err = bcrypt.CompareHashAndPassword(decodedHash, []byte("factorio"))
		if err == nil {
			// password is "factorio" .. change it
			newPassword := GenerateRandomPassword()

			bcryptPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
			if err != nil {
				log.Printf("Error generating has from password: %s", err)
				panic(err)
			}

			datum.Hash = base64.StdEncoding.EncodeToString(bcryptPassword)

			log.Println(`Migrated user in database. It still had default password "factorio" set. New credentials:`)
			log.Printf("Username: %s", datum.Username)
			log.Printf("Password: %s", newPassword)
		}

		user := &User{
			Username: datum.Username,
			Password: datum.Hash,
			Role:     datum.Role,
			Email:    datum.Email,
		}

		newDB.Create(user)
	}

	oldDB.Close()

	// delete oldDB
	log.Println("Deleting old leveldb database.")
	err = os.RemoveAll(oldDBFile)
	if err != nil {
		log.Printf("Error removing leveldb: %s", err)
		panic(err)
	}
}

// randLetters includes alphanumeric characters (excluding confusing ones like 0, O, l, 1, I)
// and special characters for stronger passwords
var randLetters = []rune("abcdefghjkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789!@#$%^&*()-_=+")

// GenerateRandomPassword generates a cryptographically secure random password
func GenerateRandomPassword() string {
	pass := make([]rune, 24)
	letterLen := big.NewInt(int64(len(randLetters)))
	for i := range pass {
		n, err := rand.Int(rand.Reader, letterLen)
		if err != nil {
			// Fall back to a less secure method only if crypto/rand fails
			log.Printf("Warning: crypto/rand failed, using fallback: %s", err)
			n = big.NewInt(int64(i % len(randLetters)))
		}
		pass[i] = randLetters[n.Int64()]
	}
	return string(pass)
}
