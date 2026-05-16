with open('./extracted') as four:
    lines = four.readlines()

    found = set()

    for i in range(len(lines)):
        for j in range(len(lines)):
            if lines[j] in found:
                continue

            if lines[i] == lines[j] and i != j:
                found.add(lines[i])
                print(f"Duplicate {lines[i]}", end="")
