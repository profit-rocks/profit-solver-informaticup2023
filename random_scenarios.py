import random
import sqlite3
from dataclasses import dataclass, field
import subprocess


def output_to_fitness(output):
    for line in output.decode().split("\n"):
        if "final fitness" in line:
            try:
                fitness = int(line.split(" ")[-3])
            except:
                fitness = 0
            return fitness
    return 0


def output_to_needed_turns(output):
    for line in output.decode().split("\n"):
        if "final fitness" in line:
            try:
                turns = int(line.split(" ")[-1])
            except:
                turns = 0
            return turns
    return 0


def output_to_optimum(output):
    for line in output.decode().split("\n"):
        if "theoretical optimum" in line:
            try: 
                optimum = int(line.split(" ")[-1])
            except:
                optimum = 0
            return optimum
    return 0


OBSTACLE_TYPES = [
    {'width': 1, 'height': 20},
    {'width': 2, 'height': 2},
    {'width': 3, 'height': 3},
    {'width': 6, 'height': 6},
    {'width': 7, 'height': 7},
    {'width': 20, 'height': 1},
    {'width': 3, 'height': 6},
    {'width': 6, 'height': 3},
    {'width': 10, 'height': 2},
    {'width': 2, 'height': 10},
]

DEPOSIT_TYPES = [
    {'width': 1, 'height': 1},
    {'width': 3, 'height': 3},
    {'width': 5, 'height': 5},
    {'width': 5, 'height': 10},
    {'width': 10, 'height': 5},
    {'width': 7, 'height': 3},
    {'width': 3, 'height': 7},
    {'width': 10, 'height': 10},
    {'width': 12, 'height': 2},
    {'width': 2, 'height': 12},
]

HIGHEST_REWARD = 7


class RNG:
    numbers: list[int]
    state: int
    size: int

    def __init__(self, size: int):
        self.size = size
        self.state = 0
        self.numbers = random.sample(range(0, size), size)

    def next(self) -> tuple[int, bool]:
        self.state += 1
        return self.numbers[self.state - 1], self.state >= self.size


@dataclass
class Position:
    """
    Represents a position on the grid.
    """
    x: int = 0
    y: int = 0


@dataclass
class Obstacle:
    """
    A obstacle is defined by its width, height and position.
    It doesn't interact with its environment in a meaningful way.
    """
    position: Position
    width: int
    height: int
    type: int

    def toJson(self):
        return f"""{{
            "type": "obstacle",
            "x": {self.position.x},
            "y": {self.position.y},
            "width": {self.width},
            "height": {self.height}
        }}"""


@dataclass
class Deposit:
    """A deposit has a width and a height. Those two parameters define the capacity of a deposit.
    The subtype defines which ressource it provides.
    All connected mines can get ressources from the deposit. The used variable counts the number of mined ressources.
    A Deposit is depleted, when the number of mined ressources is equal to the capacity.
    """
    position: Position
    width: int
    height: int
    subtype: int
    type: int

    def toJson(self):
        return f"""{{
            "type": "deposit",
            "subtype": {self.subtype},
            "x": {self.position.x},
            "y": {self.position.y},
            "width": {self.width},
            "height": {self.height}
        }}"""


class Product:
    """
    A product is defined by its type and the amount of ressources it needs.
    """
    subtype: int
    needed_ressources: list[int]
    reward: int

    def __init__(self, subtype: int,
                 needed_ressources: list[int], reward: int) -> None:
        self.subtype = subtype
        self.needed_ressources = needed_ressources
        self.reward = reward

    def toJson(self):
        return f"""{{
            "type": "product",
            "subtype": {self.subtype},
            "resources": {self.needed_ressources},
            "points": {self.reward}
        }}"""


