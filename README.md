# I AM ZUUL
This service runs alongside the login section of my dev resume. It handles the callback url for the GitHub SSO, and generates a JWT for control permissions.

## Persistance
Attached to the service is a Postgres DB (currently running in Heroku along with this service). When a user logs in to generate a short lived jwt, it saves the user's information in the User's table. There is also a permissions table that controls what to what resources a user has access. However, at the moment, the only premissions granted are the default.

## Claude
This project is built using AI assistance and the cursor IDE. For every 1 hour of coding, I spend about 4 hours debugging. This may seem bad, but I suspect I would have spent those 4 hours debugging no matter what, and I've just cut the coding from 3 hours to 1.
