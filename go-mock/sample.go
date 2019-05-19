package main

import "fmt"

type DB interface {
	Get(key string) error
}


type MyDB struct {}

func (m *MyDB) Get(key string) error {
	fmt.Println("I'm MyDB.")
	return nil
}


type TestDB struct {}

func (t *TestDB) Get(key string) error {
	fmt.Println("I'm TestDB.")
	return nil
}


type Server struct {
	DB DB
}

func (s *Server) Start() {
	s.DB.Get("aaa")
}
