# Make admin Firebase tool

Used to update the user's admin status (i.e. make user admin/remove admin status) on Firebase.

### Prerequisites

`FIREBASE_CREDENTIALS` variable in `.env` file.

### Usage

- Make user admin:
1. from the terminal, run: `go run cmd/make_admin/main.go`
2. you will be prompted to add the user's email address you wish to make admin

- Remove user admin status:
1. from the terminal, run: `go run cmd/make_admin/main.go make-admin=false`
2. you will be prompted to add the user's email address



