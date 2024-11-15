db-up:
	docker run -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=test-maxit -p 5432:5432 --name test_db -d --rm postgres
db-down:
	docker stop test_db
