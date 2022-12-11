# Solver for the challenge Profit! of informaticup2023

## Run using docker

Build the docker image using `docker build . -t profit.rocks`. Run with task input by doing `cat task.json | docker run -i profit.rocks > output.json`

## Setup

- Install `lpsolve` package ([see here for instructions](https://pkg.go.dev/github.com/draffensperger/golp#section-readme))
- Use the following environment variables when compiling:

```
CGO_CFLAGS="-I/usr/include/lpsolve"
CGO_LDFLAGS="-llpsolve55 -lm -ldl -lcolamd"
```

## Run the code

- Build a scenario with deposits and obstacles [here](https://profit.phinau.de)
- Export the task as json 
- Run `go run profit-solver-icup23` with input and output as command line arguments
- Import the exported json on phinau to see how the solution looks like

## Profiling

- Use the command line argument to run the code with profiling

- Run `go tool pprof -http localhost:8080 PATH_TO_PTOFILE_FILE` to get a visual representation of the results 
- You might need to install `graphviz`
