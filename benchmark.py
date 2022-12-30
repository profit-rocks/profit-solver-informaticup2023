#!/usr/bin/env python3
import json
import os
import subprocess
import sys
import time

NUM_RUNS_PER_FILE = 1


def output_to_fitness(output):
    for line in output.decode().split("\n"):
        if "final fitness" in line:
            fitness = int(line.split(" ")[-1])
            return fitness


def input_to_benchmark_dicts(file):
    output_file = file + ".out"
    start = time.time()
    p = subprocess.Popen(["./profit-solver-icup23", "-input", "tasks/" + file, "-output", output_file], stderr=subprocess.PIPE)
    p.wait()
    end = time.time()

    if p.returncode != 0:
        print("return code", p.returncode)
        print("stderr", p.stderr.read())
        exit(1)
    os.unlink(output_file)
    return output_to_fitness(p.stderr.read()), end-start


if __name__ == '__main__':
    l = []
    for file in os.listdir("tasks"):
        print(file, file=sys.stderr)
        total_time = 0
        total_fitness = 0
        for _ in range(NUM_RUNS_PER_FILE):
            fitness, elapsed_time = input_to_benchmark_dicts(file)
            total_time += elapsed_time
            total_fitness += fitness
        fitness_dict = {
            "name": f"{file} - fitness",
            "unit": "points",
            "value": total_fitness / NUM_RUNS_PER_FILE,
        }
        time_dict = {
            "name": f"{file} - time",
            "unit": "seconds",
            "value": total_time / NUM_RUNS_PER_FILE,
        }
        l.append(fitness_dict)
        l.append(time_dict)
        print("fitness", total_fitness / NUM_RUNS_PER_FILE, "time", total_time / NUM_RUNS_PER_FILE)
    print(json.dumps(l))

