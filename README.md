[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/mini-maxit/backend)  [![codecov](https://codecov.io/gh/mini-maxit/backend/graph/badge.svg?token=AS7XEYOXBB)](https://codecov.io/gh/mini-maxit/backend)
# Development

## Run

To run the appication you need running **docker** with **docker compose**

You also need to have local image of file-storage build and stored. The tag for image should be `maxit/file-storage`. Refer to [documentation](https://github.com/mini-maxit/file-storage?tab=readme-ov-file#build) on how to do it.

```bash
docker compose up --build -d
```

## Migrations

[Atlas](https://atlasgo.io/getting-started#installation) is used for database migrations. You can find migration files in `migrations` folder.

To generate new migration file you can use the following command:

```bash
atlas migrate diff --env gorm
```

To apply migrations to database you can use:

```bash
atlas migrate apply --env gorm --url <database_url>
```


## You might need it

### DUMP=true

If you set `DUMP=true` in the .env file, and start application using provided dev docker compose the app will start with some values present in database. For now this includes:

- Admin user. Email=`admin@admin.com` Password=`adminadmin`
- Teacher user. Email=`teacher@teacher.com` Password=`teacherteacher`
- Student user. Email=`student@student.com` Password=`studentstudent`

### Swagger docs
In order to generate swagger documentation you need to have installed `swag` tool. You can install it using:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

Next you can generate swagger documentation using:

```bash
make docs
```

You can access it at `/api/v1/docs` when the application is running.

### Mocks
In order to generate mocks for services and repositories, you need to have the `mockgen` tool installed. You can install it using:

```bash
go install go.uber.org/mock/mockgen@latest
make mocks
```


# Endpoints

You can check automaticaly generated swagger documentation [here](https://mini-maxit.github.io/backend/). Documentation is generated from code comments, so please excuse if there are mistakes.
