![logo_profit_rocks_crop](https://user-images.githubusercontent.com/46268468/216763253-023acb21-618f-4f29-95e1-b128b185327f.png)

# Genetic Algorithm for _Profit!_ for InformatiCup 2023

InformatiCup is a competition for computer science students in Germany, Austria and Switzerland. The competition is held by the [German Society of Computer Science](https://gi.de/). The 2023 edition of InformatiCup is called _Profit!_. More information about the competition is available in the [official repository](https://github.com/informatiCup/informatiCup2023) and the [official challenge description](https://github.com/informatiCup/informatiCup2023/blob/main/informatiCup%202023%20-%20Profit!.pdf)

This repository contains our solution to _Profit!_ which is based on an approach using genetic algorithms. See [report.pdf](report.pdf) for a detailed description of the theoretical approach, further thoughts on our development process, and an evaluation of our solution.

An interactive playground for the challenge is available at [https://profit.phinau.de](https://profit.phinau.de). The website offers scenario visualization, simulation, import and export. By clicking _Export (task)_, you download a scenario file which can be imported by our solution.

## Example task
We chose `tasks/jacob.json` as our example task. The task is a medium-sized scenario, with the time set to 60 seconds and turns set to 75. The following products are enabled:
| Resource  | 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 |
|-----------|---|---|---|---|---|---|---|---|
| Product 1 | 1 | 1 | 1 | 1 | 0 | 0 | 0 | 0 |
| Product 4 | 0 | 0 | 4 | 2 | 0 | 0 | 0 | 0 |

Our algorithm produces the following solution, which scores 1830 points after 41 turns. The theoretical optimum is 1850 points:

![image](https://user-images.githubusercontent.com/46268468/212553297-891d44c0-a0c8-422f-ba87-6cd1d965eec9.png)

## Run using docker

If you want to try out our solution, you can do so using Docker.

1. If you do not have a Docker installation, consult the [official documentation](https://docs.docker.com/)
1. Build the docker image using `docker build . -t profit.rocks`.
1. Run with task input by running `cat task.json | docker run -i --rm --network none --cpus 2.000 --memory 2G --memory-swap 2G profit.rocks > output.json`. These are the official parameters from the challenge description.

## Development Setup

You can also run our solution without Docker using the instructions below.

### Building

1. If you do not have Go installed, consult the [official documentation](https://go.dev/doc/install). We tested our solution with Go version 1.19.3.
2. To calculate optimal scores, we are using a Go package called `lpsolve`. Documentation on how to install the package and its dependencies, can be found on their [official website](https://pkg.go.dev/github.com/draffensperger/golp#section-readme))
3. For the build, you need to set the following environment variables:

```
CGO_CFLAGS="-I/usr/include/lpsolve"
CGO_LDFLAGS="-llpsolve55 -lm -ldl -lcolamd"
```

4. Build our software using `go build`. The resulting binary is called `profit-solver-icup23`. After changing the code, you have to rebuild the binary for changes to take effect. Please note that we only tested the building process on Linux. Please look at the package documentation if you are building our software on Windows.

To set the environment variables and build the program in one command, you can run:

```bash
CGO_CFLAGS="-I/usr/include/lpsolve" CGO_LDFLAGS="-llpsolve55 -lm -ldl -lcolamd" go build
```

### Running

Execute `./profit-solver-icup23 -help` to get an overview over the command line options:

```bash
$ ./profit-solver-icup23 -help
Usage of ./profit-solver-icup23:
  -cpuprofile string
    	Path to output cpu profile
  -endonoptimal
    	End when optimal solution is found
  -exporter string
    	Export type, either "scenario" or "solution" (default "scenario")
  -input string
    	Path to input scenario json (default "-")
  -iters int
    	Number of iterations to run. Use 0 for unlimited (default 50)
  -logdir string
    	Directory to log top chromosomes in each iteration
  -output string
    	Path to output scenario json (default "-")
  -seed int
    	Seed for random number generator
  -visualizedir string
    	Directory to visualize chromosomes in each iteration
```

In its default configuration, our program reads the scenario from `stdin` and writes it output to `stdout`. Use `-input` and `-output` to read from a file and write to a file.

### Profiling

To investigate performance issues, you can use the integrated CPU profiler.

- For the graphical output, you might need to install `graphviz` on your system. Look [here](https://graphviz.org/download/) for detailed installation instructions.
- Use the command line argument `-cpuprofile PATH_TO_PROFILE_FILE` to run the code with profiling
- Run `go tool pprof -http localhost:8080 PATH_TO_PROFILE_FILE` to get a visual representation of the results

### Benchmarking

To run benchmarks on our solution, you can use the `benchmark.py` script which requires Python version 3.10 or higher. By default, it runs benchmarks on all tasks in `./tasks`. Add the `--keep-solutions` flag if you want to access the intermediate and final solutions afterward.
