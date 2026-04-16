# [Project Name]

## Introduction

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

envdash is a REST API service designed to provide airquality and general information about countries.

## Authors

This code was developed by:
TODO
* [Bror Wetlesen Vedeld] [@[BroVed]]([profile-link])
* [Your Name] [@[username]]([profile-link])
* [Your Name] [@[username]]([profile-link])

## Features

* User registration
* Provides API keys
* Allows registration and generation of Dashboard configurations
* Multilayer cache 
* Wrapper that can access external APIs

## Tech Stack
TODO Add links
 - Go 
 - Restcountries
 - OpenAq
 - Openmeteo
 - Restcurrencies
 - Firestore

## Project structure 
This section outlines the base structure of the project, details about the implementation of these components are in the docs folder: [docs](./docs).
In the docs folder we will also discuss tradeoffs and the like, if you are evaluating the project for a grade, you should definitely  check it out. 
```
assignment-2/
├── .gitignore
├── .gitmessage
├── go.mod
├── go.sum
├── LICENSE
├── README.md
├── requests.log
├── .idea/
│   ├── .gitignore
│   ├── assignment-2.iml
│   ├── modules.xml
│   ├── vcs.xml
│   └── workspace.xml
├── cmd/
│   └── envdash/
│       ├── main.go
│       └── requests.log
├── docs/
│   ├── Authentication.md
│   ├── GitHygiene.md
│   ├── SecretHandling.md
│   └── devutils/
│       ├── gitmessage.txt
│       └── setup-commit-template.sh
└── internal/
    ├── client/
    │   ├── currency/
    │   │   ├── client.go
    │   │   ├── client_flaky_test.go
    │   │   └── client_test.go
    │   ├── openaq/
    │   │   ├── client.go
    │   │   ├── client_flaky_test.go
    │   │   └── client_test.go
    │   ├── openmeteo/
    │   │   ├── client.go
    │   │   ├── client_flaky_test.go
    │   │   └── client_test.go
    │   └── restcountries/
    │       ├── client.go
    │       ├── client_flaky_test.go
    │       └── client_test.go
    ├── handlers/
    │   ├── authentication.go
    │   ├── dashboard.go
    │   ├── defaultHandler.go
    │   ├── handler.go
    │   ├── middleware.go
    │   ├── notification.go
    │   ├── registration.go
    │   ├── status.go
    │   ├── status_http_test.go
    │   ├── status_main_test.go
    │   └── status_stub_test.go
    ├── models/
    │   ├── authentication.go
    │   ├── dashboard.go
    │   ├── errorresponse.go
    │   ├── notification.go
    │   └── registration.go
    ├── store/
    │   ├── cache.go
    │   ├── cache_http_test.go
    │   ├── cache_main_test.go
    │   ├── cache_stub_test.go
    │   └── firestore.go
    └── utils/
        ├── constants.go
        ├── http.go
        └── logger.go
```

### Cmd

cmd contains the `main.go` file for running the project.

### Internal

internal contains most of the files used in the project. These are files that will not be accessible outside the Go module developed in the project. This means that if this module is included in another `go.mod`, the internal packages will not be exposed.

#### Client

Client contains all the Application Programming Interface (API) wrappers for external APIs, as separate packages with their own tests. These are only touched by the local cache and the status endpoint. More details are available in the docs folder: [client.md](./docs/client.md).

#### Handlers
TODO
Handlers contain the endpoints, middleware, and tests for the endpoint behavior.

#### Models

Models contain modular files with shared data structures used across multiple files.

#### Store

Store contains the local cache as well as the logic required to connect to Firestore. More details are available in the docs folder: [cache.md](./docs/cache.md).

#### Utils

Utils contains the Hypertext Transfer Protocol (HTTP) client factory as well as the logic for the logger.


## API Implementation

* **Language:** Go

