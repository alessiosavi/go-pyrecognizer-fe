package redis

import (
	"fmt"
	"github.com/alicebob/miniredis/v2"
	"testing"
)

var mock *miniredis.Miniredis

func init() {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	mock = s
}

func TestConnectToDb(t *testing.T) {
	if _, err := Connect(mock.Host(), mock.Port(), 0); err != nil {
		t.Error(err)
	}
}

func TestInsertIntoClient(t *testing.T) {
	client, err := Connect(mock.Host(), mock.Port(), 0)
	if err != nil {
		t.Error(err)
	}
	if err = Insert(client, "test", "test", 0); err != nil {
		t.Error(err)
	}
	data, err := mock.Get("test")
	if err != err {
		t.Error(err)
	}
	if data != "test" {
		t.Error(fmt.Sprintf("expected %s found %s", "test", data))
	}
}

func TestGet(t *testing.T) {
	client, err := Connect(mock.Host(), mock.Port(), 0)
	if err != nil {
		t.Error(err)
	}

	if err = mock.Set("test", "test"); err != nil {
		t.Error(err)
	}
	data, err := Get(client, "test")
	if err != nil {
		t.Error(err)
	}
	if data != "test" {
		t.Error(fmt.Sprintf("expected %s found %s", "test", data))
	}
}

func TestRemove(t *testing.T) {
	client, err := Connect(mock.Host(), mock.Port(), 0)
	if err != nil {
		t.Error(err)
	}

	if err = mock.Set("test", "test"); err != nil {
		t.Error(err)
	}

	if err = Remove(client, "test"); err != nil {
		t.Error(err)
	}

	if mock.Exists("test") {
		t.Error("test was not deleted")
	}

}
