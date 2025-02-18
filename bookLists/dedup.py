# Read titles from the full list file
with open("titles.txt", "r", encoding="utf-8") as f:
    all_titles = set(line.strip() for line in f)

# Read titles from the subset file
with open("uT.txt", "r", encoding="utf-8") as f:
    subset_titles = set(line.strip() for line in f)

# Find missing titles
missing_titles = all_titles - subset_titles  # Set difference

# Print or save missing titles
print("Missing titles:")
for title in missing_titles:
    print(title)

# Optionally, save to a new file
with open("missing_titles.txt", "w", encoding="utf-8") as f:
    for title in missing_titles:
        f.write(title + "\n")

