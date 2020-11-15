## Identification Service

### About
An authentication service

---

### Entities

#### Client
A client is any external entity, who wants to consume the identification service api to register a user,
login in a user, etc.
A client must register itself with the service before it can use any of the authentication related apis.

API's available
- /register
- /revoke

#### User
A user represent anyone who will consume clients apis, before they can start consuming they need to registered here
and would need to login.

API's available
- /sign-up
- /update-password

#### Session
A session represent group of interaction a user makes after logging in for a period.

API's available
- /login
- /refresh-token
- /logout

---
 