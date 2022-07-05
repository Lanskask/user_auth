# user_auth Example with Postgres and BCrypt

## Test requests 

```
POST: 
http://localhost:5000/auth/register

Body:
{
    "email": "peter@gmail.com",
    "name": "peter",
    "password": "pass1"
}
```

```
POST
http://localhost:5000/auth/login

Body:
{
    "email": "peter@gmail.com",
    "password": "pass1"
}
```

```
POST
http://localhost:5000/auth/logout

Body:
{
    "email": "peter@gmail.com",
    "password": "pass1"
}
```

```
POST
http://localhost:5000/auth/healthcheck
```

```
GET
http://localhost:5000/user
```

## Create a postgres table

```sql
create table users (
    id serial not null unique,
    name varchar(64) not null, 
    email varchar(64) not null unique,
    password text not null,
    primary key (id)
);
```

## How to connect to postgres DB in Docker

```shell
docker ps # to get container id
docker exec -it  <container id> /bin/bash
psql -u postgres
```