### Deployment
TODO
Project is hosted on NTNU Openstack: [NTNU Openstack](http://10.212.172.108:8080/)

Must be connected to NTNU Internal Network to access.

- **Platform:** OpenStack
- **Containerization:** Docker Compose
    - **Description:** Services are containerized using Docker Compose to facilitate easy deployment and scaling.

## API Reference / Documentation

<details>
<summary><h4>Register as a user to receive an API key:</h4></summary>

```http
[METHOD] /path/to/endpoint
```

| Parameter / Header | Type     | Description                          |
| :----------------- | :------- | :----------------------------------- |
| `[name]`           | `[type]` | **Required/Optional**. [Description] |

#### Example Request Body:
TODO
```json
{
  "key": "value"
}
```

#### Response:

| Status Code | Content-Type       |
| :---------- | :----------------- |
| `200 OK`    | `application/json` |

```json
{
  "message": "example response"
}
```

</details>

<details>
<summary><h4>Check API Statuses: (Firestore, independent third party API, Version, Uptime)</h4></summary>
TODO
```http
  GET /dashboards/v1/status?token={token}
```

| Parameter | Type     | Description                |
|:----------|:---------|:---------------------------|
| `token`   | `string` | **Required**. Your API key |

#### Response:

| Status Code  | Content-Type       |
|:-------------|:-------------------|
| `200 OK`     | `application/json` |

```json
{
    "CountriesAPI": "Status of the REST Countries API",
    "MeteoAPI": "Status of the Open-Meteo API",
    "OpenAQAPI": "Status of the OpenAq API",
    "CurrencyAPI": "Status of the REST Currency API",
    "NotificationDB": "Status of the Notification database",
    "webhooks": "Number of webhooks registered",
    "version": "API Version",
    "uptime": "Time since last server reboot (In Seconds)"
}
```

</details>

details>
<summary><h4>Register a Country to get information for:</h4></summary>
TODO
```http
[METHOD] /path/to/endpoint
```

| Parameter / Header | Type     | Description                          |
| :----------------- | :------- | :----------------------------------- |
| `[name]`           | `[type]` | **Required/Optional**. [Description] |

#### Example Request Body:

```json
{
  "field": "value"
}
```

#### Response:

| Status Code   | Content-Type       |
| :------------ | :----------------- |
| `201 Created` | `application/json` |

```json
{
  "id": "created-id",
  "timestamp": "ISO8601 timestamp"
}
```

</details>

<details>
<summary><h4>Retrieve all registered countries:</h4></summary>
TODO
```http
[METHOD] /path/to/endpoint
```

#### Response:

| Status Code | Content-Type       |
| :---------- | :----------------- |
| `200 OK`    | `application/json` |

```json
[
  {
    "id": "item-1"
  }
]
```

</details>

<details>
<summary><h4>[Endpoint purpose 5]</h4></summary>

```http
[METHOD] /path/to/endpoint/{id}
```

#### Response:

| Status Code | Content-Type       |
| :---------- | :----------------- |
| `200 OK`    | `application/json` |

```json
{
  "id": "item-id",
  "name": "example"
}
```

</details>

<details>
<summary><h4>[Endpoint purpose 6]</h4></summary>

```http
[METHOD] /path/to/endpoint/{id}
```

#### Example Request Body:

```json
{
  "fieldToUpdate": "newValue"
}
```

#### Response:

| Status Code    | Content-Type       |
| :------------- | :----------------- |
| `202 Accepted` | `application/json` |

```json
{
  "lastChange": "Updated timestamp"
}
```

</details>

<details>
<summary><h4>[Endpoint purpose 7]</h4></summary>

```http
[METHOD] /path/to/endpoint/{id}
```

#### Response:

| Status Code      |
| :--------------- |
| `204 No Content` |

</details>

## Environment Variables

To run this project, you will need to add the following environment variables to your `.env` file, or project environment.

`PORT` - Port to run the project on. This defaults to 8080
`[FIREBASE_CREDENTIALS_FILE]` - path to firebase credentials file
`OPENAQ_API_KEY` - A string containing the openAQ_API_KEY

## Run Locally

* Clone the repository

```bash
git clone https://git.gvk.idi.ntnu.no/course/prog2005/prog2005-2026-workspace/broved/assignment-2.git
```

* Navigate to the project directory:

```bash
cd ./cmd/envdash
```

### Run using Go:
  ```bash
  go run ./cmd/envdash/app.go
  ```
    - From Build:
  ```bash
  go build -a -o app ./cmd/envdash
  ```
  ```bash
  ./app
  ```

* ### Run using Docker:
TODO
```bash
docker compose build
```

* #### Attached:
TODO
```bash
docker compose up [service-name]
```

* #### Detached:
TODO
```bash
docker compose up [service-name] -d
```

* #### View Logs:
TODO
```bash
docker compose logs [service-name]
```

* #### Follow Logs:
TODO
```bash
docker compose logs [service-name] -f
```

* #### Stop Services:
TODO
```bash
docker compose down [service-name]
```

## Running Tests
TODO
To run tests, navigate to the project directory:

```bash
cd [project-folder]
```

## Running Tests

In our project we do not run tests that require third party APIs by default.

You can run tests that use third party apis by adding a flaky build tag.

Note: For the clients if you run the tests without access to network(or by not using the flaky build tag) the coverage drops by about 60%. This is due to the fact that most of the code relies on requesting the external service. 

To run tests, navigate to the project directory:

```bash
cd globeboard/Go/
```
* ### Run tests using Go:

  ```bash
  go test ./...
  ```

  * With flaky tests (tests that call external APIs):

  ```bash
  go test -tags=flaky ./...
  ```

  * With Coverage using Go: (Full Project)

  ```bash
  go test -cover -coverpkg=./... ./...
  ```

  * With Coverage using Go and flaky tests:

  ```bash
  go test -tags=flaky -cover -coverpkg=./... ./...
  ```

* ### Run tests in Docker:
TODO
```bash
docker compose build
```

* #### Attached:
TODO i dont understand this.
```bash
docker compose up [test-service-name]
```

* #### Detached:

```bash
docker compose up [test-service-name] -d
```

* #### View Logs:

```bash
docker compose logs [test-service-name]
```

* #### Follow Logs:

```bash
docker compose logs [test-service-name] -f
```

* #### Stop Services:

```bash
docker compose down [test-service-name]
```

## Roadmap

* Implement Firestore caching
    - the architecture supports this, we just need to implement it.
* Optimize the openAQ api call
    - currently this works by calling on a country code and filtering through the results.
      this can often lead to 30+ calls, while the client is ratelimited so we dont disrupt external services, a better call shoould be found

## Support

For support, contact: `brorwv@stud.ntnu.no`

## License

This project is licensed under the MIT License.

You are free to use, copy, modify, merge, publish, distribute, sublicense, and sell copies of the software, as permitted by the license. The software is provided “as is”, without warranty of any kind.

