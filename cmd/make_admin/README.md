# Make admin Firebase tool

Used to update the user's admin status (i.e. make user admin/remove admin status) on Firebase.

### Prerequisites

`FIREBASE_CREDENTIALS` variable in `.env` file.

### Usage

Make user admin:

```bash
go run cmd/make_admin/main.go
```

Remove user admin status:

```bash
go run cmd/make_admin/main.go make-admin=false
```

