[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/mini-maxit/backend)

# Development

## Run

To run the appication you need running **docker** with **docker compose**

You also need to have local image of file-storage build and stored. The tag for image should be `maxit/file-storage`. Refer to [documentation](https://github.com/mini-maxit/file-storage?tab=readme-ov-file#build) on how to do it.

```bash
docker compose up --build -d
```

## You might need it

### DUMP=true

If you set `DUMP=true` in the .env file, and start application using provided dev docker compose the app will start with some values present in database. For now this includes:

- Admin user. Email=`admin@admin.com` Password=`adminadmin`
- Teacher user. Email=`teacher@teacher.com` Password=`teacherteacher`
- Student user. Email=`student@student.com` Password=`studentstudent`

# Endpoints

You can check automaticaly generated markdown documentation [here](./docs/swagger.md). Documentation is generated from swagger specification, which in turn is generated from code comments.

If you prefer swagger specification you can access it [here](https://mini-maxit.github.io/backend/)
