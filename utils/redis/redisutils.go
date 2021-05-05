package redis

import (
	"context"
	"errors"
	"github.com/alessiosavi/GoGPUtils/helper"
	stringutils "github.com/alessiosavi/GoGPUtils/string"
	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
	"time"
)

// Connect use emtpy string for hardcoded port
func Connect(addr string, port string, db int) (*redis.Client, error) {
	if stringutils.IsBlank(addr) {
		addr = "localhost"
	}
	if stringutils.IsBlank(port) {
		port = "6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr + ":" + port,
		Password: "",
		DB:       db,
	})
	log.Info("Connecting to: " + helper.MarshalIndent(client.String()))
	err := client.Ping(context.TODO()).Err()
	if err != nil {
		log.Error("Impossibile to connecto to DB ....| CLIENT: ", addr, ":", port, " | ERR: ", err)
		return nil, err
	}
	return client, nil
}

// Get is delegated to check if a key is already inserted and return the value
func Get(client *redis.Client, key string) (string, error) {
	tmp, err := client.Get(context.TODO(), key).Result()
	if err == nil {
		log.Debug("SUCCESS | Key: ", key, " | Value: ", tmp)
		return tmp, err
	} else if err == redis.Nil {
		log.Warn("Key -> ", key, " does not exist")
		return "", errors.New("keys does not exists")
	}
	log.Error("Fatal exception during retrieving of data [", key, "] | Redis: ", client, " Error: ", err)
	return "", err
}

// Remove is delegated to check if a key is alredy inserted and return the value
func Remove(client *redis.Client, key string) error {
	err := client.Del(context.TODO(), key).Err()
	if err == nil {
		log.Debug("SUCCESS | Key: ", key, " | Removed")
		return nil
	} else if err == redis.Nil {
		log.Warn("Remove | Key -> ", key, " does not exist")
		return nil
	}
	log.Error("Fatal exception during retrieving of data [", key, "] | Redis: ", client)
	return err
}

// Insert set the two value into the Databased pointed from the client
func Insert(client *redis.Client, key string, value string, expire int) error {
	log.Info("Inserting -> (", key, ":", value, ")")
	err := client.Set(context.TODO(), key, value, time.Second*time.Duration(expire)).Err() // Inserting the values into the DB
	if err != nil {
		return err
	}
	log.Info("INSERTED SUCCESFULLY!! | (", key, ":", value, ")")
	return nil
}
