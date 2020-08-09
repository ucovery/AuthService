# AuthService
 Test auth service project


ROUTS description:
==================
1# Generates access-refresh tokens with user guid:
http://localhost:8000/signin

POST for Postman
{
	"userguid": "85e321eb-d0d2-4e88-90cf-4236a1675199"
}
==================
2# Updates access-refresh tokens-pair:
http://localhost:8000/refresh
*GET for Postman

==================
3# Delete current token
http://localhost:8000/deleteone
{
"token": "<TOKEN_STRING>"
}

==================
4# Delete all tokens in DB for choosen user
http://localhost:8000/deleteall
POST for Postman
{
	"userguid": "85e321eb-d0d2-4e88-90cf-4236a1675199"
}

