In the case that docker-compose up does not work, please run each docker containers manually
before start running the containers, please create a new docker network called 'my-net'
create network by using bash: $docker network create my-net
Start With: sql -> gosql -> goweb
1. HOW TO START SQL:
    move to the sql directory, then run //$ docker build -t sqlbuild .
    then run //$ docker run -p 3306:3306 --name sql1 --network=my-net -d sqlbuild

2. HOW TO START GOSQL (This initiates the DB):
    first off, WAIT for about 10 seconds before running the script as it takes time for the SQL container to bind its port
    then move to the directory 'gosql'
    then run //$ docker build -t sqlinitiate .
    then run //$ docker run -p 8080:8080 --network=my-net sqlinitiate

3. Finally, start the Go script that handles all the SQL query and manage the web
    move to the directory 'goweb'
    then run //$ docker build -t goweb .
    then run //$ docker run -p 8080:8080 --network=my-net goweb

Once the website is running, visit it by inputting the localhost address port 8080 in your browser.
Example starting address: http://localhost:8080/addCurr/form