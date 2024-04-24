package main 

import (
  "os"
  "errors"
  "sync"
  "time"
  "encoding/json"
) 

type DB struct {
	path string
	mu  *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users map[int]User `json:"users"`
}

type User struct {
  Hash string `json:"hash"`
  Email string `json:"email"`
  ID int `json:"id"`
  AccessToken string `json:"access_token"`
  RefreshToken string `json:"refresh_token"`
  RefreshTokenRevokedAt string `json:"refresh_token_revoked_at"`
  AccessTokenRevokedAt string `json:"access_token_revoked_at"`
}

type Chirp struct {
  Body string `json:"body"`
  ID int `json:"id"`
  Author_ID int `json:"author_id"`
}

// NewDB creates a new database connection
func NewDB(path string) (*DB, error) {
	db := &DB{
		path: path,
		mu:   &sync.RWMutex{},
	}
	err := db.ensureDB()
	return db, err
}

// if the database file does not exist, create it 
func (db *DB) ensureDB() error {
	_, err := os.ReadFile(db.path)
	if errors.Is(err, os.ErrNotExist) {
		return db.createStructureDB()
	}
	return err
}

// this creates the structure of the database to write to file 
func (db *DB) createStructureDB() error {
	dbStructure := DBStructure{
		Chirps: map[int]Chirp{},
		Users: map[int]User{},
	}
	return db.writeDB(dbStructure)
}

func (db *DB) writeDB(dbStructure DBStructure) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	dat, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}

	err = os.WriteFile(db.path, dat, 0600)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) loadDB() (DBStructure, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	dbStructure := DBStructure{}
	dat, err := os.ReadFile(db.path)
	if errors.Is(err, os.ErrNotExist) {
		return dbStructure, err
	}
	err = json.Unmarshal(dat, &dbStructure)
	if err != nil {
		return dbStructure, err
	}

	return dbStructure, nil
}

func (db *DB) CreateChirp(body string, user User) (Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	id := len(dbStructure.Chirps) + 1
	chirp := Chirp{
		ID:   id,
		Body: body,
    Author_ID: user.ID,
	}
	dbStructure.Chirps[id] = chirp

	err = db.writeDB(dbStructure)
	if err != nil {
		return Chirp{}, err
	}

	return chirp, nil
}

func (db *DB) GetChirps() ([]Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	chirps := make([]Chirp, 0, len(dbStructure.Chirps))
	for _, chirp := range dbStructure.Chirps {
		chirps = append(chirps, chirp)
	}

	return chirps, nil
}

func (db *DB) DeleteChrip (chirp Chirp) (error) {
  dbStructure, err := db.loadDB()
  if err != nil {
    return err
  }

  // we are deleting the key of ID from the map
  delete(dbStructure.Chirps, chirp.ID)

  errW := db.writeDB(dbStructure)
	if errW != nil {
		return errW
	}

  return nil
}

func (db *DB) CreateUser(email string, password string) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	id := len(dbStructure.Users) + 1
  hash, _ := HashPassword(password)

	user := User{
		ID:   id,
		Email: email,
    Hash: hash,
	}
	dbStructure.Users[id] = user

	err = db.writeDB(dbStructure)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (db *DB) UpdateUser(id int, email, hashedPassword string) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	user, ok := dbStructure.Users[id]
	if !ok {
		return User{}, errors.New("already exists")
	}

	user.Email = email
	user.Hash = hashedPassword
	dbStructure.Users[id] = user

	err = db.writeDB(dbStructure)
	if err != nil {
		return User{}, err
	}

	return user, nil
}
func (db *DB) SetUserTokens(id int, accessToken, refreshToken string) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	user, ok := dbStructure.Users[id]
	if !ok {
		return User{}, errors.New("already exists")
	}

	user.AccessToken = accessToken
	user.RefreshToken = refreshToken
	dbStructure.Users[id] = user

	err = db.writeDB(dbStructure)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (db *DB) FindUserByRefreshToken (refreshToken string) (User, error) {
	dbUsers, err := db.GetUsers()
	if err != nil {
    return User{}, err
	}
  
	for _, dbUser := range dbUsers {
    if dbUser.RefreshToken == refreshToken {
      return dbUser, nil
    }
	}

  return User{}, nil
}
func (db *DB) FindUserByAccessToken (accessToken string) (User, error) {
	dbUsers, err := db.GetUsers()
	if err != nil {
    return User{}, err
	}
  
	for _, dbUser := range dbUsers {
    if dbUser.AccessToken == accessToken {
      return dbUser, nil
    }
	}

  return User{}, nil
}
func (db *DB) RevokeRefreshToken(refreshToken string) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	user, err := db.FindUserByRefreshToken(refreshToken)
	if err != nil {
		return User{}, err
	}

	user.RefreshTokenRevokedAt = time.Now().String()
	dbStructure.Users[user.ID] = user

	err = db.writeDB(dbStructure)
	if err != nil {
		return User{}, err
	}

	return user, nil
}
func (db *DB) GetUsers() ([]User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	users := make([]User, 0, len(dbStructure.Users))
	for _, user := range dbStructure.Users {
		users = append(users, user)
	}

	return users, nil
}
