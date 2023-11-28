# SWAPI challenge [WIP]

## Quick start

### Clone the repository and `cd` into it

```bash
git clone https://github.com/0xdeadbad/swapi-challenge
cd swapi-challenge
```

### Use `docker compose` to pull and build the images and lastly start the services

```bash
docker compose pull
docker compose build
docker compose up -d
```

## What is being deployed

The service fetches data from the [SWAPI](https://swapi.dev/) REST API and stores it locally in a MongoDB database. It also caches the data fetched into a Redis in-memory database for faster queries.  

In summary 3 containers are deployed, the MongoDB database, the Redis for caching and the swapi-challenge which is this API itself.
