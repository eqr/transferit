package auth

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"log"
	"strings"

	"github.com/boltdb/bolt"
	"golang.org/x/crypto/bcrypt"
)

var usersBucket = []byte("users")

type LoginInformation struct {
	Login    string
	Password string
	Salt     string
	Id       uint64
}

type LoginService interface {
	LoginUser(login, password string) (bool, uint64)
	CreateUser(login, password string) (uint64, error)
	ListUsers() ([]LoginInformation, error)
	DeleteUser(id uint64) error
	GetUserLogin(id uint64) (string, error)
}

type boltLoginService struct {
	db *bolt.DB
}

// CreateUser creates a new user and returns her id.
func (service *boltLoginService) CreateUser(login string, password string) (uint64, error) {
	var createdId uint64
	err := service.db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(usersBucket); err != nil {
			return fmt.Errorf("cannot validate users bucket: %w", err)
		}

		bucket := tx.Bucket(usersBucket)

		// check if such login alerady exists
		existing := bucket.Get([]byte(login))
		if len(existing) > 0 {
			return fmt.Errorf("user with login %s already exists", login)
		}

		id, err := bucket.NextSequence()
		if err != nil {
			return fmt.Errorf("cannot generate sequence for bucket 'users': %v", err.Error())
		}

		pass, salt := hashPassword(password)

		loginInformation := LoginInformation{
			Login:    login,
			Password: pass,
			Id:       id,
			Salt:     salt,
		}

		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		if err := enc.Encode(loginInformation); err != nil {
			return fmt.Errorf("error encoding user info: %v", err.Error())
		} else {
			storeId := getStoreId(id)

			// create users/1 record to store user information
			err = bucket.Put([]byte(storeId), buf.Bytes())
			if err != nil {
				return fmt.Errorf("error putting new user info to the db: %v", err.Error())
			}

			// create login - user/1 record to be able to login users
			err = bucket.Put([]byte(login), []byte(storeId))
			if err != nil {
				return fmt.Errorf("error putting new login to the db: %v", err.Error())
			}
		}

		createdId = id
		return nil
	})

	return createdId, err
}

func (service *boltLoginService) GetUserLogin(id uint64) (string, error) {
	var login string
	err := service.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(usersBucket)
		if b == nil {
			return nil
		}
		user, err := getLoginInformationFromBucket(id, b)
		if err != nil {
			return fmt.Errorf("cannot find user %d login information id db: %v", id, err.Error())
		}

		login = user.Login
		return nil
	})

	if err != nil {
		return "", err
	}

	return login, nil
}

// ListUsers returns the list of existing users
func (service *boltLoginService) ListUsers() ([]LoginInformation, error) {
	var users []LoginInformation
	err := service.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(usersBucket)
		if b == nil {
			return nil
		}
		return b.ForEach(func(k, v []byte) error {
			var user LoginInformation
			buf := bytes.NewBuffer(v)
			dec := gob.NewDecoder(buf)

			id := string(k)
			if !strings.Contains(id, "users/") {
				// ignore login records
				return nil
			}

			if decodeErr := dec.Decode(&user); decodeErr != nil {
				return fmt.Errorf("error reading login information: %v", decodeErr.Error())
			} else {
				users = append(users, user)
			}
			return nil
		})
	})

	if err != nil {
		return nil, fmt.Errorf("error reading  users from database: %v", err.Error())
	}

	return users, nil
}

// LoginUser returns true if authenticated and user id
func (service *boltLoginService) LoginUser(login string, password string) (bool, uint64) {
	var loggedUser LoginInformation
	var found bool
	err := service.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(usersBucket)
		if bucket == nil {
			return nil
		}

		loginRead := bucket.Get([]byte(login))
		if loginRead == nil {
			return nil
		}

		userRead := bucket.Get(loginRead)
		if userRead == nil {
			return nil
		}

		var user LoginInformation

		buf := bytes.NewBuffer(userRead)
		dec := gob.NewDecoder(buf)

		if decoderErr := dec.Decode(&user); decoderErr != nil {
			return fmt.Errorf("error decoding user info for login %v: %v", login, decoderErr.Error())
		}

		if comparePassword(password, user.Salt, user.Password) {
			loggedUser = user
			found = true
			return nil
		}

		return nil
	})

	if err != nil {
		log.Printf("error logging user %v: %v\n", login, err.Error())
		return false, 0
	}

	if found {
		return true, loggedUser.Id
	}

	return false, 0
}

func (service *boltLoginService) DeleteUser(id uint64) error {
	err := service.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(usersBucket)
		if bucket == nil {
			return nil
		}

		user, err := getLoginInformationFromBucket(id, bucket)
		if err != nil {
			return fmt.Errorf("cannot find user %d login information id db: %v", id, err.Error())
		}

		login := user.Login

		deleteErr := bucket.Delete([]byte(getStoreId(id)))
		if deleteErr != nil {
			return fmt.Errorf("error deleting user by users/%d: %v", id, deleteErr.Error())
		}

		deleteErr = bucket.Delete([]byte(login))
		if deleteErr != nil {
			return fmt.Errorf("error deleting login by %v: %v", login, deleteErr.Error())
		}

		return nil
	})

	return err
}

func NewLoginService(db *bolt.DB) LoginService {
	return &boltLoginService{
		db: db,
	}
}

func getLoginInformationFromBucket(id uint64, bucket *bolt.Bucket) (LoginInformation, error) {
	storeId := getStoreId(id)

	userRead := bucket.Get([]byte(storeId))
	if userRead == nil {
		return LoginInformation{}, fmt.Errorf("user not found: %v", storeId)
	}

	var user LoginInformation

	buf := bytes.NewBuffer(userRead)
	dec := gob.NewDecoder(buf)

	if decoderErr := dec.Decode(&user); decoderErr != nil {
		return LoginInformation{}, fmt.Errorf("error decoding user info for user %v: %v", storeId, decoderErr.Error())
	}

	return user, nil
}

func getStoreId(id uint64) string {
	return fmt.Sprintf("users/%d", id)
}

func hashPassword(password string) (string, string) {
	salt := generateSalt()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password+salt), bcrypt.DefaultCost)
	return string(hashedPassword), salt
}

func generateSalt() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func comparePassword(password string, salt string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password+salt))
	return err == nil
}
