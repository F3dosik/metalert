package main

// generate:reset
type User struct {
	ID       int
	Name     string
	Email    string
	Active   bool
	Score    float64
	Tags     []string
	Metadata map[string]string
	Address  *Address
	Profile  Profile
}

// generate:reset
type Address struct {
	Street  string
	City    string
	Country string
	Zip     *string
	Coords  []float64
}

// Profile не помечена — Reset() для неё не генерируется
type Profile struct {
	Bio    string
	Avatar string
}
