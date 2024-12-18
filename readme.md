Budgie is a program for managing and reviewing monthly expenses.
Data is stored in a local MongoDB server.

Data is imported into the database via csv files.
Managed by a TUI, but reports are viewed on a web server (TODO).

Environment:
- Ubuntu 22.04 (or WSL)

Installation
1. Install MongoDB locally (https://www.mongodb.com/docs/manual/tutorial/install-mongodb-on-ubuntu/)
2. Install Go (https://go.dev/doc/install)
3. Clone this repo
4. `cd` into the repo
3. Enter `go run .` to launch the TUI

Managing mongodb from mongosh:

```
show dbs
use budgie
db.expenses.find

Deleting from db: `db.expenses.deleteMany({})`