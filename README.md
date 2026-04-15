# README


** More coming soon .. **





# README V
Version 1


## Service Specification:
Our service have these endpoints:

/envdash/v1/auth/
/envdash/v1/registrations/
/envdash/v1/dashboards/
/envdash/v1/notifications/
/envdash/v1/status/

---

## Authentication

### Why authenticate?
You should authenticate to get your own API key.  
You will need an API key to access this service.

You do **not** need your own API key to:
- Check the status of our service ()
- Register a new user  (You wil get your own API key here!)

You can register 5 keys per user.

---

### Getting authenticated
Simply **POST** your name and email in JSON format to `/auth/`.
```json
{
  "name": "Alice",
  "email": "alice@mail.com"
}
```
Example:
URL:
POST xxxxx:8080/envdash/v1/auth
Body:
{
  "name": "Alice",
  "email": "alice@mail.com"
}

Your email must contain `@` to be accepted.

You will then receive an API key:

{
  "key": "sk-envdash-..fakeAPIkey..",
  "createdAt": "20260317 20:32"
}

- key: Your personal API key  
- createdAt: When the API key was created  

---

### Using your API key
You must include your API key to use this service.  
This allows the server to authenticate you without requiring login.

When sending requests, add the key to your **header**:

x-api-key
| Key | Value |
|--------|--------|
| x-api-key | The key you got when registering |

---

### Deleting your API key
Simply **DELETE** your API key using:
/envdash/v1/auth/{apikey}

Replace {apikey} with your actual API key.

Example:
DELETE xxxxx:8080/envdash/v1/auth/sk-envdash-fakeAPIkey...

You do not need to include you API key in the request header to delete it.

If successful, you will receive:
204 No Content

If you receive any other status code, the API key was not deleted.

---
## Notification:
If you want a notification when something changes you can register a notification.

Explaining the requierd fealds in a notification:
url: This is where you want your notificatoin to go
country: What country do you want to be notified about
 - If empty, every country is selected
event: Two types of events:
  - Lifecycle events:
  REGISTER, CHANGE, DELETE, INVOKE and THRESHOLD
  You wil then get a notification each time a registration have the event that you are looking for

  - Threshold:
  Threshold have several different fields embedded:
  - Field: what you want a threshold for
        - Allowed: "PM25", "PM10", "TEMPERATURE", "PRECIPITATION"
  - Operator:
        - Allowed: ">", "<", ">=", "<=", "=="
  - Value: What is the threshold

#### Example:
Threshold:
{
   "url":     "https://webhook.site/YOUR_SITE",
   "country": "NO",
   "event":   "THRESHOLD",
   "threshold": {
      "field":    "pm25",
      "operator": "==",
      "value":    35.0
   }
}
Lifecycle:
{
   "url":     "https://webhook.site/YOUR_SITE",
   "country": "NO",
   "event":   "delete"
}


### Making a notification:
You create a notification by sending a POST request to:
/envdash/v1/notifications
The body should be a filled out correct notification body (take a look at the examples above)

If created, you wil get a 201 response

#### Example of response:
{
    "id": "u33FnEzOSGHZ1iHmlql6",
    "country": "NO",
    "event": "DELETE",
    "time": "20260415 12:59"
}

### Looking at notiifcations
If you have forgotten or want to see all registerd notifications, you send:
GET
/envdash/v1/notifications

You wil then receve all notifications registerd to your account.

### Looking at a specific notification
If you know a ID, you can take a look at one secific ID, you send:
GET
/envdash/v1/notifications/{Notificaiton_ID}

This is importaint because when you recive a notification you wil always get the ID for that notification
Looking at a specific notification do not require your user to be the owner of that notification!

### Deleting a notification
If you want to delete your notification simply send:
DELETE
/envdash/v1/notifications/{Notificaiton_ID}

You can not delete someone elses notification
Also costum message if that registration can not be found

## Dependencies
API key storage depends on Firebase. Firebase is used to securely store keys.

If the `/status` endpoint returns anything other than 200 for the Firebase column, API authentication will not work.

Docker
From docker we get golang version etc

#### API dependencies:
| APIs | Endpoint / Documentation | Notes |
|--------|--------|--------|
| REST Countries API | API: http://129.241.150.113:8080/v3.1Docs: http://129.241.150.113:8080| |
| Open-Meteo | Docs: https://open-meteo.com/en/features | |
| OpenAQ v3 | API: https://api.openaq.org/v3Docs: https://docs.openaq.org/ | Need API Key |
| Nominatim (OSM) | |Docs: https://nominatim.org/release-docs/develop/api/Overview/ | |
| Currency API | |API: http://129.241.150.113:9090/currency/Docs: http://129.241.150.113:9090/ | |
| getNamesMakeMap | | | 

 
## Deployment
You will need two credentials:
Open AQ API key: https://docs.openaq.org/
Firestore credentials: 
Make these an enviorment variable in your enviorment


---

## Extra features:
## Logger:
We have deployed a costom logger that loggs all incomming http requests, accross hanlders
## Firestore cache
We have implemented a firestore cache that caches reasently collected data, and stores in firestore
Also inclouded a TTL field in the stored data.

## Authentification:
We have implemented authentification across our service.

### How the server creates API keys (Security)
API keys are generated using a hash of:
- The registered email  
- The exact time the key is created  

This makes duplicate keys extremely unlikely and allows users to create multiple keys safely.
If we detect a duplicate key, we will try a given amout of times (const: MAXATTEMPTSFORKEYGENERATION times).
Also api keys are stored per user (email), each user can only have (const: MAXAPIKEYS api keys).

As an additional security layer, the server uptime is also included in the hash.  
This makes brute-forcing or guessing API keys significantly harder, since an attacker would need:
- The email  
- The exact timestamp (down to fractions of a second)  
- The server uptime  

Furhtermore, notifications have implemented api keys to store a notification in a way that
only lets the user that owns the notification to delete the notification.
