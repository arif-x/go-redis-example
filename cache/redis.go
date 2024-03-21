package cache

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/google/uuid"
)

type Person struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type PersonService interface {
	GetPerson(id string) (*Person, error)
	GetPersons(page, pageSize int) ([]*Person, error)
	CreatePerson(person *Person) (*Person, error)
	UpdatePerson(person *Person) (*Person, error)
	DeletePerson(id string) error
}

type redisCache struct {
	host string
	db   int
	exp  time.Duration
}

func NewRedisCache(host string, db int, exp time.Duration) PersonService {
	return &redisCache{
		host: host,
		db:   db,
		exp:  exp,
	}
}

func (cache redisCache) getClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cache.host,
		Password: "",
		DB:       cache.db,
	})
}

func (cache redisCache) CreatePerson(person *Person) (*Person, error) {
	c := cache.getClient()
	person.Id = uuid.New().String()
	json, err := json.Marshal(person)
	if err != nil {
		return nil, err
	}
	c.HSet("persons", person.Id, json)
	if err != nil {
		return nil, err
	}
	return person, nil
}

func (cache redisCache) GetPerson(id string) (*Person, error) {
	c := cache.getClient()
	val, err := c.HGet("persons", id).Result()

	if err != nil {
		return nil, err
	}
	person := &Person{}
	err = json.Unmarshal([]byte(val), person)

	if err != nil {
		return nil, err
	}
	return person, nil
}

func (cache redisCache) GetPersons(page int, pageSize int) ([]*Person, error) {
	c := cache.getClient()
	persons := []*Person{}

	startIndex := (page - 1) * pageSize
	endIndex := startIndex + pageSize - 1

	val, err := c.HGetAll("persons").Result()
	if err != nil {
		return nil, err
	}

	i := 0
	for _, item := range val {
		// Skip items until reaching the start index
		if i < startIndex {
			i++
			continue
		}

		// Stop iteration if reached the end index
		if i > endIndex {
			break
		}

		person := &Person{}
		err := json.Unmarshal([]byte(item), person)
		if err != nil {
			return nil, err
		}
		persons = append(persons, person)
		i++
	}

	return persons, nil
}

func (cache redisCache) UpdatePerson(person *Person) (*Person, error) {
	c := cache.getClient()
	json, err := json.Marshal(&person)
	if err != nil {
		return nil, err
	}
	c.HSet("persons", person.Id, json)
	if err != nil {
		return nil, err
	}
	return person, nil
}
func (cache redisCache) DeletePerson(id string) error {
	c := cache.getClient()
	numDeleted, err := c.HDel("persons", id).Result()
	if numDeleted == 0 {
		return errors.New("person to delete not found")
	}
	if err != nil {
		return err
	}
	return nil
}
