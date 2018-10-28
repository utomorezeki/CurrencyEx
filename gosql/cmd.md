docker build -t hello-world .
docker run -p 8080:8080 hello-world

docker network create -d bridge my-net

docker run --name sql1 -p 3306:3306 -e MYSQL_ROOT_PASSWORD=pass --network=my-net -d mysql/mysql-server:8.0

docker build -t sqlbuild .
docker run -p 3306:3306 --name sql1 --network=my-net -d sqlbuild

docker run -p 8080:8080 --network=my-net sqlinitiate

	insert, err := db.Query("INSERT INTO Testt VALUES('333','111');")

	defer insert.Close()


	if err != nil {
		panic(err.Error())
	}

    	db, err := sql.Open("mysql","root:pass@tcp(sql1:3306)/")

	if err != nil {
		panic(err)
	}