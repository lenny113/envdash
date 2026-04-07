# README


** More coming soon .. **





# README

**More coming soon...**

---

## Authentication

### Why authenticate?
You should authenticate to get your own API key.  
You will need an API key to access this service.

You do **not** need your own API key to:
- Check the status of upstream APIs  
- Register a new user  

You can register as many keys as you want, even with the same email.

---

### Getting authenticated
Simply **POST** your name and email in JSON format to `/auth/`.

Example:
POST xxxxx:8080/auth/
{
  "name": "Alice",
  "email": "alice@mail.com"
}

Your email must contain `@` to be accepted.

You will then receive an API key:

{
  "key": "sk-envdash-fakeAPIkey...",
  "createdAt": "20260317 20:32"
}

- key: Your personal API key  
- createdAt: When the API key was created  

---

### Using your API key
You must include your API key in all requests to this service.  
This allows the server to identify you without requiring login each time.

More info about using API keys is coming soon.

---

### Deleting your API key
Simply **DELETE** your API key using:
/auth/{apikey}

Replace {apikey} with your actual API key.

Example:
DELETE xxxxx:8080/auth/sk-envdash-fakeAPIkey...

You do **not** need to include an API key in the request to delete one.

If successful, you will receive:
204 No Content

If you receive any other status code, the API key was not deleted.

---

### Dependencies
API key storage depends on Firebase. Firebase is used to securely store keys.

If the `/status` endpoint returns anything other than 200 for the Firebase column, API authentication will not work.

---

### How the server creates API keys (Security)
API keys are generated using a hash of:
- The registered email  
- The exact time the key is created  

This makes duplicate keys extremely unlikely and allows users to create multiple keys safely.

As an additional security layer, the server uptime is also included in the hash.  
This makes brute-forcing or guessing API keys significantly harder, since an attacker would need:
- The email  
- The exact timestamp (down to fractions of a second)  
- The server uptime  
