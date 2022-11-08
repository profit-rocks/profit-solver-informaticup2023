#!/usr/bin/env python3
import json
import os
import subprocess
import sys
import time


def output_to_fitness(output):
    for line in output.decode().split("\n"):
        if "final fitness" in line:
            fitness = int(line.split(" ")[-1])
            return fitness


def input_to_benchmark_dicts(input_file):
    output_file = file + ".out"
    start = time.time()
    p = subprocess.Popen(["./profit-solver-icup23", "-input", "tasks/" + file, "-output", output_file], stderr=subprocess.PIPE)
    p.wait()
    end = time.time()

    if p.returncode != 0:
        print("return code", p.returncode)
        print("stderr", p.stderr.read())
        exit(1)
    fitness_dict = {
        "name": f"{input_file} - fitness",
        "unit": "points",
        "value": output_to_fitness(p.stderr.read())
    }
    time_dict = {
        "name": f"{input_file} - time",
        "unit": "seconds",
        "value": end - start
    }
    os.unlink(output_file)
    return fitness_dict, time_dict


if __name__ == '__main__':
    l = []
    for file in os.listdir("tasks"):
        print(file, file=sys.stderr)
        fitness_dict, time_dict = input_to_benchmark_dicts(file)
        l.append(fitness_dict)
        l.append(time_dict)
    print(json.dumps(l))

