# Solver for the challenge Profit! of informaticup2023
The informaticup is a competition for computer science students in Germany, Austria and Switzerland. The competition is held by the German Society of Informatics. More information about the competition is available in the [official repository](https://github.com/informatiCup/informatiCup2023) and the [official challenge description](https://github.com/informatiCup/informatiCup2023/blob/main/informatiCup%202023%20-%20Profit!.pdf)

## Run using docker

Docker is the easy way of running our solution. If you do not want to develop our solution further, it is the preffered method of running our solution. 
Build the docker image using `docker build . -t profit.rocks`. Run with task input by doing `cat task.json | docker run -i --rm --network none --cpus 2.000 --memory 2G --memory-swap 2G profit.rocks > output.json`. These are the official parameters mentioned in the challenge description.

## Installation (Development setup)
After completing these steps, you can run our solution locally and start with the development.

### 0. Prerequisites
Our solution is written in Golang. To run it, you need a Golang installation on your computer. Instructions on how to get a running Golang installation can be found [here](https://go.dev/doc/install). For our solution you need at least Golang version 1.19.3.

To calculate optimal scores, we are using a package called `lpsolve`. Documentation on how to install the package and its dependencies, can be found on their official website ([see here for instructions](https://pkg.go.dev/github.com/draffensperger/golp#section-readme))

As mentioned in the instructions, it is important to set the right environment variables, when building our program. Please note this is the configuration for Linux and we haven't tested it on Windows. Please take a look in the instructions linked above, if you're using windows.
```
CGO_CFLAGS="-I/usr/include/lpsolve"
CGO_LDFLAGS="-llpsolve55 -lm -ldl -lcolamd"
```

If you already have a valid task you want to solve, go to step 1. Else continue reading. 
Tasks can be built on the website [https://profit.phinau.de](https://profit.phinau.de). The website offers a visualization of the json format scenarios are in. Create a valid scenario there and export it into a json file. Afterwards you can use our solution, to solve the task. If you want a visual representation of the computed solution, you can import it once again there.

### 1. Executing our programm

Go is a compiled language. In order to start the compile run `go build` in the root of the repository. There is now a binary called `profit-solver-icup23`. Execute `./profit-solver-icup23 --help` to get an overview over the command line options. In its default configuration our program expects the scenario via `stdin` and writes it output to `stdout`. This is quite incovenient for development purposes. Use `-input` and `-output` to read from a file and write to a file.

Instead of compiling and the executing the binary in seperate steps, use `go run profit-solver-icup23` to build and run the program at the same time. Use the command line arguments in the same way. After changing the code, you have to rebuild the binary for changes to take effect.

If the execution fails, make sure you set the environment variables as mentioned above.

## Profiling

- Use the command line argument `-cpuprofile PATH_TO_PROFILE_FILE` to run the code with profiling
- Run `go tool pprof -http localhost:8080 PATH_TO_PROFILE_FILE` to get a visual representation of the results 
- You might need to install `graphviz` on your system. Look [here](https://graphviz.org/download/) for detailed installation instructions.

## Benchmarking

- You can benchmark all tasks in `./tasks` using `python benchmark.py`.
- Add `--keep-solutions` if you want to access the intermediate and final solutions afterwards
- If you want to add a task for benchmarking, simply put it in `./tasks`
