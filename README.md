# evl-book-server
A book server with loan, and upload features.

This repo contains an API server which is built using Golang. It uses Gorilla mux library for HTTP API calls,Cobra CLI for making a CLI of the application and go-jwt for adding jwt authorization token.
It also uses Negroni to provide authentication middlewares. Library admin and users have different middlewares to maintain a role based access. As the expected size of data is small, and in-memory-storage should provide really fast access, Redis is used as the database.

The server uses a configuration file (config.toml) to take user configurable inputs, such as server port, redis port, redis password etc. Data from config file is extracted using Viper.

To keep administration part simple, everytime the server starts, a default account for admin is created (username: admin, password: admin) and saved in the database if one is not already available. You can change the password later by using the `/api/update-profile` endpoint.

### Start the server with default configuration

```
./run.sh
```

this will build the server and start it using the configuration found in `config.toml`.

###Sign Up

Yo have to post valid credentials in json format at endpoint `/api/signup` to create a new account. The valid fields for submittable credentials are:
```
{
	"username": "johndoe",
	"name": "Mr. Doe",
	"password": "supersecretpassword"
}
```
You can use postman or curl to signup. Here's how to sigup using curl:

```shell script
$ curl --header "Content-Type: application/json" \
    --request POST \
    --data '{"username":"johndoe","password":"supersecretpassword", "name": "Mr. Doe"}' \
    http://localhost:3000/api/signup
```
Upon successful registration you will receive `signed up successfully` message.

### To get a JWT token

Yo have to login using valid credentials to get a valid token.

For user:
```shell script
$ curl -u johndoe:supersecretpassword http://localhost:3000/api/login
```

For Admin:
```shell script
$ curl -u admin:admin http://localhost:3000/api/login
```
This should produce a token inside a quotation mark.
```
"eyJhbGciOiJIUzI5cCI6IkpXVCJ9.eyJhZG1......4EnyaQIbxGvwMwh3Pl0"
```
### To change password
You might want to change password or name of any user, specially the admin password. To do so:

```shell script
$ curl --header "Content-Type: application/json" \
    -H "Authorization: Bearer <token>" \
    --request POST \
    --data '{"password":"newsecretpassword", "name": "Mr. Admin"}' \
    http://localhost:3000/api/update-profile
```
You should get a  `profile updated` message.


### User Actions

#####Browse books and authors
Users can browse books and authors. They can also make loan requests.
`curl` to browse:
```shell script
$ curl --header "Content-Type: application/json" \
    -H "Authorization: Bearer <user-token>" \
    --request GET \
    http://localhost:3000/api/books
$ curl --header "Content-Type: application/json" \
    -H "Authorization: Bearer <user-token>" \
    --request GET \
    http://localhost:3000/api/books/<book_id>

$ curl --header "Content-Type: application/json" \
    -H "Authorization: Bearer <user-token>" \
    --request GET \
    http://localhost:3000/api/authors
$ curl --header "Content-Type: application/json" \
    -H "Authorization: Bearer <user-token>" \
    --request GET \
    http://localhost:3000/api/author/<author_id>
```

#####Loan books
Users can also request to loan a book by its book_id, and see all, pending, and accepted loans requests:

```shell script
$ curl --header "Content-Type: application/json" \
    -H "Authorization: Bearer <user-token>" \
    --request GET \
    http://localhost:3000/api/loan/request/<book_id>

$ curl --header "Content-Type: application/json" \
    -H "Authorization: Bearer <user-token>" \
    --request GET \
    http://localhost:3000/api/loan/<loan_id>

$ curl --header "Content-Type: application/json" \
    -H "Authorization: Bearer <user-token>" \
    --request GET \
    http://localhost:3000/api/loans

$ curl --header "Content-Type: application/json" \
    -H "Authorization: Bearer <user-token>" \
    --request GET \
    http://localhost:3000/api/loans/pending

$ curl --header "Content-Type: application/json" \
    -H "Authorization: Bearer <user-token>" \
    --request GET \
    http://localhost:3000/api/loans/active
```

