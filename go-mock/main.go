package main

func main() {
	myDB := &MyDB{}
	testDB:= &TestDB{}

	foo := &Server{
		DB: myDB,
	}
	bar := &Server{
		DB: testDB,
	}

	foo.Start()
	bar.Start()
}