class Scenario:
    width: int
    height: int
    obstacles: list[Obstacle]
    deposits: list[Deposit]
    ressources_available: dict[int, bool]
    products: list[Product]
    used_space: int
    covered_percent: float

    def __init__(self) -> None:
        self.used_space = 0
        self.width = random.randrange(10, 101, 10)
        self.height = random.randrange(10, 101, 10)
        self.num_deposit_types = random.randint(1, 7)
        self.ressources_available = {}
        self.covered_percent = random.randrange(10, 36)
        self.deposits = []
        self.obstacles = []
        for i in range(8):
            self.ressources_available[i] = False
        self.num_deposits = random.randint(1, 10)
        self.num_products = random.randint(1, 7)
        self.populate()
        self.select_products()

    def toJson(self) -> str:
        result = f"""{{
            "width": {self.width},
            "height": {self.height},
            "objects": ["""
        for i, deposit in enumerate(self.deposits):
            if i != 0:
                result += ","
            result += deposit.toJson()

        for obstacle in self.obstacles:
            result += ","
            result += obstacle.toJson()

        result += '], "products": ['
        for i, product in enumerate(self.products):
            if i != 0:
                result += ","
            result += product.toJson()
        result += "],"
        return result

    def populate(self):
        """
        Populates the grid with obstacles and deposits.
        """
        while self.used_space < self.width * self.height * (self.covered_percent / 100):
            if random.randint(0, 4) == 0:
                self.build_obstacle()
            else:
                self.build_deposit()

    def on_grid(self, pos: Position) -> bool:
        """
        Checks if this position is on the current grid. Returns true if is on grid, false otherwise.
        """
        return -1 < pos.x < self.width and -1 < pos.y < self.height

    def select_products(self):
        """
        Selects a random number of products from the available ressources.
        """
        self.products = []
        for k in range(self.num_products):
            needed_ressources = [
                random.randint(
                    0, 10) if self.ressources_available[z] else 0 for z in range(8)]
            valid = False
            for i in needed_ressources:
                if i != 0:
                    valid = True
                    break
            if not valid:
                continue
            product = Product(
                k, needed_ressources, random.randint(
                    0, HIGHEST_REWARD))
            self.products.append(product)

    def build_obstacle(self):
        """
        Builds a list of obstacles with random sizes and positions.
        """
        typeRNG = RNG(len(OBSTACLE_TYPES))
        allTypesChecked = False
        while not allTypesChecked:
            obstacle_type_index, allTypesChecked = typeRNG.next()
            obstacle_type = OBSTACLE_TYPES[obstacle_type_index]
            obstacle = Obstacle(
                Position(
                    0,
                    0),
                obstacle_type['width'],
                obstacle_type['height'],
                obstacle_type_index)
            positionRNG = RNG(self.width * self.height)
            finished = False
            built = False
            while not finished and not built:
                position, finished = positionRNG.next()
                obstacle.position = Position(
                    position %
                    self.width,
                    position //
                    self.width)
                if self.position_available_for_obstacle(obstacle):
                    built = True
            if built:
                obstacle.type = obstacle_type_index
                self.obstacles.append(obstacle)
                self.used_space += obstacle.width * obstacle.height
                break

    def build_deposit(self):
        typeRNG = RNG(len(DEPOSIT_TYPES))
        allTypesChecked = False
        while not allTypesChecked:
            deposit_type_index, allTypesChecked = typeRNG.next()
            deposit_type = DEPOSIT_TYPES[deposit_type_index]
            deposit = Deposit(
                Position(
                    0,
                    0),
                deposit_type['width'],
                deposit_type['height'],
                random.randint(
                    0,
                    self.num_deposit_types),
                deposit_type_index)
            positionRNG = RNG(self.width * self.height)
            finished = False
            built = False
            while not finished and not built:
                position, finished = positionRNG.next()
                deposit.position = Position(
                    position %
                    self.width,
                    position //
                    self.width)
                if self.position_available_for_deposit(deposit):
                    built = True
            if built:
                self.ressources_available[deposit.subtype] = True
                deposit.type = deposit_type_index
                self.deposits.append(deposit)
                self.used_space += deposit.width * deposit.height
                break

    def position_available_for_obstacle(self, obstacle: Obstacle) -> bool:
        """
        Checks if the given obstacle can be placed on the grid. Returns true if it can be placed, false otherwise.
        """
        if not self.on_grid(obstacle.position) or not self.on_grid(Position(
                obstacle.position.x + obstacle.width, obstacle.position.y + obstacle.height)):
            return False
        for other_obstacle in self.obstacles:
            if obstacle.position.x < other_obstacle.position.x + other_obstacle.width and obstacle.position.x + \
                    obstacle.width > other_obstacle.position.x and obstacle.position.y < other_obstacle.position.y + other_obstacle.height and obstacle.position.y + obstacle.height > other_obstacle.position.y:
                return False
        for deposit in self.deposits:
            if obstacle.position.x < deposit.position.x + deposit.width and obstacle.position.x + \
                    obstacle.width > deposit.position.x and obstacle.position.y < deposit.position.y + deposit.height and obstacle.position.y + obstacle.height > deposit.position.y:
                return False
        return True

    def position_available_for_deposit(self, deposit: Deposit) -> bool:
        """
        Checks if the given deposit can be placed on the grid. Returns true if it can be placed, false otherwise.
        """
        if not self.on_grid(deposit.position) or not self.on_grid(Position(
                deposit.position.x + deposit.width, deposit.position.y + deposit.height)):
            return False
        for other_obstacle in self.obstacles:
            if deposit.position.x < other_obstacle.position.x + other_obstacle.width and deposit.position.x + \
                    deposit.width > other_obstacle.position.x and deposit.position.y < other_obstacle.position.y + other_obstacle.height and deposit.position.y + deposit.height > other_obstacle.position.y:
                return False
        for other_deposit in self.deposits:
            if deposit.position.x < other_deposit.position.x + other_deposit.width and deposit.position.x + \
                    deposit.width > other_deposit.position.x and deposit.position.y < other_deposit.position.y + other_deposit.height and deposit.position.y + deposit.height > other_deposit.position.y:
                return False
        return True


