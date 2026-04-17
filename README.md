# [Project Name]

## Introduction

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

envdash is a REST API service designed to provide airquality and general information about countries.

## Authors

This code was developed by:
TODO
* [Bror Wetlesen Vedeld] [@[BroVed]]([profile-link])
* [Lennart Krogh] [@[Lennart]]([profile-link])
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
<summary> <h2> Acquire Api Key </h2> </summary>

Simply **POST** your name and email in JSON format to `/envdash/v1/auth/`

Example URL:
`POST xxxxx:8080/envdash/v1/auth/`
Body:
``` json
{
  "name": "Alice",
  "email": "alice@mail.com"
}
```
| Parameter      | Type  | Description                       |
|:-----------|:------------|:----------------------------    |
| `name`     | `string`    | *Required*                      |
| `email`    | `string`    | *Required*. Need to contain "@" |



#### Response:

| Status Code | Content-Type       |
| :---------- | :----------------- |
| `201 Created`    | `application/json` |


You will then receive an API key:
``` json
{
  "key": "sk-envdash-YourAPIkey...",
  "createdAt": "20260317 20:32"
}
```

| Fields      | Description                 |
|:----------- |:----------------------------|
| `key`       | Your personal API key  |
| `createdAt` | When the API key was created     |


</details>


<details>
<summary> <h2> Delete Api Key </h2> </summary>

Simply **DELETE** your api key using api you want to delete in url `/envdash/v1/auth/{apiKey}` 

Example URL:
`DELETE xxxxx:8080/envdash/v1/auth/sk-envdash-YourAPIkey`

| Header      | Value: Type | Description                 |
|:-----------|:------------|:----------------------------|
| `x-api-key` | {YourAPIkey}       | Needs to be an api key from the same user. You have to be allowed to delete the key. You can delete your own key asswell |


#### Response:

| Status Code   |
|:--------------|
| `204 No Content` |

When you get the 204, you know that the api key is deleted.
If you receive any other status code, the API key was not deleted.
You will get a helpfull error message, try using that to understand
why the key cant be deleted.

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

---

<details>
<summary><h2> Create Notification </h2></summary>

Simply **POST** your notification query with correct body `/envdash/v1/notifications/`
 [Remember to have Api Key in header] 

```http
POST /envdash/v1/notifications/
```
Body:
We have two type of notificaitons, lifecycle and threshold:
````json
{
   "url":     "https://webhook.site/your-unique-URL",
   "country": "NO",
   "event":   "INVOKE"
}
````
````json
{
   "url":     "https://webhook.site/your-unique-URL",
   "country": "NO",
   "event":   "THRESHOLD",
   "threshold": {
      "field":    "pm25",
      "operator": ">",
      "value":    35.0
   }
}
````


#### Parameters:
For a detailed look at the parameters take a look here: [Parameters](docs/Notifications.md#parameters)


#### Response:

| Status Code      |
| :--------------- |
| `201 Created` |
Body:
````json
{
ìd:   IdOfYourNotifcation
}
````
You have now created a notfication, it wil send a webhook when the
event you registerd is fulfilled

</details>

--- 

<details>
<summary><h2> Retrieve a Specific Webhook  </h2></summary>

Simply **GET** your notifcation by `/envdash/v1/notifications/{ID_Of_Notifcation}`

Example URL:
`GET xxxxx:8080/envdash/v1/notifications/uUmLcayWY9WgGL26ASDp`

-Body empty-

#### Response:

Header:

| Status Code | Content-Type       |
| :---------- | :----------------- |
| `200 OK`    | `application/json` |


Body:
``` json
{
    "id": "ID_Of_Notifcation",
    "url": "https://webhook.site/your-unique-URL",
    "country": "NO",
    "event": "THRESHOLD",
    "threshold": {
        "field": "PM25",
        "operator": "==",
        "value": 35
    }
}
```

### Parameters
For a detailed look at the parameters take a look here: [Parameters](docs/Notifications.md#parameters)


considerations: everyone can read your webhook even non owners of the particular webhooks as long as they know the ID

</details>

---

<details>
<summary> <h2> List All Your Registered Webhooks </h2> </summary>

Simply **GET** `/envdash/v1/notifications/`

Example URL:
`GET xxxxx:8080/envdash/v1/notifications/`

-Body empty-

#### Response:

| Status Code | Content-Type       |
| :---------- | :----------------- |
| `200 OK`    | `application/json` |


You wil now see every notification registerd to your account
``` json
{
[
    {
        "id": "ID_Of_Notifcation",
        "url": "https://webhook.site/your-unique-URL",
        "country": "NO",
        "event": "CHANGE"
    },
    {
        "id": "Another_ID_Of_Notifcation",
        "url": "https://webhook.site/your-unique-URL",
        "country": "NO",
        "event": "DELETE"
    }
]
}
```

### Parameters
For a detailed look at the parameters take a look here: [Parameters](docs/Notifications.md#parameters)|

</details>

---

<details>

<summary> <h2> Delete Notification </h2> </summary>

Simply **DELETE** your notification `/envdash/v1/notifications/{NotificationID}`

Example URL:
`DELETE xxxxx:8080/envdash/v1/notifications/6pSNoPNL08oroGqRWoAR`
-Body empty-

#### Response:

| Status Code | Content-Type       |
| :---------- | :----------------- |
| `204 No Content`    | `application/json` |

You have now deleted your notification.
Remember, you can not delete someone leses notification
This is desided on your api key you use in header.

</details>

---




---

s

---

s

---

| Fields      | Description                 |
|:----------- |:----------------------------|
| `key`       | Your personal API key  |
| `createdAt` | When the API key was created     |



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
git clone https://github.com/lenny113/Cloud.git
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
    - Currently this works by calling on a country code and using the first page of the result.
      this is not an accurate measurement. But it was needed to get theuntryinfo in one api call.

## Support

For support, contact: `brorwv@stud.ntnu.no`

## License

This project is licensed under the MIT License.

You are free to use, copy, modify, merge, publish, distribute, sublicense, and sell copies of the software, as permitted by the license. The software is provided “as is”, without warranty of any kind.

