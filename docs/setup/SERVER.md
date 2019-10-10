# Cloning The Project

Download the project or clone it via git:

```bash
git clone https://github.com/bjarnemagnussen/go-submarine-swaps.git
```

You should then checkout to the branch corresponding to the article part you are following, e.g. for part one:

```bash
git checkout part-1
```

# Running The Web Server

Open a console and navigate to the directory of the project. Inside the root directory start the web server on the default port `8080` with:

```bash
go run cmd/web/*.go
```

Navigate to `http://localhost:8080` and you should see a page that looks something like:

![][docs/images/main.png]

If you need to change the port the web server is listening on you can provide the `port` option via command-line:

```bash
go run cmd/web/*.go --port 8080
```
