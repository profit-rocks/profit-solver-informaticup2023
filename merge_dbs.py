import sqlite3
hosts = ['bphk1', 'bphk2', 'bphk3', 'bphk4', 'bphk5']
workers = ['run1', 'run2', 'run3']


con = sqlite3.connect("merged.db")
cursor = con.cursor()
cursor.execute("CREATE TABLE IF NOT EXISTS scenarios (id INTEGER PRIMARY KEY, host_id int, worker_id int, width int, height int, coverage int, optimum int, num_obstacles int, num_deposits int, num_products int, num_deposit_types int)")
cursor.execute("CREATE TABLE IF NOT EXISTS obstacles (id INTEGER PRIMARY KEY, type int, x int, y int, width int, height int, scenario_id int, FOREIGN KEY(scenario_id) REFERENCES scenarios(id))")
cursor.execute("CREATE TABLE IF NOT EXISTS products (id INTEGER PRIMARY KEY, points int, scenario_id int, amount_0 int, amount_1 int, amount_2 int, amount_3 int, amount_4 int, amount_5 int, amount_6 int, amount_7 int, FOREIGN KEY (scenario_id) REFERENCES scenarios(id))")
cursor.execute("CREATE TABLE IF NOT EXISTS deposits (id INTEGER PRIMARY KEY, type int, subtype int, x int, y int, width int, height int, scenario_id int, FOREIGN KEY(scenario_id) REFERENCES scenarios(id))")
cursor.execute("CREATE TABLE IF NOT EXISTS runs (id INTEGER PRIMARY KEY, scenario_id int, allowed_turns int, score int, needed_turns int, deadline int, FOREIGN KEY(scenario_id) REFERENCES scenarios(id))")
con.commit()

for host_id, host in enumerate(hosts):
    for worker_id, worker in enumerate(workers):
        old_con = sqlite3.connect(f"{host}_{worker}.db")
        old_cursor = old_con.cursor()
        old_cursor.execute("SELECT * FROM scenarios")
        for scenario in old_cursor.fetchall():
            cursor.execute("INSERT INTO scenarios (host_id, worker_id, width, height, coverage, optimum, num_obstacles, num_deposits, num_products, num_deposit_types) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", (host_id, worker_id, *scenario[1:]))
            scenario_id = cursor.lastrowid
            old_cursor.execute("SELECT * FROM obstacles WHERE scenario_id=?", (scenario[0],))
            for obstacle in old_cursor.fetchall():
                cursor.execute("INSERT INTO obstacles (type, x, y, width, height, scenario_id) VALUES (?, ?, ?, ?, ?, ?)", (*obstacle[1:6], scenario_id))
            old_cursor.execute("SELECT * FROM products WHERE scenario_id=?", (scenario[0],))
            for product in old_cursor.fetchall():
                cursor.execute("INSERT INTO products (points, scenario_id, amount_0, amount_1, amount_2, amount_3, amount_4, amount_5, amount_6, amount_7) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", (product[1], *product[3:], scenario_id))
            old_cursor.execute("SELECT * FROM deposits WHERE scenario_id=?", (scenario[0],))
            for deposit in old_cursor.fetchall():
                cursor.execute("INSERT INTO deposits (type, subtype, x, y, width, height, scenario_id) VALUES (?, ?, ?, ?, ?, ?, ?)", (*deposit[1:7], scenario_id))
            old_cursor.execute("SELECT * FROM runs WHERE scenario_id=?", (scenario[0],))
            for run in old_cursor.fetchall():
                cursor.execute("INSERT INTO runs (scenario_id, allowed_turns, score, needed_turns, deadline) VALUES (?, ?, ?, ?, ?)", (scenario_id, *run[2:]))
            con.commit()