#####Upload a profile picture
Users can upload a profile picture by posting a form that has a input file with the name `profile-picture` using the api endpoint `/api/upload/finalize`.
Since it ties directly into the front-end, we have another endpoint `/api/upload/` to test the upload feature on localhost by calling the actual endpoint `/api/upload/finalize`.
`curl` to test picture upload on localhost:

```shell script
$ curl --header "Content-Type: application/json" \
    -H "Authorization: Bearer <user-token>" \
    -H "filepath: /path/to/a/picture.png" \
    --request POST \
    http://localhost:3000/api/upload
```
Note: The test endpoint `/api/upload/` will only work on localhost. The actual endpoint `/api/upload/finalize` will work anywhere.

### Admin actions
#####CRUD operations on Authors, and Books
Admin can create, update, and delete Authors, and Books.
`curl` to perform CRUD operations on Authors:
```shell script
$ curl --header "Content-Type: application/json" \
    -H "Authorization: Bearer <admin-token>" \
    --request POST \
    --data '{"author_id": 0, "author_name": "An Author"}' \
    http://localhost:3000/api/admin/author/create

$ curl --header "Content-Type: application/json" \
    -H "Authorization: Bearer <admin-token>" \
    --request POST \
    --data '{"author_id": 0, "author_name": "An Author of Quality"}' \
    http://localhost:3000/api/admin/author/update

$ curl --header "Content-Type: application/json" \
    -H "Authorization: Bearer <admin-token>" \
    --request POST \
    http://localhost:3000/api/admin/author/delete/1
```

`curl` to perform CRUD operations on Books:
```shell script
$ curl --header "Content-Type: application/json" \
    -H "Authorization: Bearer <admin-token>" \
    --request POST \
    --data '{"book_id": 1, "author_id": 0, "book_name": "A Book", "count": 10}' \
    http://localhost:3000/api/admin/book/create

$ curl --header "Content-Type: application/json" \
    -H "Authorization: Bearer <admin-token>" \
    --request POST \
    --data '{"book_id": 1, "author_id": 1, "book_name": "A Book"}' \
    http://localhost:3000/api/admin/book/update

$ curl --header "Content-Type: application/json" \
    -H "Authorization: Bearer <admin-token>" \
    --request POST \
    http://localhost:3000/api/admin/book/delete/1
```

Note: use the `count` field during update only if you want to add the given number of books to existing number of books.


##### Browse loans
Admin can also browse all loans, all pending loans, and all active loan requests.
`curl` to browse these loans:

```shell script
$ curl --header "Content-Type: application/json" \
    -H "Authorization: Bearer <admin-token>" \
    --request GET \
    http://localhost:3000/api/admin/loans

$ curl --header "Content-Type: application/json" \
    -H "Authorization: Bearer <admin-token>" \
    --request GET \
    http://localhost:3000/api/admin/loan/<loan_id>

$ curl --header "Content-Type: application/json" \
    -H "Authorization: Bearer <admin-token>" \
    --request GET \
    http://localhost:3000/api/admin/loans/pending

$ curl --header "Content-Type: application/json" \
    -H "Authorization: Bearer <admin-token>" \
    --request GET \
    http://localhost:3000/api/admin/loans/active
```

#####Accept/Reject pending loan requests

Admin can accept/reject pending loan requests made by users by loan_id.
`curl` to perform accept/reject operations on Loans:

```shell script
$ curl --header "Content-Type: application/json" \
    -H "Authorization: Bearer <admin-token>" \
    --request GET \
    http://localhost:3000/api/admin/loans/approve/<loan_id>

$ curl --header "Content-Type: application/json" \
    -H "Authorization: Bearer <admin-token>" \
    --request GET \
    http://localhost:3000/api/admin/loans/reject/<loan_id>
```

Note: Rejecting a loan request removes it from the system entirely

#####Accept a return
Admin can accept a returned book.
`curl` to perform accept a retuned book by loan_id:

```shell script
$ curl --header "Content-Type: application/json" \
    -H "Authorization: Bearer <admin-token>" \
    --request GET \
    http://localhost:3000/api/admin/loans/returned/<loan_id>
```
Note: Accepting a return removes the associated loan from the system entirely