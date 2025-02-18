import json
import re

def standardize_date(date_str):
    """Converts date to YYYY-MM-DD format, defaulting missing values to '01'."""
    if re.fullmatch(r"\d{4}-\d{2}-\d{2}", date_str):  # Already correct
        return date_str
    elif re.fullmatch(r"\d{4}-\d{2}", date_str):  # YYYY-MM -> YYYY-MM-01
        return date_str + "-01"
    elif re.fullmatch(r"\d{4}", date_str):  # YYYY -> YYYY-01-01
        return date_str + "-01-01"
    else:
        return "1970-01-01"  # Return start of epoch


# Load JSON
with open("populatedBooks.json", "r", encoding="utf-8") as file:
    data = json.load(file)

# Process JSON
for entry in data:
    if "published_date" in entry and entry["published_date"]:
        standardized = standardize_date(entry["published_date"])
        if standardized:
            entry["published_date"] = standardized

# Save JSON
with open("data_standardized.json", "w", encoding="utf-8") as file:
    json.dump(data, file, indent=4)

print("Date standardization complete! Output saved as data_standardized.json")