if __name__ == "__main__":
    # create database
    con = sqlite3.connect("database.db")
    cursor = con.cursor()
    cursor.execute("CREATE TABLE IF NOT EXISTS scenarios (id INTEGER PRIMARY KEY, width int, height int, coverage int, optimum int, num_obstacles int, num_deposits int, num_products int, num_deposit_types int)")
    cursor.execute("CREATE TABLE IF NOT EXISTS obstacles (id INTEGER PRIMARY KEY, type int, x int, y int, width int, height int, scenario_id int, FOREIGN KEY(scenario_id) REFERENCES scenarios(id))")
    cursor.execute("CREATE TABLE IF NOT EXISTS products (id INTEGER PRIMARY KEY, points int, scenario_id int, amount_0 int, amount_1 int, amount_2 int, amount_3 int, amount_4 int, amount_5 int, amount_6 int, amount_7 int, FOREIGN KEY (scenario_id) REFERENCES scenarios(id))")
    cursor.execute("CREATE TABLE IF NOT EXISTS deposits (id INTEGER PRIMARY KEY, type int, subtype int, x int, y int, width int, height int, scenario_id int, FOREIGN KEY(scenario_id) REFERENCES scenarios(id))")
    cursor.execute("CREATE TABLE IF NOT EXISTS runs (id INTEGER PRIMARY KEY, scenario_id int, allowed_turns int, score int, needed_turns int, deadline int, FOREIGN KEY(scenario_id) REFERENCES scenarios(id))")
    con.commit()
    while True:
        scenario = Scenario()
        if len(scenario.deposits) == 0:
            continue

        testFactory = Obstacle(Position(0, 0), 5, 5, 0)

        positionRNG = RNG(scenario.width * scenario.height)

        finished = False
        built = False
        while not finished and not built:
            position, finished = positionRNG.next()
            testFactory.position = Position(
                position %
                scenario.width,
                position //
                scenario.width)
            if scenario.position_available_for_obstacle(testFactory):
                built = True
                break
        if not built:
            continue

        json_without_turns = scenario.toJson()
        # Update db
        cursor.execute("INSERT INTO scenarios (width, height, coverage, num_obstacles, num_deposits, num_products, num_deposit_types) VALUES (?, ?, ?, ?, ?, ?, ?)",
                       (scenario.width, scenario.height, scenario.covered_percent, len(scenario.obstacles), len(scenario.obstacles), len(scenario.products), scenario.num_deposit_types))
        scenario_id = cursor.lastrowid
        for product in scenario.products:
            cursor.execute(
                "INSERT INTO products (points, amount_0, amount_1, amount_2, amount_3, amount_4, amount_5, amount_6, amount_7, scenario_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
                (product.reward,
                 product.needed_ressources[0],
                 product.needed_ressources[1],
                 product.needed_ressources[2],
                 product.needed_ressources[3],
                 product.needed_ressources[4],
                 product.needed_ressources[5],
                 product.needed_ressources[6],
                 product.needed_ressources[7],
                 scenario_id))
        for obstacle in scenario.obstacles:
            cursor.execute(
                "INSERT INTO obstacles (type, x, y, width, height, scenario_id) VALUES (?, ?, ?, ?, ?, ?)",
                (obstacle.type,
                 obstacle.position.x,
                 obstacle.position.y,
                 obstacle.width,
                 obstacle.height,
                 scenario_id))
        for deposit in scenario.deposits:
            cursor.execute(
                "INSERT INTO deposits (type, subtype, x, y, width, height, scenario_id) VALUES (?, ?, ?, ?, ?, ?, ?)",
                (deposit.type,
                 deposit.subtype,
                 deposit.position.x,
                 deposit.position.y,
                 deposit.width,
                 deposit.height,
                 scenario_id))

        con.commit()
        # Run the scenario
        for allowed_turns in range(
                10, 100, 20):
            for deadline in range(20, 100, 20):
                json = json_without_turns + \
                    f'"turns": {allowed_turns}, "time": {deadline}}}'
                output_file = f"scenario_{scenario_id}_allowed_turns_{allowed_turns}_deadline_{deadline}_solution.json"
                p = subprocess.Popen(["./profit-solver-icup23",
                                      "-iters",
                                      "0",
                                      "-output",
                                      output_file],
                                     stdin=subprocess.PIPE,
                                     stderr=subprocess.PIPE)
                with open(f"scenario_{scenario_id}_allowed_turns_{allowed_turns}_deadline_{deadline}.json", "w") as file:
                    file.write(json)
                try:
                    outs, err = p.communicate(
                        input=json.encode(), timeout=deadline + 5)
                except subprocess.TimeoutExpired:
                    p.kill()
                    outs, err = p.communicate()
                    print("Timeout")
                print(err)
                needed_turns = output_to_needed_turns(err)
                score = output_to_fitness(err)
                optimum = output_to_optimum(err)
                print(
                    f"Scenario {scenario_id} with {allowed_turns} allowed turns and {deadline} deadline needed {needed_turns} turns and scored {score}")
                cursor.execute(
                    "INSERT INTO runs (scenario_id, allowed_turns, score, needed_turns, deadline) VALUES (?, ?, ?, ?, ?)",
                    (scenario_id,
                     allowed_turns,
                     score,
                     needed_turns,
                     deadline))
                cursor.execute(
                    "UPDATE scenarios SET optimum = ? WHERE id = ?", (optimum, scenario_id))
                con.commit()
