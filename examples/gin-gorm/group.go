package main

type group struct {
	ID      uint
	Name    string
	Members []member
}

type member struct {
	ID      uint
	GroupID uint
	Name    string
}
