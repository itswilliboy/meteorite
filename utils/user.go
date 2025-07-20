package utils

import "context"

type User struct {
	Id   int
	Name string
}

func GetUserByID(id int) User {
	var name string
	DB.QueryRow(context.Background(), "SELECT name FROM users WHERE id = $1", id).Scan(&name)
	return User{Id: id, Name: name}
}
