import json
import os

def main():
    upBooks = "updatedBooks.json"
    books = "books.json"
    bf = json.load(open(books, "r"))
    updatedBooks = json.load(open(upBooks, "r"))
    print(len(bf))
    for idx, book in enumerate(updatedBooks):
        if not book:
            updatedBooks.pop(idx)
            
    print(len(updatedBooks))
    json.dump(updatedBooks, open(upBooks, "w"), indent=4)

if __name__ == "__main__":
    main()
